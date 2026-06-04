package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"syntheticvision/internal/auth"
	"syntheticvision/internal/config"
	"syntheticvision/internal/db"
	"syntheticvision/internal/handlers"
	"syntheticvision/internal/models"
	"syntheticvision/internal/provider"
	"syntheticvision/internal/server"
	"syntheticvision/internal/service"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	if err := os.MkdirAll(cfg.ImagesDir(), 0o755); err != nil {
		log.Fatalf("create images dir: %v", err)
	}

	gdb, err := db.Open(cfg.DBPath())
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	if err := seed(gdb, cfg); err != nil {
		log.Fatalf("seed: %v", err)
	}

	prov := provider.New(cfg)
	log.Printf("image provider: %s", prov.Name())

	svc := service.New(gdb, prov, cfg)
	svc.Start()
	defer svc.Stop()

	h := handlers.New(gdb, svc, cfg)
	r := server.Router(h, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("synthetic-vision listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	svc.Stop()
	log.Println("bye")
}

// demoUser is a fixed demo account seeded for design fidelity.
type demoUser struct {
	username string
	email    string
	credits  int
	publicID string
}

var demoUsers = []demoUser{
	{username: "j.smith_ai", email: "john.smith@example.com", credits: 4500, publicID: "USR-8829A"},
	{username: "elara.design", email: "elara@studio.io", credits: 120, publicID: "USR-9102B"},
	{username: "m.k_research", email: "mk@university.edu", credits: 0, publicID: "USR-1044C"},
}

// seed creates the admin account on first run and, if enabled and no demo
// users exist yet, the three fixed demo accounts.
func seed(gdb *gorm.DB, cfg config.Config) error {
	var adminCount int64
	if err := gdb.Model(&models.User{}).Where("email = ?", cfg.AdminEmail).Count(&adminCount).Error; err != nil {
		return err
	}
	if adminCount == 0 {
		hash, err := auth.HashPassword(cfg.AdminPassword)
		if err != nil {
			return err
		}
		admin := &models.User{
			ID:           uuid.NewString(),
			PublicID:     newPublicID(),
			Username:     "operator",
			Email:        cfg.AdminEmail,
			PasswordHash: hash,
			Role:         "admin",
			Plan:         "pro",
			Credits:      1250,
			AvatarSeed:   newAvatarSeed(),
		}
		if err := gdb.Create(admin).Error; err != nil {
			return err
		}
		log.Printf("seeded admin account: email=%s (password from ADMIN_PASSWORD)", cfg.AdminEmail)
	}

	if !cfg.SeedDemoUsers {
		return nil
	}

	var demoCount int64
	if err := gdb.Model(&models.User{}).Where("is_demo = ?", true).Count(&demoCount).Error; err != nil {
		return err
	}
	if demoCount > 0 {
		return nil
	}

	hash, err := auth.HashPassword("demo-pass-1234")
	if err != nil {
		return err
	}
	for _, d := range demoUsers {
		u := &models.User{
			ID:           uuid.NewString(),
			PublicID:     d.publicID,
			Username:     d.username,
			Email:        d.email,
			PasswordHash: hash,
			Role:         "user",
			Plan:         "free",
			Credits:      d.credits,
			AvatarSeed:   newAvatarSeed(),
			IsDemo:       true,
		}
		if err := gdb.Create(u).Error; err != nil {
			return err
		}
	}
	log.Printf("seeded %d demo users", len(demoUsers))
	return nil
}

func newPublicID() string {
	id := uuid.New()
	b := id[:]
	const hexUpper = "0123456789ABCDEF"
	out := make([]byte, 5)
	for i := 0; i < 5; i++ {
		out[i] = hexUpper[int(b[i])%16]
	}
	return "USR-" + string(out)
}

func newAvatarSeed() string {
	return uuid.NewString()[:12]
}
