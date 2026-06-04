package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// defaultJWTSecret is the well-known dev fallback. Booting with it (or any
// secret shorter than 32 bytes) is refused in production — see validateSecurity.
const defaultJWTSecret = "dev-insecure-change-me"

// Config holds all runtime configuration, populated from environment
// variables with sensible defaults via Load.
type Config struct {
	Port          string // PORT, default "8080"
	DataDir       string // DATA_DIR, default "./data"
	JWTSecret     string // JWT_SECRET, default "dev-insecure-change-me"
	JWTTTLHours   int    // JWT_TTL_HOURS, default 168
	ImageProvider string // IMAGE_PROVIDER, default "mock"  ("mock"|"openai")
	OpenAIBaseURL string // OPENAI_BASE_URL
	OpenAIAPIKey  string // OPENAI_API_KEY
	ImageModel    string // IMAGE_MODEL, default "dall-e-3"
	MockDelayMs   int    // MOCK_DELAY_MS, default 2500
	AdminEmail    string // ADMIN_EMAIL, default "admin@synthetic.ai"
	AdminPassword string // ADMIN_PASSWORD, default "synthavision"
	Workers       int    // GEN_WORKERS, default 4
	SeedDemoUsers bool   // SEED_DEMO_USERS, default true
}

// Load reads the environment and returns a fully-populated Config,
// applying defaults for any unset values.
func Load() Config {
	// Load a local .env (if present) before reading the environment, so a bare
	// `go run` picks up the same config as `docker compose`. Real environment
	// variables always win — see loadEnvFile.
	loadEnvFile()
	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		DataDir:       getEnv("DATA_DIR", "./data"),
		JWTSecret:     getEnv("JWT_SECRET", defaultJWTSecret),
		JWTTTLHours:   getEnvInt("JWT_TTL_HOURS", 168),
		ImageProvider: getEnv("IMAGE_PROVIDER", "mock"),
		OpenAIBaseURL: getEnv("OPENAI_BASE_URL", ""),
		OpenAIAPIKey:  getEnv("OPENAI_API_KEY", ""),
		ImageModel:    getEnv("IMAGE_MODEL", "dall-e-3"),
		MockDelayMs:   getEnvInt("MOCK_DELAY_MS", 2500),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@synthetic.ai"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "synthavision"),
		Workers:       getEnvInt("GEN_WORKERS", 4),
		SeedDemoUsers: getEnvBool("SEED_DEMO_USERS", true),
	}
	validateSecurity(cfg)
	return cfg
}

// validateSecurity refuses to boot with the insecure default JWT secret in any
// production-flavored configuration, and warns loudly otherwise. A forgeable
// secret lets anyone mint an admin token, so this is a hard stop where it counts.
func validateSecurity(cfg Config) {
	if cfg.JWTSecret != defaultJWTSecret && len(cfg.JWTSecret) >= 32 {
		return
	}
	prod := cfg.ImageProvider == "openai" || getEnvBool("PRODUCTION", false)
	if prod {
		log.Fatal("config: refusing to start — JWT_SECRET is unset/default/too short. " +
			"Set JWT_SECRET to a random value of at least 32 characters.")
	}
	log.Println("config: WARNING — JWT_SECRET is the insecure default or shorter than 32 chars. " +
		"Set JWT_SECRET to a strong random value before any real deployment.")
}

// ImagesDir returns the directory where generated PNGs are stored.
func (c Config) ImagesDir() string { return filepath.Join(c.DataDir, "images") }

// DBPath returns the SQLite database file path.
func (c Config) DBPath() string { return filepath.Join(c.DataDir, "synthetic-vision.db") }

// loadEnvFile reads a dotenv file (KEY=VALUE per line) if one exists and sets
// any keys NOT already present in the environment. It looks in the working dir
// and one level up, so `go run ./cmd/server` (run from backend/) still finds a
// project-root .env. Real environment variables (e.g. docker compose env_file)
// always take precedence and are never overwritten.
func loadEnvFile() {
	for _, p := range []string{".env", filepath.Join("..", ".env")} {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		applyDotEnv(string(data))
		log.Printf("config: loaded environment overrides from %s", p)
		return
	}
}

func applyDotEnv(content string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.Trim(strings.TrimSpace(line[eq+1:]), `"'`)
		if key == "" {
			continue
		}
		if _, ok := os.LookupEnv(key); ok {
			continue // never override a real environment variable
		}
		_ = os.Setenv(key, val)
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
