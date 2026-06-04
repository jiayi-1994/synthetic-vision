package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"syntheticvision/internal/auth"
	"syntheticvision/internal/middleware"
	"syntheticvision/internal/models"
)

// signupBonus is the number of credits granted to every newly registered user.
const signupBonus = 1250

// Register creates a new user account, grants the signup bonus, and returns a
// freshly minted JWT alongside the user record.
//
//	POST /api/auth/register {username,email,password} -> 201 {token,user}
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if username == "" || email == "" {
		writeError(w, http.StatusUnprocessableEntity, "username and email are required")
		return
	}
	if !strings.Contains(email, "@") {
		writeError(w, http.StatusUnprocessableEntity, "invalid email address")
		return
	}
	if len(password) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "password must be at least 8 characters")
		return
	}

	// Uniqueness pre-check for clean 409 responses (the unique index is the
	// authoritative guard against races below).
	var existing models.User
	err := h.DB.Where("email = ? OR username = ?", email, username).First(&existing).Error
	if err == nil {
		writeError(w, http.StatusConflict, "email or username already taken")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		ID:           uuid.NewString(),
		PublicID:     newPublicID(),
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         "user",
		Plan:         "free",
		Credits:      signupBonus,
		AvatarSeed:   newAvatarSeed(),
	}

	txErr := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		bonus := models.CreditTransaction{
			ID:     uuid.NewString(),
			UserID: user.ID,
			Amount: signupBonus,
			Reason: "signup_bonus",
		}
		return tx.Create(&bonus).Error
	})
	if txErr != nil {
		// Unique-constraint violation from a concurrent insert maps to 409.
		writeError(w, http.StatusConflict, "email or username already taken")
		return
	}

	token, err := h.issueToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  user,
	})
}

// Login authenticates a user by email + password and returns a JWT.
//
//	POST /api/auth/login {email,password} -> 200 {token,user}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Touch last-activity so the admin directory reflects recent logins.
	h.DB.Model(&user).Update("updated_at", time.Now())

	token, err := h.issueToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

// Me returns the currently authenticated user.
//
//	GET /api/auth/me -> {user}
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)
	var user models.User
	if err := h.DB.Where("id = ?", uid).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// issueToken signs a JWT for the given user using the configured secret and TTL.
func (h *Handler) issueToken(user models.User) (string, error) {
	ttl := time.Duration(h.Cfg.JWTTTLHours) * time.Hour
	return auth.GenerateToken(h.Cfg.JWTSecret, user.ID, user.Role, ttl)
}
