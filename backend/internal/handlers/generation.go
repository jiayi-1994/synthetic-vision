package handlers

import (
	"errors"
	"fmt"
	"io"
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

const maxGenerationUploadBytes = 10 << 20

// CreateGeneration validates the request, deducts credits, and enqueues async
// renders. Returns the freshly created generation(s) with 202 status.
//
//	POST /api/generations -> 202 {generations:[...]}
//	402 {"error":"insufficient credits"} when the user is broke.
func (h *Handler) CreateGeneration(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
		h.createMultipartGeneration(w, r)
		return
	}

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
		Mode:           req.Mode,
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Style:          req.Style,
		Resolution:     req.Resolution,
		AspectRatio:    req.AspectRatio,
		Count:          req.Count,
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

	writeJSON(w, http.StatusAccepted, map[string]any{"generations": gen})
}

func (h *Handler) createMultipartGeneration(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, (2*maxGenerationUploadBytes)+(4<<20))
	if err := r.ParseMultipartForm(2 * maxGenerationUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart body")
		return
	}

	prompt := strings.TrimSpace(r.FormValue("prompt"))
	if prompt == "" {
		writeError(w, http.StatusUnprocessableEntity, "prompt is required")
		return
	}
	count, err := parseGenerationCount(r.FormValue("count"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid count")
		return
	}

	sourceImage, err := readGenerationUpload(r, "source_image", map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	maskImage, err := readGenerationUpload(r, "mask_image", map[string]bool{
		"image/png": true,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	uid := middleware.UserID(r)
	gen, err := h.Svc.CreateGeneration(uid, service.CreateGenInput{
		Mode:           r.FormValue("mode"),
		Prompt:         prompt,
		NegativePrompt: r.FormValue("negative_prompt"),
		Style:          r.FormValue("style"),
		Resolution:     r.FormValue("resolution"),
		AspectRatio:    r.FormValue("aspect_ratio"),
		Count:          count,
		SourceImage:    sourceImage,
		MaskImage:      maskImage,
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

	writeJSON(w, http.StatusAccepted, map[string]any{"generations": gen})
}

func parseGenerationCount(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func readGenerationUpload(r *http.Request, field string, allowed map[string]bool) (*service.UploadedImage, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, nil
		}
		return nil, fmt.Errorf("invalid %s upload", field)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxGenerationUploadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s upload", field)
	}
	if len(data) > maxGenerationUploadBytes {
		return nil, fmt.Errorf("%s must be 10MB or smaller", field)
	}
	mime := sniffImageMime(data)
	if !allowed[mime] {
		return nil, fmt.Errorf("%s has unsupported image type", field)
	}
	return &service.UploadedImage{
		Data:      data,
		MimeType:  mime,
		Extension: extensionForImageMime(mime),
		Filename:  header.Filename,
	}, nil
}

func sniffImageMime(b []byte) string {
	if len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff {
		return "image/jpeg"
	}
	if len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WEBP" {
		return "image/webp"
	}
	ct := http.DetectContentType(b)
	if strings.HasPrefix(ct, "image/") {
		return ct
	}
	return "application/octet-stream"
}

func extensionForImageMime(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
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
	for _, assetPath := range []string{gen.SourceImagePath, gen.MaskImagePath} {
		if assetPath == "" {
			continue
		}
		if err := os.Remove(assetPath); err != nil && !os.IsNotExist(err) {
			_ = err
		}
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
