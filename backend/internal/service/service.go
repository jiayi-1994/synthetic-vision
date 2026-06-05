package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"syntheticvision/internal/config"
	"syntheticvision/internal/models"
	"syntheticvision/internal/provider"
)

// Sentinel errors returned by the service layer.
var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrInvalidParam        = errors.New("invalid parameter")
	ErrUserNotFound        = errors.New("user not found")
)

// CreateGenInput is the validated input for CreateGeneration.
type CreateGenInput struct {
	Mode           string
	Prompt         string
	NegativePrompt string
	Style          string
	Resolution     string
	AspectRatio    string
	SourceImage    *UploadedImage
	MaskImage      *UploadedImage
}

// UploadedImage is a bounded in-memory upload handed from the HTTP layer to
// the service so the pending generation can persist its private source assets.
type UploadedImage struct {
	Data      []byte
	MimeType  string
	Extension string
	Filename  string
}

// Service owns the generation worker pool and all credit-mutating logic.
type Service struct {
	db    *gorm.DB
	prov  provider.Provider
	cfg   config.Config
	queue chan string

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
}

// New constructs a Service. Call Start to launch workers.
func New(db *gorm.DB, prov provider.Provider, cfg config.Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		db:     db,
		prov:   prov,
		cfg:    cfg,
		queue:  make(chan string, 256),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start launches cfg.Workers goroutines consuming the queue. After start it
// re-enqueues any jobs left in pending/processing from a previous run.
func (s *Service) Start() {
	workers := s.cfg.Workers
	if workers < 1 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.worker()
	}
	s.requeuePending()
	go s.dispatchLoop()
}

// dispatchLoop periodically re-scans pending rows and tops up the worker queue.
// Because enqueue is non-blocking (it never stalls the HTTP request goroutine),
// a job created while the queue is full is left as "pending" and picked up here
// on the next tick. The atomic claim in process() makes a duplicate enqueue
// harmless (only one worker wins the pending->processing transition).
func (s *Service) dispatchLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.scanPending()
		}
	}
}

func (s *Service) scanPending() {
	var ids []string
	if err := s.db.Model(&models.Generation{}).
		Where("status = ?", "pending").
		Order("created_at asc").
		Limit(256).
		Pluck("id", &ids).Error; err != nil {
		log.Printf("service: dispatch scan failed: %v", err)
		return
	}
	for _, id := range ids {
		s.enqueue(id)
	}
}

// Stop signals the workers to drain and blocks until they exit.
func (s *Service) Stop() {
	s.once.Do(func() {
		s.cancel()
		close(s.queue)
	})
	s.wg.Wait()
}

// CostFor returns the credit cost for a resolution, or -1 if unknown.
func CostFor(resolution string) int {
	switch resolution {
	case "1K":
		return 5
	case "2K":
		return 15
	case "4K":
		return 40
	default:
		return -1
	}
}

// ResolveDimensions returns the nominal (w,h) for a resolution + aspect ratio.
// The longest side receives the base size; the other side is scaled by the
// ratio and rounded to an even integer.
func ResolveDimensions(resolution, aspect string) (w, h int) {
	var base int
	switch resolution {
	case "1K":
		base = 1024
	case "2K":
		base = 2048
	case "4K":
		base = 4096
	default:
		return 0, 0
	}
	switch aspect {
	case "1:1":
		return base, base
	case "4:3":
		// landscape: width longest
		return base, evenInt(float64(base) * 3.0 / 4.0)
	case "16:9":
		return base, evenInt(float64(base) * 9.0 / 16.0)
	case "9:16":
		// portrait: height longest
		return evenInt(float64(base) * 9.0 / 16.0), base
	default:
		return 0, 0
	}
}

