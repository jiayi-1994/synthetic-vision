package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"syntheticvision/internal/config"
	"syntheticvision/internal/db"
	"syntheticvision/internal/middleware"
	"syntheticvision/internal/models"
)

func TestAnalyticsHandlerPersonalizedAggregation(t *testing.T) {
	t.Parallel()

	h, _ := newAnalyticsTestFixture(t)
	uidA := seedAnalyticsUser(t, h.DB, 120, "creator-a", "alice@example.com")
	uidB := seedAnalyticsUser(t, h.DB, 80, "creator-b", "bob@example.com")

	now := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	seedAnalyticsGeneration(t, h.DB, uidA, "completed", now.Add(-3*time.Hour), "1K", "1:1", 5, "晨光林间")
	seedAnalyticsGeneration(t, h.DB, uidA, "failed", now.Add(-2*time.Hour), "2K", "16:9", 15, "霓虹赛博")
	seedAnalyticsGeneration(t, h.DB, uidA, "pending", now.Add(-1*time.Hour), "2K", "4:3", 15, "静物研究")
	seedAnalyticsGeneration(t, h.DB, uidA, "processing", now.Add(-30*time.Minute), "1K", "9:16", 5, "草地")

	seedCreditTxn(t, h.DB, uidA, "generation", -5)
	seedCreditTxn(t, h.DB, uidA, "generation", -15)
	seedCreditTxn(t, h.DB, uidA, "generation", -15)
	seedCreditTxn(t, h.DB, uidA, "refund", 15)
	seedCreditTxn(t, h.DB, uidA, "signup_bonus", 1250)
	seedCreditTxn(t, h.DB, uidA, "admin_injection", 1200)
	seedCreditTxn(t, h.DB, uidB, "generation", -99)

	res := getAnalyticsResponse(t, h, uidA)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}

	var payload AnalyticsResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if payload.User.ID != uidA {
		t.Fatalf("expected user %s, got %s", uidA, payload.User.ID)
	}
	if payload.Summary.TotalGenerations != 4 {
		t.Fatalf("total_generations=%d", payload.Summary.TotalGenerations)
	}
	if payload.Summary.CompletedGenerations != 1 || payload.Summary.FailedGenerations != 1 ||
		payload.Summary.PendingGenerations != 1 || payload.Summary.ProcessingGenerations != 1 {
		t.Fatalf("unexpected status breakdown: %+v", payload.Summary)
	}
	if payload.Summary.SuccessRate != 25 {
		t.Fatalf("success_rate=%v", payload.Summary.SuccessRate)
	}
	if payload.Summary.CreditsSpent != 35 || payload.Summary.CreditsRefunded != 15 {
		t.Fatalf("credits summary mismatch: %+v", payload.Summary)
	}
	if payload.CreditBreakdown.AdminTopups != 1200 || payload.CreditBreakdown.SignupBonus != 1250 {
		t.Fatalf("credit breakdown mismatch: %+v", payload.CreditBreakdown)
	}
	if len(payload.ResolutionDistribution) != 3 || payload.ResolutionDistribution[0].Count != 2 || payload.ResolutionDistribution[1].Count != 2 {
		t.Fatalf("resolution distribution mismatch: %+v", payload.ResolutionDistribution)
	}
	if len(payload.AspectDistribution) != 4 {
		t.Fatalf("aspect distribution mismatch: %+v", payload.AspectDistribution)
	}
	if payload.RecentGenerations[0].Prompt != "草地" || payload.RecentGenerations[0].Resolution != "1K" {
		t.Fatalf("recent order should be newest-first: %+v", payload.RecentGenerations[:1])
	}
}

func TestAnalyticsHandlerEmptyUserGetsZeroState(t *testing.T) {
	t.Parallel()

	h, _ := newAnalyticsTestFixture(t)
	uid := seedAnalyticsUser(t, h.DB, 10, "new", "new@example.com")

	res := getAnalyticsResponse(t, h, uid)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}

	var payload AnalyticsResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if payload.Summary.TotalGenerations != 0 || payload.Summary.CreditsSpent != 0 || payload.Summary.CreditsRefunded != 0 {
		t.Fatalf("expected zero metrics, got: %+v", payload.Summary)
	}
	if payload.Summary.SuccessRate != 0 {
		t.Fatalf("expected success_rate 0, got %v", payload.Summary.SuccessRate)
	}
	if len(payload.RecentGenerations) != 0 {
		t.Fatalf("expected no recent generations, got %d", len(payload.RecentGenerations))
	}
	if payload.ResolutionDistribution[0].Count != 0 || payload.AspectDistribution[0].Count != 0 {
		t.Fatalf("expected zero distributions, got %v / %v", payload.ResolutionDistribution, payload.AspectDistribution)
	}
}

