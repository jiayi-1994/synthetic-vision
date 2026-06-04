// Package server wires the chi router: API routes, the /images file server,
// and the embedded single-page-application with history fallback.
package server

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"syntheticvision/internal/config"
	"syntheticvision/internal/handlers"
	"syntheticvision/internal/middleware"
	"syntheticvision/web"
)

// Router builds the full HTTP handler tree for the application.
func Router(h *handlers.Handler, cfg config.Config) http.Handler {
	r := chi.NewRouter()

	// Base middleware chain.
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// ---- API ----
	r.Route("/api", func(api chi.Router) {
		// Public auth endpoints.
		api.Post("/auth/register", h.Register)
		api.Post("/auth/login", h.Login)

		// Authenticated endpoints.
		api.Group(func(auth chi.Router) {
			auth.Use(middleware.RequireAuth(cfg.JWTSecret))

			auth.Get("/auth/me", h.Me)
			auth.Get("/me/stats", h.Stats)
			auth.Get("/me/analytics", h.Analytics)

			// Generations. NOTE: the literal "/cost" sub-route is declared
			// before "/{id}" so it is matched ahead of the wildcard.
			auth.Post("/generations", h.CreateGeneration)
			auth.Get("/generations", h.ListGenerations)
			auth.Get("/generations/cost", h.GenerationCost)
			auth.Get("/generations/{id}", h.GetGeneration)
			auth.Delete("/generations/{id}", h.DeleteGeneration)

			// Admin-only endpoints. RequireAdminDB re-checks the persisted role
			// on every call so demotions/deletions take effect immediately.
			auth.Group(func(admin chi.Router) {
				admin.Use(middleware.RequireAdminDB(h.DB))

				admin.Get("/admin/users", h.AdminUsers)
				admin.Post("/admin/credits", h.AdminInjectCredits)
				admin.Get("/admin/cluster", h.AdminCluster)
			})
		})
	})

	// Unknown /api paths get a JSON 404 rather than the SPA shell.
	r.NotFound(notFoundHandler(cfg))

	// ---- Static: generated images ----
	// Deliberate design: images are served as unauthenticated, unguessable-URL
	// public hosting. Filenames are server-generated UUIDs returned only to the
	// owner in generation JSON, so possession of the URL is the access control.
	// (chi path-cleaning + StripPrefix + http.Dir block ../ traversal.)
	imagesFS := http.FileServer(http.Dir(cfg.ImagesDir()))
	r.Handle("/images/*", http.StripPrefix("/images/", imagesFS))

	// ---- SPA: embedded frontend with history fallback ----
	dist, err := web.DistFS()
	if err != nil {
		// If the embed failed we still serve the API; static returns 500.
		r.Get("/*", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend assets unavailable", http.StatusInternalServerError)
		})
		return r
	}
	r.Get("/*", spaHandler(dist))

	return r
}

// notFoundHandler returns JSON 404 for /api paths and serves the SPA shell for
// everything else (so deep links resolve client-side).
func notFoundHandler(cfg config.Config) http.HandlerFunc {
	dist, err := web.DistFS()
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/images/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
			return
		}
		if err != nil {
			http.Error(w, "frontend assets unavailable", http.StatusInternalServerError)
			return
		}
		serveIndex(w, r, dist)
	}
}

// spaHandler serves static assets from the embedded dist FS when they exist,
// and falls back to index.html (history-mode routing) otherwise.
func spaHandler(dist fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(dist))
	return func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(r.URL.Path, "/")
		if clean == "" {
			serveIndex(w, r, dist)
			return
		}
		// If the requested asset exists in the embedded FS, serve it directly.
		if f, statErr := fs.Stat(dist, clean); statErr == nil && !f.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		// Otherwise fall back to the SPA shell for client-side routing.
		serveIndex(w, r, dist)
	}
}

// serveIndex writes dist/index.html with a no-cache header so the SPA shell is
// always fresh after a redeploy.
func serveIndex(w http.ResponseWriter, r *http.Request, dist fs.FS) {
	data, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "index.html", time.Time{}, strings.NewReader(string(data)))
}
