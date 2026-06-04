package service

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"syntheticvision/internal/config"
	"syntheticvision/internal/db"
	"syntheticvision/internal/models"
	"syntheticvision/internal/provider"
)

// failingProvider always errors, exercising the refund path.
type failingProvider struct{}

func (failingProvider) Name() string { return "failing" }
func (failingProvider) Generate(context.Context, provider.GenerateRequest) (*provider.GenerateResult, error) {
	return nil, errors.New("boom")
}

func newTestService(t *testing.T, prov provider.Provider) *Service {
	t.Helper()
	dir := t.TempDir()
	gdb, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Close the SQLite handle before TempDir cleanup, else Windows refuses to
	// remove the still-open .db file.
	if sqlDB, derr := gdb.DB(); derr == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	return New(gdb, prov, config.Config{DataDir: dir, Workers: 4})
}

func seedTestUser(t *testing.T, s *Service, credits int) string {
	t.Helper()
	id := uuid.NewString()
	u := &models.User{
		ID:           id,
		PublicID:     "USR-" + id[:5],
		Username:     "tester",
		Email:        id + "@example.com",
		PasswordHash: "x",
		Role:         "user",
		Plan:         "free",
		Credits:      credits,
	}
	if err := s.db.Create(u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

// TestConcurrentDeductNoDoubleSpend fires more generation requests than the
// balance can fund, concurrently. Exactly the funded number must succeed, the
// balance must land at 0 (never negative, never over-credited), and no request
// may fail with a SQLITE_BUSY-style error — which validates both the atomic
// SQL decrement and the single-connection pool pin (db.Open).
func TestConcurrentDeductNoDoubleSpend(t *testing.T) {
	s := newTestService(t, failingProvider{})
	// Note: Start() is intentionally NOT called — this isolates the synchronous
	// credit-deduction path from the async worker pool.

	const cost = 5 // 1K
	const funded = 4
	uid := seedTestUser(t, s, cost*funded)

	var ok, insufficient int64
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.CreateGeneration(uid, CreateGenInput{
				Prompt:      "concurrent",
				Resolution:  "1K",
				AspectRatio: "1:1",
			})
			switch {
			case err == nil:
				atomic.AddInt64(&ok, 1)
			case errors.Is(err, ErrInsufficientCredits):
				atomic.AddInt64(&insufficient, 1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if ok != funded {
		t.Errorf("expected %d successes, got ok=%d insufficient=%d", funded, ok, insufficient)
	}
	var u models.User
	if err := s.db.Where("id = ?", uid).First(&u).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if u.Credits != 0 {
		t.Errorf("expected final credits 0, got %d", u.Credits)
	}
}

// TestDoubleRefundGuard simulates a duplicate job delivery: several workers call
// process() on the same id concurrently with a failing provider. The atomic
// pending->processing claim plus the guarded fail() must refund exactly once.
func TestDoubleRefundGuard(t *testing.T) {
	s := newTestService(t, failingProvider{})

	const cost = 15 // 2K
	uid := seedTestUser(t, s, cost)

	gen, err := s.CreateGeneration(uid, CreateGenInput{
		Prompt:      "refund",
		Resolution:  "2K",
		AspectRatio: "1:1",
	})
	if err != nil {
		t.Fatalf("create generation: %v", err)
	}

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.process(gen.ID)
		}()
	}
	wg.Wait()

	var u models.User
	if err := s.db.Where("id = ?", uid).First(&u).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if u.Credits != cost {
		t.Errorf("expected exactly one refund (credits=%d), got %d", cost, u.Credits)
	}

	var refunds int64
	s.db.Model(&models.CreditTransaction{}).
		Where("user_id = ? AND reason = ?", uid, "refund").
		Count(&refunds)
	if refunds != 1 {
		t.Errorf("expected 1 refund transaction, got %d", refunds)
	}

	var g models.Generation
	s.db.Where("id = ?", gen.ID).First(&g)
	if g.Status != "failed" {
		t.Errorf("expected status failed, got %q", g.Status)
	}
}