// renderDimensions scales the nominal dims so the longest side is at most 1024,
// preserving aspect ratio and rounding to even ints. Keeps mock PNGs small.
func renderDimensions(nw, nh int) (int, int) {
	longest := nw
	if nh > longest {
		longest = nh
	}
	if longest <= 1024 {
		return nw, nh
	}
	scale := 1024.0 / float64(longest)
	return evenInt(float64(nw) * scale), evenInt(float64(nh) * scale)
}

func evenInt(v float64) int {
	n := int(v + 0.5)
	if n%2 != 0 {
		n++
	}
	if n < 2 {
		n = 2
	}
	return n
}

// CreateGeneration validates input, deducts cost atomically, creates a pending
// Generation row, then enqueues it for async rendering.
func (s *Service) CreateGeneration(userID string, in CreateGenInput) (*models.Generation, error) {
	mode := normalizeMode(in.Mode)
	if mode == "text" && in.SourceImage != nil {
		mode = "image"
		if in.MaskImage != nil {
			mode = "edit"
		}
	}
	if mode == "" {
		return nil, ErrInvalidParam
	}
	if mode == "image" && in.SourceImage == nil {
		return nil, ErrInvalidParam
	}
	if mode == "edit" && (in.SourceImage == nil || in.MaskImage == nil) {
		return nil, ErrInvalidParam
	}
	cost := CostFor(in.Resolution)
	if cost < 0 {
		return nil, ErrInvalidParam
	}
	if !validAspect(in.AspectRatio) {
		return nil, ErrInvalidParam
	}
	if in.Prompt == "" {
		return nil, ErrInvalidParam
	}
	nw, nh := ResolveDimensions(in.Resolution, in.AspectRatio)
	if nw == 0 || nh == 0 {
		return nil, ErrInvalidParam
	}

	seed := time.Now().UnixNano() ^ int64(uuid.New().ID())
	gen := &models.Generation{
		ID:             uuid.NewString(),
		UserID:         userID,
		Mode:           mode,
		Prompt:         in.Prompt,
		NegativePrompt: in.NegativePrompt,
		Resolution:     in.Resolution,
		AspectRatio:    in.AspectRatio,
		Style:          in.Style,
		Width:          nw,
		Height:         nh,
		Seed:           seed,
		Status:         "pending",
		Cost:           cost,
		HasSourceImage: in.SourceImage != nil,
		HasMaskImage:   in.MaskImage != nil,
		CreatedAt:      time.Now(),
	}

	var written []string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		// NOTE: clause.Locking FOR UPDATE is a no-op under the glebarez/sqlite
		// dialector (SQLite has no row-level locking). Correctness of the
		// balance check+decrement instead relies on db.Open pinning the pool to
		// a single connection (SetMaxOpenConns(1)), which serializes writers.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return err
		}
		if user.Credits < cost {
			return ErrInsufficientCredits
		}
		if err := tx.Model(&user).Update("credits", gorm.Expr("credits - ?", cost)).Error; err != nil {
			return err
		}
		txn := &models.CreditTransaction{
			ID:        uuid.NewString(),
			UserID:    userID,
			Amount:    -cost,
			Reason:    "generation",
			CreatedAt: time.Now(),
		}
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		if err := tx.Create(gen).Error; err != nil {
			return err
		}
		updates := map[string]any{}
		if in.SourceImage != nil {
			path, err := s.writeUploadedImage(gen.ID, "source", in.SourceImage)
			if err != nil {
				return err
			}
			written = append(written, path)
			gen.SourceImagePath = path
			updates["source_image_path"] = path
			updates["has_source_image"] = true
		}
		if in.MaskImage != nil {
			path, err := s.writeUploadedImage(gen.ID, "mask", in.MaskImage)
			if err != nil {
				return err
			}
			written = append(written, path)
			gen.MaskImagePath = path
			updates["mask_image_path"] = path
			updates["has_mask_image"] = true
		}
		if len(updates) > 0 {
			if err := tx.Model(gen).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		for _, path := range written {
			_ = os.Remove(path)
		}
		return nil, err
	}

	s.enqueue(gen.ID)
	return gen, nil
}

