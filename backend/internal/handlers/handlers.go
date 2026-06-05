// Package handlers implements the HTTP layer for Synthetic Vision: auth,
// generation, profile, and admin endpoints. Handlers depend on the service
// layer for business logic and GORM for direct reads.
package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"syntheticvision/internal/config"
	"syntheticvision/internal/service"
)

// Handler bundles the shared dependencies every HTTP handler needs.
type Handler struct {
	DB  *gorm.DB
	Svc *service.Service
	Cfg config.Config
}

// New constructs a Handler with its dependencies.
func New(db *gorm.DB, svc *service.Service, cfg config.Config) *Handler {
	return &Handler{DB: db, Svc: svc, Cfg: cfg}
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// writeError emits a JSON error body of the shape {"error": msg}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---- Request DTOs (request bodies decoded from JSON) ----

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createGenRequest struct {
	Mode           string `json:"mode"`
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt"`
	Style          string `json:"style"`
	Resolution     string `json:"resolution"`
	AspectRatio    string `json:"aspect_ratio"`
}

type injectCreditsRequest struct {
	TargetPublicID string `json:"target_public_id"`
	Amount         int    `json:"amount"`
}

// decodeJSON decodes the request body into dst, returning false (and writing a
// 400 response) when the payload is malformed.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "missing request body")
		return false
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
