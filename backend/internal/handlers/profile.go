package handlers

import (
	"net/http"

	"syntheticvision/internal/middleware"
	"syntheticvision/internal/models"
)

// Stats returns the authenticated user's profile alongside aggregate
// generation counts and the current credit balance.
//
//	GET /api/me/stats -> {user, total_generations, completed_generations, credit_balance}
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)

	var user models.User
	if err := h.DB.Where("id = ?", uid).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var total int64
	if err := h.DB.Model(&models.Generation{}).
		Where("user_id = ?", uid).
		Count(&total).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count generations")
		return
	}

	var completed int64
	if err := h.DB.Model(&models.Generation{}).
		Where("user_id = ? AND status = ?", uid, "completed").
		Count(&completed).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count generations")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":                  user,
		"total_generations":     total,
		"completed_generations": completed,
		"credit_balance":        user.Credits,
	})
}