func TestAnalyticsHandlerUserIsolation(t *testing.T) {
	t.Parallel()

	h, _ := newAnalyticsTestFixture(t)
	uidA := seedAnalyticsUser(t, h.DB, 15, "owner", "owner@example.com")
	uidB := seedAnalyticsUser(t, h.DB, 15, "intruder", "intruder@example.com")

	seedAnalyticsGeneration(t, h.DB, uidA, "completed", time.Now(), "1K", "1:1", 5, "A1")
	seedAnalyticsGeneration(t, h.DB, uidA, "completed", time.Now(), "1K", "4:3", 5, "A2")
	seedCreditTxn(t, h.DB, uidA, "generation", -10)

	seedAnalyticsGeneration(t, h.DB, uidB, "failed", time.Now(), "2K", "16:9", 15, "B1")
	seedCreditTxn(t, h.DB, uidB, "generation", -15)

	res := getAnalyticsResponse(t, h, uidA)
	var payload AnalyticsResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Summary.TotalGenerations != 2 {
		t.Fatalf("expected 2 generations for uidA, got %d", payload.Summary.TotalGenerations)
	}
	if payload.Summary.CreditsSpent != 10 {
		t.Fatalf("expected uidA spent 10, got %v", payload.Summary.CreditsSpent)
	}
}

func TestAnalyticsHandlerRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	h, _ := newAnalyticsTestFixture(t)
	_ = seedAnalyticsUser(t, h.DB, 5, "anonymous", "anon@example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/me/analytics", nil)
	// No middleware context => this handler should not leak user data.
	rec := httptest.NewRecorder()
	h.Analytics(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on missing user context, got %d", rec.Code)
	}
}

func getAnalyticsResponse(t *testing.T, h *Handler, uid string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/me/analytics", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uid)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.Analytics(rec, req)
	return rec
}

func newAnalyticsTestFixture(t *testing.T) (*Handler, *gorm.DB) {
	t.Helper()
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "analytics.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if sqlDB, err := database.DB(); err == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	return New(database, nil, config.Config{}), database
}

func seedAnalyticsUser(t *testing.T, gdb *gorm.DB, credits int, username, email string) string {
	t.Helper()
	id := uuid.NewString()
	u := &models.User{
		ID:           id,
		PublicID:     "USR-" + strings.ToUpper(id[:5]),
		Username:     username,
		Email:        email,
		PasswordHash: "hash",
		Role:         "user",
		Plan:         "free",
		Credits:      credits,
		AvatarSeed:   "seed-" + id[:8],
	}
	if err := gdb.Create(u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func seedAnalyticsGeneration(t *testing.T, gdb *gorm.DB, uid, status string, created time.Time, resolution, aspect string, cost int, prompt string) {
	t.Helper()
	id := uuid.NewString()
	now := created
	var completedAt *time.Time
	if status == "completed" || status == "failed" {
		c := now.Add(2 * time.Minute)
		completedAt = &c
	}

	// Keep dimensions stable with supported combinations.
	width, height := 1024, 1024
	switch resolution {
	case "2K":
		width, height = 2048, 2048
	case "4K":
		width, height = 4096, 4096
	}
	switch aspect {
	case "4:3":
		height = int(float64(width) * 3.0 / 4.0)
	case "16:9":
		height = int(float64(width) * 9.0 / 16.0)
	case "9:16":
		width = int(float64(height) * 9.0 / 16.0)
	}

	gen := &models.Generation{
		ID:             id,
		UserID:         uid,
		Prompt:         prompt,
		Resolution:     resolution,
		AspectRatio:    aspect,
		Status:         status,
		Cost:           cost,
		Width:          width,
		Height:         height,
		Seed:           int64(1000 + len(id)),
		CreatedAt:      now,
		CompletedAt:    completedAt,
		ImageURL:       "/images/" + id + ".png",
		NegativePrompt: "",
		Style:          "Cinematic",
	}
	if status == "failed" {
		gen.Error = "generation failed"
	}
	if err := gdb.Create(gen).Error; err != nil {
		t.Fatalf("seed generation: %v", err)
	}
}

func seedCreditTxn(t *testing.T, gdb *gorm.DB, uid, reason string, amount int) {
	t.Helper()
	err := gdb.Create(&models.CreditTransaction{
		ID:        uuid.NewString(),
		UserID:    uid,
		Amount:    amount,
		Reason:    reason,
		CreatedAt: time.Now(),
	}).Error
	if err != nil {
		t.Fatalf("seed credit txn: %v", err)
	}
}
