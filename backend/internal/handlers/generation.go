package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"syntheticvision/internal/middleware"
	"syntheticvision/internal/models"
	"syntheticvision/internal/service"
)

// CreateGeneration validates the request, deducts credits, and enqueues an
// async render. Returns the freshly created (pending) generation.
//
//	POST /api/generations -> 202 generation
//	402 {"error":"insufficient credits"} when the user is broke.
func (h *Handler) CreateGeneration(w http.ResponseWriter, r *http.Request) {
	var req createGenRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusUnprocessableEntity, "prompt is required")
		return
	}

	uid := middleware.UserID(r)
	gen, err := h.Svc.CreateGeneration(uid, service.CreateGenInput{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Style:          req.Style,
		Resolution:     req.Resolution,
		AspectRatio:    req.AspectRatio,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientCredits):
			writeError(w, http.StatusPaymentRequired, "insufficient credits")
		case errors.Is(err, service.ErrInvalidParam):
			writeError(w, http.StatusUnprocessableEntity, "invalid parameter")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create generation")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, gen)
}

// ListGenerations returns the caller's generations, newest first, optionally
// filtered by status and capped by limit.
//
//	GET /api/generations?status=&limit= -> {generations:[...]}
func (h *Handler) ListGenerations(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)

	q := h.DB.Where("user_id = ?", uid).Order("created_at DESC")

	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		q = q.Where("status = ?", status)
	}
	if limitStr := strings.TrimSpace(r.URL.Query().Get("limit")); limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			q = q.Limit(n)
		}
	}

	gens := []models.Generation{}
	if err := q.Find(&gens).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list generations")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"generations": gens})
}

// GetGeneration returns a single generation owned by the caller.
//
//	GET /api/generations/{id} -> generation (404/403)
func (h *Handler) GetGeneration(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)
	id := chi.URLParam(r, "id")

	var gen models.Generation
	if err := h.DB.Where("id = ?", id).First(&gen).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "generation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if gen.UserID != uid {
		writeError(w, http.StatusForbidden, "not your generation")
		return
	}

	writeJSON(w, http.StatusOK, gen)
}

// DeleteGeneration removes a generation owned by the caller and deletes its
// rendered image file from disk.
//
//	DELETE /api/generations/{id} -> {ok:true}
func (h *Handler) DeleteGeneration(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)
	id := chi.URLParam(r, "id")

	var gen models.Generation
	if err := h.DB.Where("id = ?", id).First(&gen).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "generation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if gen.UserID != uid {
		writeError(w, http.StatusForbidden, "not your generation")
		return
	}

	if err := h.DB.Delete(&models.Generation{}, "id = ?", id).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete generation")
		return
	}

	// Best-effort removal of the rendered PNG. The canonical on-disk name is
	// "<id>.png" under ImagesDir; never trust the stored URL for a filesystem
	// path. Missing files are not an error.
	imgPath := filepath.Join(h.Cfg.ImagesDir(), gen.ID+".png")
	if err := os.Remove(imgPath); err != nil && !os.IsNotExist(err) {
		// Log-worthy but not fatal; the row is already gone.
		_ = err
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GenerationCost reports the credit cost for a given resolution.
//
//	GET /api/generations/cost?resolution=2K -> {cost: 15}
func (h *Handler) GenerationCost(w http.ResponseWriter, r *http.Request) {
	resolution := strings.TrimSpace(r.URL.Query().Get("resolution"))
	cost := service.CostFor(resolution)
	if cost < 0 {
		writeError(w, http.StatusUnprocessableEntity, "invalid resolution")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cost": cost})
}