// AdminInjectCredits adds amount credits to the user identified by public_id and
// records an admin_injection transaction. Returns the updated user.
func (s *Service) AdminInjectCredits(adminID, targetPublicID string, amount int) (*models.User, error) {
	// Injection is an additive top-up only. Reject non-positive amounts (a
	// negative would drive a balance below zero and permanently lock the user
	// out of generating) and cap the upper bound to a sane ceiling.
	if amount <= 0 || amount > 1_000_000 {
		return nil, ErrInvalidParam
	}
	var updated models.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("public_id = ?", targetPublicID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return err
		}
		if err := tx.Model(&user).Update("credits", gorm.Expr("credits + ?", amount)).Error; err != nil {
			return err
		}
		txn := &models.CreditTransaction{
			ID:        uuid.NewString(),
			UserID:    user.ID,
			Amount:    amount,
			Reason:    "admin_injection",
			AdminID:   adminID,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(txn).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", user.ID).First(&updated).Error
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// enqueue offers id to the worker queue without ever blocking. If the queue is
// full the job stays "pending" and dispatchLoop will re-offer it on the next
// tick, so a burst of requests never stalls the HTTP handler goroutine.
func (s *Service) enqueue(id string) {
	defer func() { _ = recover() }() // queue may be closed during shutdown
	select {
	case s.queue <- id:
	default:
	}
}

func (s *Service) requeuePending() {
	// Reset jobs left "processing" by a crashed previous run back to "pending"
	// so the claim in process() (which only matches "pending") can pick them up
	// exactly once.
	if err := s.db.Model(&models.Generation{}).
		Where("status = ?", "processing").
		Update("status", "pending").Error; err != nil {
		log.Printf("service: requeue reset failed: %v", err)
	}
	var ids []string
	if err := s.db.Model(&models.Generation{}).
		Where("status = ?", "pending").
		Order("created_at asc").
		Pluck("id", &ids).Error; err != nil {
		log.Printf("service: requeue scan failed: %v", err)
		return
	}
	for _, id := range ids {
		s.enqueue(id)
	}
}

func (s *Service) worker() {
	defer s.wg.Done()
	for id := range s.queue {
		s.process(id)
	}
}

func (s *Service) process(genID string) {
	// Atomically CLAIM the job: flip pending -> processing in a single UPDATE.
	// SQLite serializes writes, so if two workers race on the same id (e.g. a
	// duplicate enqueue), exactly one UPDATE matches status="pending" and the
	// loser sees RowsAffected==0 and bails — preventing double render/refund.
	claim := s.db.Model(&models.Generation{}).
		Where("id = ? AND status = ?", genID, "pending").
		Update("status", "processing")
	if claim.Error != nil {
		log.Printf("service: claim generation %s failed: %v", genID, claim.Error)
		return
	}
	if claim.RowsAffected == 0 {
		// Already claimed, completed, or failed by another worker. Not ours.
		return
	}

	var gen models.Generation
	if err := s.db.Where("id = ?", genID).First(&gen).Error; err != nil {
		log.Printf("service: load generation %s failed: %v", genID, err)
		return
	}

	rw, rh := renderDimensions(gen.Width, gen.Height)
	sourceImage, err := loadProviderImage(gen.SourceImagePath)
	if err != nil {
		s.fail(&gen, err)
		return
	}
	maskImage, err := loadProviderImage(gen.MaskImagePath)
	if err != nil {
		s.fail(&gen, err)
		return
	}
	result, err := s.prov.Generate(s.ctx, provider.GenerateRequest{
		Mode:           gen.Mode,
		Prompt:         gen.Prompt,
		NegativePrompt: gen.NegativePrompt,
		Style:          gen.Style,
		Width:          rw,
		Height:         rh,
		Seed:           gen.Seed,
		SourceImage:    sourceImage,
		MaskImage:      maskImage,
	})
	if err != nil {
		s.fail(&gen, err)
		return
	}
	if gen.Mode == "edit" && sourceImage != nil && maskImage != nil {
		composited, err := compositeMaskedEdit(result.Image, sourceImage.Data, maskImage.Data)
		if err != nil {
			s.fail(&gen, err)
			return
		}
		result.Image = composited
		result.MimeType = "image/png"
	}

	if err := os.MkdirAll(s.cfg.ImagesDir(), 0o755); err != nil {
		s.fail(&gen, err)
		return
	}
	path := filepath.Join(s.cfg.ImagesDir(), gen.ID+".png")
	if err := os.WriteFile(path, result.Image, 0o644); err != nil {
		s.fail(&gen, err)
		return
	}

	now := time.Now()
	res := s.db.Model(&models.Generation{}).Where("id = ?", gen.ID).Updates(map[string]any{
		"status":       "completed",
		"image_url":    "/images/" + gen.ID + ".png",
		"completed_at": &now,
		"error":        "",
	})
	if res.Error == nil && res.RowsAffected == 0 {
		// Row was deleted while we were rendering (user hit DELETE mid-flight).
		// Don't leave an orphaned PNG on disk.
		_ = os.Remove(path)
	}
}

// fail marks the generation failed and refunds its cost to the user. The status
// transition is an atomic compare-and-set guarded on the row NOT already being
// terminal, and the refund runs only when this call actually performed the
// transition — so a duplicate fail() (or a fail racing a completion) refunds at
// most once. The detailed cause is logged server-side; the user-visible Error
// field gets a generic message to avoid leaking upstream provider/host detail.
func (s *Service) fail(gen *models.Generation, cause error) {
	log.Printf("service: generation %s failed: %v", gen.ID, cause)
	now := time.Now()
	_ = s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Generation{}).
			Where("id = ? AND status NOT IN ?", gen.ID, []string{"failed", "completed"}).
			Updates(map[string]any{
				"status":       "failed",
				"error":        "generation failed",
				"completed_at": &now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// Already terminal (or row gone) — do not double-refund.
			return nil
		}
		if err := tx.Model(&models.User{}).Where("id = ?", gen.UserID).
			Update("credits", gorm.Expr("credits + ?", gen.Cost)).Error; err != nil {
			return err
		}
		refund := &models.CreditTransaction{
			ID:        uuid.NewString(),
			UserID:    gen.UserID,
			Amount:    gen.Cost,
			Reason:    "refund",
			CreatedAt: time.Now(),
		}
		return tx.Create(refund).Error
	})
}

func validAspect(a string) bool {
	switch a {
	case "1:1", "4:3", "16:9", "9:16":
		return true
	default:
		return false
	}
}

func normalizeMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case "", "text":
		return "text"
	case "image", "edit":
		return strings.TrimSpace(mode)
	default:
		return ""
	}
}

func (s *Service) writeUploadedImage(genID, suffix string, img *UploadedImage) (string, error) {
	if img == nil || len(img.Data) == 0 {
		return "", ErrInvalidParam
	}
	if err := os.MkdirAll(s.cfg.ReferencesDir(), 0o755); err != nil {
		return "", err
	}
	ext := normalizedExtension(img)
	path := filepath.Join(s.cfg.ReferencesDir(), genID+"-"+suffix+ext)
	if err := os.WriteFile(path, img.Data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func normalizedExtension(img *UploadedImage) string {
	ext := strings.ToLower(strings.TrimSpace(img.Extension))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		if ext == ".jpeg" {
			return ".jpg"
		}
		return ext
	}
	switch img.MimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

func loadProviderImage(path string) (*provider.InputImage, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &provider.InputImage{
		Data:     data,
		MimeType: detectImageMime(data),
		Filename: filepath.Base(path),
	}, nil
}

func detectImageMime(b []byte) string {
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
