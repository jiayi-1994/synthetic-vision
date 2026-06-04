package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"syntheticvision/internal/middleware"
	"syntheticvision/internal/models"
	"syntheticvision/internal/service"
)

// adminUserRow is the computed projection of a User for the admin directory.
// It deliberately omits sensitive/internal fields and adds the display-only
// `last_activity_at` and `initials`.
type adminUserRow struct {
	PublicID       string `json:"public_id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Credits        int    `json:"credits"`
	Role           string `json:"role"`
	Plan           string `json:"plan"`
	LastActivityAt string `json:"last_activity_at"`
	Initials       string `json:"initials"`
	CreatedAt      string `json:"created_at"`
}

// AdminUsers lists users with search + pagination for the admin directory.
//
//	GET /api/admin/users?search=&page=1&page_size=10
//	-> {users:[adminUser],total,page,page_size}
func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(strings.TrimSpace(q.Get("page")))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(strings.TrimSpace(q.Get("page_size")))
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	base := h.DB.Model(&models.User{})
	if search := strings.TrimSpace(q.Get("search")); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		base = base.Where(
			"LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(public_id) LIKE ?",
			like, like, like,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count users")
		return
	}

	var users []models.User
	if err := base.
		Order("updated_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&users).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	rows := make([]adminUserRow, 0, len(users))
	for _, u := range users {
		rows = append(rows, adminUserRow{
			PublicID:       u.PublicID,
			Username:       u.Username,
			Email:          u.Email,
			Credits:        u.Credits,
			Role:           u.Role,
			Plan:           u.Plan,
			LastActivityAt: u.UpdatedAt.Format(timeRFC3339),
			Initials:       initialsOf(u.Username),
			CreatedAt:      u.CreatedAt.Format(timeRFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"users":     rows,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AdminInjectCredits adds credits to a target user by public_id.
//
//	POST /api/admin/credits {target_public_id, amount} -> {user}
func (h *Handler) AdminInjectCredits(w http.ResponseWriter, r *http.Request) {
	var req injectCreditsRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	target := strings.TrimSpace(req.TargetPublicID)
	if target == "" {
		writeError(w, http.StatusUnprocessableEntity, "target_public_id is required")
		return
	}
	if req.Amount == 0 {
		writeError(w, http.StatusUnprocessableEntity, "amount must be non-zero")
		return
	}

	adminID := middleware.UserID(r)
	user, err := h.Svc.AdminInjectCredits(adminID, target, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam):
			writeError(w, http.StatusUnprocessableEntity, "invalid parameter")
		default:
			writeError(w, http.StatusNotFound, "target user not found")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// AdminCluster reports a synthetic compute-cluster status for the admin UI.
//
//	GET /api/admin/cluster -> {status:"Operational", load_percent:<int 60-99>}
func (h *Handler) AdminCluster(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "Operational",
		"load_percent": randInt(60, 99),
	})
}

// timeRFC3339 is the time layout used for JSON timestamps emitted by admin rows.
const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

// initialsOf returns a two-letter avatar monogram derived from the username.
// It splits on common separators (. _ - space) and takes the first letter of
// the first two segments (e.g. "j.smith_ai" -> "JS", "m.k_research" -> "MK").
// For a single segment it uses the first two letters ("operator" -> "OP").
// Falls back gracefully for short or empty names.
func initialsOf(name string) string {
	segs := strings.FieldsFunc(name, func(r rune) bool {
		return r == '.' || r == '_' || r == '-' || r == ' '
	})
	if len(segs) >= 2 {
		return strings.ToUpper(firstRune(segs[0]) + firstRune(segs[1]))
	}
	if len(segs) == 1 {
		runes := []rune(segs[0])
		if len(runes) >= 2 {
			return strings.ToUpper(string(runes[:2]))
		}
		return strings.ToUpper(string(runes))
	}
	return "??"
}

// firstRune returns the first rune of s as a string ("" if s is empty).
func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}

// randInt returns a cryptographically-random int in [min, max] inclusive.
func randInt(min, max int) int {
	if max <= min {
		return min
	}
	span := int64(max - min + 1)
	n, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		return min
	}
	return min + int(n.Int64())
}

// newPublicID returns a unique-ish human id of the form "USR-XXXXX" (5 hex
// upper). Used at user registration time.
func newPublicID() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		// Fallback: deterministic-but-still-formatted id.
		return "USR-00000"
	}
	return "USR-" + strings.ToUpper(hex.EncodeToString(b))[:5]
}

// newAvatarSeed returns a random hex string used to derive a deterministic
// DiceBear avatar on the frontend.
func newAvatarSeed() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "seed"
	}
	return hex.EncodeToString(b)
}
