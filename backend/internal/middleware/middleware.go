package middleware

import (
	"context"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"syntheticvision/internal/auth"
	"syntheticvision/internal/models"
)

type ctxKey string

// Context keys for the authenticated user id and role.
const (
	UserIDKey ctxKey = "uid"
	RoleKey   ctxKey = "role"
)

// RequireAuth returns middleware that validates the Bearer token using secret.
// It responds 401 when the token is missing or invalid; otherwise it stores
// the user id and role in the request context.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := bearerToken(r)
			if tokenStr == "" {
				unauthorized(w)
				return
			}
			claims, err := auth.ParseToken(secret, tokenStr)
			if err != nil {
				unauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdminDB returns middleware that re-loads the authenticated user from
// the database and requires a *persisted* admin role. Unlike trusting the JWT
// claim alone, this revokes access immediately when a user is demoted or
// deleted, instead of waiting up to the full token TTL for it to expire.
func RequireAdminDB(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := UserID(r)
			var user models.User
			if uid == "" ||
				db.Select("role").Where("id = ?", uid).First(&user).Error != nil ||
				user.Role != "admin" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserID returns the authenticated user id from the request context.
func UserID(r *http.Request) string {
	if v, ok := r.Context().Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

// Role returns the authenticated user role from the request context.
func Role(r *http.Request) string {
	if v, ok := r.Context().Value(RoleKey).(string); ok {
		return v
	}
	return ""
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
