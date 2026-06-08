# Synthetic Vision — Build Spec (single source of truth)

> AI image generation/editing SaaS. Credit-based. Dark glassmorphism UI ("Midnight Spectrum").
> Backend: **Go 1.24** (chi + GORM + pure-Go SQLite). Frontend: **Vue 3 + TS + Vite + Pinia + Tailwind 3 + axios**.
> Design reference: `_design/*.html` (exact Stitch Tailwind markup) and `_design/*.png` (screenshots).
> **Every agent MUST read this whole file and match the names/shapes EXACTLY. Cross-file calls only work if signatures match this spec verbatim.**

Project root: `F:\other code\sub2api-hf\synthetic-vision\` (referred to below as `<root>`). All file paths are absolute under `<root>`.

---

## 0. Directory layout (who writes what)

```
<root>/
├─ README.md                         (OPS)
├─ docker-compose.yml                (OPS)
├─ Dockerfile                        (OPS)  multi-stage: frontend build → embed → go build
├─ .env.example                      (OPS)
├─ .gitignore                        (OPS)
├─ Makefile                          (OPS)
├─ backend/                          (BACKEND owns all of this)
│  ├─ go.mod
│  ├─ cmd/server/main.go
│  ├─ internal/config/config.go
│  ├─ internal/db/db.go
│  ├─ internal/models/models.go
│  ├─ internal/auth/jwt.go
│  ├─ internal/auth/password.go
│  ├─ internal/middleware/middleware.go
│  ├─ internal/provider/provider.go
│  ├─ internal/provider/mock.go
│  ├─ internal/provider/openai.go
│  ├─ internal/service/service.go          (generation worker + credit logic)
│  ├─ internal/handlers/handlers.go        (Handler struct, ctor, json helpers)
│  ├─ internal/handlers/auth.go
│  ├─ internal/handlers/generation.go
│  ├─ internal/handlers/profile.go
│  ├─ internal/handlers/admin.go
│  ├─ internal/server/router.go            (chi routes + static SPA + /images)
│  └─ web/web.go                            (go:embed dist)
│  └─ web/dist/index.html                   (placeholder so embed compiles; Docker overwrites)
└─ frontend/                         (FE-CORE + FE-VIEWS)
   ├─ package.json                   (FE-CORE)
   ├─ vite.config.ts                 (FE-CORE)
   ├─ tailwind.config.js             (FE-CORE)
   ├─ postcss.config.js              (FE-CORE)
   ├─ tsconfig.json                  (FE-CORE)
   ├─ tsconfig.node.json             (FE-CORE)
   ├─ .npmrc                         (FE-CORE)  registry mirror
   ├─ index.html                     (FE-CORE)
   ├─ env.d.ts                       (FE-CORE)
   └─ src/
      ├─ main.ts                     (FE-CORE)
      ├─ App.vue                     (FE-CORE)
      ├─ style.css                   (FE-CORE)
      ├─ types.ts                    (FE-CORE)
      ├─ api/client.ts               (FE-CORE)
      ├─ router/index.ts             (FE-CORE)
      ├─ stores/auth.ts              (FE-CORE)
      ├─ stores/generations.ts       (FE-CORE)
      ├─ lib/format.ts               (FE-CORE)  relative-time + cost helpers
      ├─ components/AppShell.vue      (FE-CORE)
      ├─ components/Sidebar.vue       (FE-CORE)
      ├─ components/TopBar.vue        (FE-CORE)
      ├─ components/ImageCard.vue     (FE-VIEWS)
      ├─ views/Login.vue             (FE-VIEWS)
      ├─ views/Dashboard.vue         (FE-VIEWS)
      ├─ views/Gallery.vue           (FE-VIEWS)
      ├─ views/Admin.vue             (FE-VIEWS)
      ├─ views/Marketplace.vue       (FE-VIEWS)
      ├─ views/Analytics.vue         (FE-VIEWS)
      └─ lib/presets.ts              (FE-CORE)
```

---

## 1. Design tokens (use VERBATIM in tailwind.config.js and Go where noted)

Colors (Tailwind custom colors, hex):
```
primary #d2bbff   on-primary #3f008e   primary-container #7c3aed   on-primary-container #ede0ff
primary-fixed #eaddff   primary-fixed-dim #d2bbff   on-primary-fixed #25005a   on-primary-fixed-variant #5a00c6
inverse-primary #732ee4   surface-tint #d2bbff
secondary #89ceff   on-secondary #00344d   secondary-container #00a2e6   on-secondary-container #00344e
secondary-fixed #c9e6ff   secondary-fixed-dim #89ceff   on-secondary-fixed #001e2f   on-secondary-fixed-variant #004c6e
tertiary #ffb784   on-tertiary #4f2500   tertiary-container #a15100   on-tertiary-container #ffe0cd
tertiary-fixed #ffdcc6   tertiary-fixed-dim #ffb784   on-tertiary-fixed #301400   on-tertiary-fixed-variant #713700
error #ffb4ab   on-error #690005   error-container #93000a   on-error-container #ffdad6
background #0b1326   on-background #dae2fd
surface #0b1326   surface-dim #0b1326   surface-bright #31394d   on-surface #dae2fd   on-surface-variant #ccc3d8
surface-container-lowest #060e20   surface-container-low #131b2e   surface-container #171f33
surface-container-high #222a3d   surface-container-highest #2d3449   surface-variant #2d3449
outline #958da1   outline-variant #4a4455   inverse-surface #dae2fd   inverse-on-surface #283044
```
borderRadius: `DEFAULT 0.25rem`, `lg 0.5rem`, `xl 0.75rem`, `full 9999px` (KEEP only these — Stitch overrides default scale).
spacing extras: `margin-lg 40px`, `margin-sm 16px`, `gutter 24px`, `unit 4px`, `container-max 1440px`.
fontFamily: `display-lg/body-md/headline-lg/headline-lg-mobile ["Inter"]`, `label-sm ["JetBrains Mono"]`.
fontSize: `display-lg [48px,{lineHeight 1.1,letterSpacing -0.02em,fontWeight 700}]`, `headline-lg [32px,{1.2,-0.01em,600}]`, `headline-lg-mobile [24px,{1.2,_,600}]`, `body-md [16px,{1.6,_,400}]`, `label-sm [12px,{1.0,0.05em,500}]`.
darkMode: `"class"`. `<html class="dark">`.

Global CSS (style.css) — copy these utilities exactly (from `_design`):
```css
body { background-color:#020617; color:#dae2fd; }
.glass-panel { background-color:rgba(30,41,59,0.6); backdrop-filter:blur(20px); -webkit-backdrop-filter:blur(20px); border:1px solid rgba(51,65,85,0.5); }
.glow-shadow { box-shadow:0 8px 32px rgba(124,58,237,0.15); }
.glow-hover:hover { box-shadow:0 8px 32px rgba(124,58,237,0.15); border-color:rgba(210,187,255,0.3); }
::-webkit-scrollbar{width:8px} ::-webkit-scrollbar-track{background:transparent}
::-webkit-scrollbar-thumb{background:#4a4455;border-radius:4px} ::-webkit-scrollbar-thumb:hover{background:#958da1}
.material-symbols-outlined{font-variation-settings:'FILL' 0,'wght' 300,'GRAD' 0,'opsz' 24}
```
Fonts + icons: `index.html` `<head>` includes the Google Fonts links from `_design/login.html` lines 8–13 (Material Symbols Outlined, Inter 400;600;700, JetBrains Mono 500). Add `@tailwind base/components/utilities` in style.css.

Icons: Material Symbols Outlined `<span class="material-symbols-outlined">name</span>`. Nav icons: Dashboard `auto_awesome`, Gallery `grid_view`, Marketplace `storefront`, Admin `admin_panel_settings`, Analytics `insights`, Support `help_outline`, Settings `settings`.

---

## 2. Domain rules

**Credit costs** by resolution (single source — backend authoritative, frontend mirrors): `1K → 5`, `2K → 15`, `4K → 40`. (Design shows "-15 Credits" for the selected 2K.) Aspect ratio does not change cost.

**Nominal dimensions** = `ResolveDimensions(resolution, aspect)`:
- base longest side: 1K=1024, 2K=2048, 4K=4096.
- aspect ratios → (w:h): `1:1`→1:1, `4:3`→4:3, `16:9`→16:9, `9:16`→9:16. Longest side gets the base; other side scaled by ratio, rounded to even int.
- e.g. 2K + 16:9 → 2048×1152; 2K + 1:1 → 2048×2048; 2K + 9:16 → 1152×2048; 1K + 4:3 → 1024×768.

**Render dimensions** (actual pixels the mock encodes; keep files small): scale nominal so longest side = `min(nominalLongest, 1024)`, preserve aspect, even ints.

**Signup bonus**: new users get **1250** credits. `signup_bonus` transaction.

**Generation lifecycle** (async): `pending → processing → completed | failed`.
- On create: deduct cost immediately (transaction), status `pending`, enqueue.
- Worker: `processing` → call provider → save PNG → `completed` (+image_url,+completed_at). On provider error → `failed` (+error) and **refund** cost (`refund` transaction).
- Cost is recorded on the Generation row.

**Roles**: `user`, `admin`. Admin endpoints require `role=admin`.

**public_id**: human id like `USR-XXXXX` (5 hex upper). Unique. Generated at user create.

**avatar_seed**: random string; frontend renders avatar via `https://api.dicebear.com/9.x/bottts-neutral/svg?seed=<avatar_seed>` (deterministic, no upload needed).

---

## 3. Backend — exact package APIs (BACKEND agent: expose these symbols verbatim)

Module path: `syntheticvision`. `go 1.24`. Deps (require in go.mod; main thread runs `go mod tidy`):
`github.com/go-chi/chi/v5`, `github.com/go-chi/cors`, `gorm.io/gorm`, `github.com/glebarez/sqlite`, `github.com/golang-jwt/jwt/v5`, `github.com/google/uuid`, `golang.org/x/crypto/bcrypt` (via `golang.org/x/crypto`).

### internal/config — `package config`
```go
type Config struct {
    Port            string // PORT, default "8080"
    DataDir         string // DATA_DIR, default "./data"
    JWTSecret       string // JWT_SECRET, default "dev-insecure-change-me"
    JWTTTLHours     int    // JWT_TTL_HOURS, default 168
    ImageProvider   string // IMAGE_PROVIDER, default "mock"  ("mock"|"openai")
    OpenAIBaseURL   string // OPENAI_BASE_URL
    OpenAIAPIKey    string // OPENAI_API_KEY
    ImageModel      string // IMAGE_MODEL, default "dall-e-3"
    MockDelayMs     int    // MOCK_DELAY_MS, default 2500
    AdminEmail      string // ADMIN_EMAIL, default "admin@synthetic.ai"
    AdminPassword   string // ADMIN_PASSWORD, default "synthavision"
    Workers         int    // GEN_WORKERS, default 4
    SeedDemoUsers   bool   // SEED_DEMO_USERS, default true
}
func Load() Config        // read env with defaults
func (c Config) ImagesDir() string      // filepath.Join(DataDir, "images")
func (c Config) ReferencesDir() string  // filepath.Join(DataDir, "references")
func (c Config) DBPath() string         // filepath.Join(DataDir, "synthetic-vision.db")
```

### internal/models — `package models`
```go
type User struct {
    ID           string    `gorm:"primaryKey;size:36" json:"id"`
    PublicID     string    `gorm:"uniqueIndex;size:16" json:"public_id"`
    Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
    Email        string    `gorm:"uniqueIndex;size:160" json:"email"`
    PasswordHash string    `gorm:"size:120" json:"-"`
    Role         string    `gorm:"size:16;default:user" json:"role"`
    Plan         string    `gorm:"size:16;default:free" json:"plan"`
    Credits      int       `gorm:"default:0" json:"credits"`
    AvatarSeed   string    `gorm:"size:64" json:"avatar_seed"`
    IsDemo       bool      `gorm:"default:false" json:"-"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
type Generation struct {
    ID             string     `gorm:"primaryKey;size:36" json:"id"`
    UserID         string     `gorm:"index;size:36" json:"-"`
    Mode           string     `gorm:"size:16;default:text" json:"mode"` // text|image|edit
    Prompt         string     `gorm:"type:text" json:"prompt"`
    NegativePrompt string     `gorm:"type:text" json:"negative_prompt"`
    Resolution     string     `gorm:"size:8" json:"resolution"`
    AspectRatio    string     `gorm:"size:8" json:"aspect_ratio"`
    Style          string     `gorm:"size:32" json:"style"`
    Width          int        `json:"width"`
    Height         int        `json:"height"`
    Seed           int64      `json:"seed"`
    Status         string     `gorm:"size:16;index" json:"status"`
    Cost           int        `json:"cost"`
    ImageURL       string     `gorm:"size:255" json:"image_url"`   // e.g. "/images/<id>.png"
    SourceImagePath string    `gorm:"size:512" json:"-"`
    MaskImagePath   string    `gorm:"size:512" json:"-"`
    HasSourceImage  bool      `gorm:"default:false" json:"has_source_image"`
    HasMaskImage    bool      `gorm:"default:false" json:"has_mask_image"`
    Error          string     `gorm:"type:text" json:"error"`
    CreatedAt      time.Time  `json:"created_at"`
    CompletedAt    *time.Time `json:"completed_at"`
}
type CreditTransaction struct {
    ID        string    `gorm:"primaryKey;size:36" json:"id"`
    UserID    string    `gorm:"index;size:36" json:"user_id"`
    Amount    int       `json:"amount"`   // + credit, - debit
    Reason    string    `gorm:"size:32" json:"reason"` // generation|admin_injection|signup_bonus|refund
    AdminID   string    `gorm:"size:36" json:"admin_id"`
    CreatedAt time.Time `json:"created_at"`
}
```
The User/Generation structs ARE the JSON DTOs (json tags above). Admin user rows add computed fields — see admin handler. `PasswordHash` is `json:"-"`.

### internal/db — `package db`
```go
func Open(path string) (*gorm.DB, error)   // glebarez/sqlite, PRAGMA foreign_keys, busy_timeout; AutoMigrate(User,Generation,CreditTransaction)
```

### internal/auth — `package auth`
```go
func HashPassword(pw string) (string, error)          // bcrypt
func CheckPassword(hash, pw string) bool
type Claims struct { jwt.RegisteredClaims; UserID string `json:"uid"`; Role string `json:"role"` }
func GenerateToken(secret, userID, role string, ttl time.Duration) (string, error)
func ParseToken(secret, tokenStr string) (*Claims, error)
```

### internal/middleware — `package middleware`
```go
type ctxKey string
const UserIDKey ctxKey = "uid"; const RoleKey ctxKey = "role"
func RequireAuth(secret string) func(http.Handler) http.Handler  // 401 if no/invalid bearer; sets UserIDKey,RoleKey in ctx
func RequireAdmin(next http.Handler) http.Handler                 // 403 if RoleKey != "admin"
func UserID(r *http.Request) string
func Role(r *http.Request) string
```

### internal/provider — `package provider`
```go
type InputImage struct { Data []byte; MimeType string; Filename string }
type GenerateRequest struct { Mode, Prompt, NegativePrompt, Style string; Width, Height int; Seed int64; SourceImage, MaskImage *InputImage }
type GenerateResult struct { Image []byte; MimeType string }   // PNG bytes
type Provider interface { Name() string; Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) }
func New(cfg config.Config) Provider   // returns Mock or OpenAI by cfg.ImageProvider (default mock)
// mock.go: MockProvider — deterministic on-brand abstract art (see §6); honors cfg.MockDelayMs; source/mask bytes affect the render seed and overlay.
// openai.go: OpenAIProvider — text mode POSTs JSON to {BaseURL}/v1/images/generations; image/edit modes POST multipart to {BaseURL}/v1/images/edits with image[] and optional mask. Decode b64_json → PNG; fall back to url field (download). If decode is non-PNG, still store raw bytes with detected mime.
```

### internal/service — `package service`
```go
type Service struct { /* db *gorm.DB; prov provider.Provider; cfg config.Config; queue chan string; ... */ }
func New(db *gorm.DB, prov provider.Provider, cfg config.Config) *Service
func (s *Service) Start()   // launch cfg.Workers goroutines consuming queue
func (s *Service) Stop()
// CreateGeneration: validates resolution∈{1K,2K,4K}, aspect∈{1:1,4:3,16:9,9:16}; computes cost+dims+seed;
// in a db transaction: load user FOR UPDATE-ish, ensure Credits>=cost, decrement, write CreditTransaction(reason "generation", -cost),
// create Generation(status "pending"); then enqueue id; returns the Generation (or ErrInsufficientCredits).
func (s *Service) CreateGeneration(userID string, in CreateGenInput) ([]*models.Generation, error)
type UploadedImage struct { Data []byte; MimeType, Extension, Filename string }
type CreateGenInput struct { Mode, Prompt, NegativePrompt, Style, Resolution, AspectRatio string; Count int; SourceImage, MaskImage *UploadedImage }
var ErrInsufficientCredits = errors.New("insufficient credits")
var ErrInvalidParam = errors.New("invalid parameter")
func CostFor(resolution string) int        // 1K→5,2K→15,4K→40; unknown→ -1
func ResolveDimensions(resolution, aspect string) (w, h int)
// AdminInjectCredits: add amount to target user by public_id, write CreditTransaction(reason "admin_injection", +amount, AdminID).
func (s *Service) AdminInjectCredits(adminID, targetPublicID string, amount int) (*models.User, error)
```
Worker writes PNG to `cfg.ImagesDir()/<genID>.png`, sets ImageURL `/images/<genID>.png`.

### internal/handlers — `package handlers`
```go
type Handler struct { DB *gorm.DB; Svc *service.Service; Cfg config.Config }
func New(db *gorm.DB, svc *service.Service, cfg config.Config) *Handler
// helpers (unexported ok): writeJSON(w,status,v), writeError(w,status,msg)
// auth.go
func (h *Handler) Register(w,r)   // POST /api/auth/register {username,email,password} → 201 {token,user}
func (h *Handler) Login(w,r)      // POST /api/auth/login {email,password} → {token,user}
func (h *Handler) Me(w,r)         // GET  /api/auth/me → {user}
// generation.go
func (h *Handler) CreateGeneration(w,r) // POST /api/generations JSON or multipart → 202 generation
func (h *Handler) ListGenerations(w,r)  // GET  /api/generations?status=&limit= → {generations:[...]}
func (h *Handler) GetGeneration(w,r)    // GET  /api/generations/{id} (owner) → generation
func (h *Handler) DeleteGeneration(w,r) // DELETE /api/generations/{id} (owner) → {ok:true}; remove file
func (h *Handler) GenerationCost(w,r)   // GET  /api/generations/cost?resolution= → {cost}
// profile.go
func (h *Handler) Stats(w,r)            // GET /api/me/stats → {user, total_generations, completed_generations, credit_balance}
// analytics.go
func (h *Handler) Analytics(w,r)        // GET /api/me/analytics → {user, summary, status_distribution, resolution_distribution, aspect_ratio_distribution, credit_breakdown, recent_generations}
// admin.go
func (h *Handler) AdminUsers(w,r)       // GET /api/admin/users?search=&page=&page_size= → {users:[...],total,page,page_size}
func (h *Handler) AdminInjectCredits(w,r) // POST /api/admin/credits {target_public_id, amount} → {user}
func (h *Handler) AdminCluster(w,r)     // GET /api/admin/cluster → {status:"Operational",load_percent:<int 60-99>}
```
Admin user row JSON: `{public_id,username,email,credits,role,plan,is_demo? omit,last_activity_at(=updated_at),initials,created_at}`. `initials` = first 2 of username upper. `total` = total matching count for pagination footer ("Showing X-Y of N").

### internal/server — `package server`
```go
func Router(h *handlers.Handler, cfg config.Config) http.Handler
// chi router. middleware: RequestID, RealIP, Recoverer, Logger, cors.Handler (allow all in dev).
// mount:
//   /api/auth/register POST, /api/auth/login POST  (public)
//   group with middleware.RequireAuth(cfg.JWTSecret):
//     GET /api/auth/me; GET /api/me/stats; GET /api/me/analytics
//     POST /api/generations; GET /api/generations; GET /api/generations/cost; GET /api/generations/{id}; DELETE /api/generations/{id}
//     group RequireAdmin: GET /api/admin/users; POST /api/admin/credits; GET /api/admin/cluster
//   /images/* → http.FileServer over cfg.ImagesDir()
//   /*  → SPA: serve embedded web.Dist (sub "dist"); for unknown non-/api, non-/images paths return dist/index.html (history fallback). If a static asset exists, serve it.
```

### web — `package web`
```go
//go:embed all:dist
var Dist embed.FS
func DistFS() (fs.FS, error)  // fs.Sub(Dist, "dist")
```
`web/dist/index.html` placeholder: minimal `<!doctype html><title>Synthetic Vision</title>` so `go build` works pre-frontend-build. Docker overwrites dist with the real vite output.

### cmd/server/main.go
Wire: `cfg=config.Load()`, mkdir DataDir+ImagesDir, `gdb=db.Open(cfg.DBPath())`, seed admin (+ demo users if cfg.SeedDemoUsers and none exist), `prov=provider.New(cfg)`, `svc=service.New(...)`, `svc.Start()`, `h=handlers.New(...)`, `r=server.Router(h,cfg)`, `http.ListenAndServe(":"+cfg.Port, r)`. Graceful shutdown on SIGINT/SIGTERM → svc.Stop(). Seeding: admin{username:"operator",email:cfg.AdminEmail,role:"admin",credits:1250,plan:"pro"}. Demo users (IsDemo true) matching design: `j.smith_ai/john.smith@example.com/4500`, `elara.design/elara@studio.io/120`, `m.k_research/mk@university.edu/0`, each role user, plan free, password "demo-pass-1234", public_ids `USR-8829A`,`USR-9102B`,`USR-1044C` (fixed for these demo rows). Log admin creds once on first seed.

---

## 4. HTTP API contract (frontend ↔ backend) — JSON, prefix `/api`, bearer auth

Auth token: `Authorization: Bearer <jwt>`. Errors: `{ "error": "<message>" }` with proper status (400/401/403/404/409/422/500).

| Method | Path | Body | Resp 2xx |
|---|---|---|---|
| POST | /api/auth/register | `{username,email,password}` | 201 `{token, user}` (409 if email/username taken; 422 weak) |
| POST | /api/auth/login | `{email,password}` | 200 `{token, user}` (401 bad creds) |
| GET | /api/auth/me | — | `{user}` |
| GET | /api/me/stats | — | `{user,total_generations,completed_generations,credit_balance}` |
| GET | /api/me/analytics | — | `{user,summary,status_distribution,resolution_distribution,aspect_ratio_distribution,credit_breakdown,recent_generations}` |
| GET | /api/generations/cost | `?resolution=2K` | `{cost: 15}` |
| POST | /api/generations | JSON `{mode:"text"?,prompt,negative_prompt?,style?,resolution,aspect_ratio}` or multipart `{mode:"image"|"edit",prompt,negative_prompt?,style?,resolution,aspect_ratio,source_image,mask_image?}` | 202 `generation` (402 `{error:"insufficient credits"}` when broke) |
| GET | /api/generations | `?status=&limit=` | `{generations:[generation,...]}` |
| GET | /api/generations/{id} | — | `generation` (404/403) |
| DELETE | /api/generations/{id} | — | `{ok:true}` |
| GET | /api/admin/users | `?search=&page=1&page_size=10` | `{users:[adminUser],total,page,page_size}` |
| POST | /api/admin/credits | `{target_public_id,amount}` | `{user}` |
| GET | /api/admin/cluster | — | `{status,load_percent}` |

`user` = User JSON (§3). `generation` = Generation JSON (§3). Insufficient credits → HTTP **402** with `{"error":"insufficient credits"}`.

---

## 5. Frontend — exact contracts (FE-CORE implements; FE-VIEWS consumes)

### src/types.ts  (export these EXACT interfaces)
```ts
export interface User { id:string; public_id:string; username:string; email:string; role:'user'|'admin'; plan:string; credits:number; avatar_seed:string; created_at:string }
export type GenStatus = 'pending'|'processing'|'completed'|'failed'
export type GenerationMode = 'text'|'image'|'edit'
export type Resolution = '1K'|'2K'|'4K'
export type AspectRatio = '1:1'|'4:3'|'16:9'|'9:16'
export interface Generation { id:string; mode:GenerationMode; prompt:string; negative_prompt:string; resolution:Resolution; aspect_ratio:AspectRatio; style:string; width:number; height:number; seed:number; status:GenStatus; cost:number; image_url:string; has_source_image:boolean; has_mask_image:boolean; error:string; created_at:string; completed_at:string|null }
export interface AdminUser { public_id:string; username:string; email:string; credits:number; role:string; plan:string; last_activity_at:string; initials:string; created_at:string }
export interface Stats { user:User; total_generations:number; completed_generations:number; credit_balance:number }
export interface CreateGenInput { mode?:GenerationMode; prompt:string; negative_prompt?:string; style?:string; resolution:Resolution; aspect_ratio:AspectRatio; count?:number; source_image?:File; mask_image?:Blob|File }

export interface AnalyticsDistributionItem { label:string; count:number; percentage:number }
export interface AnalyticsSummary { total_generations:number; completed_generations:number; failed_generations:number; pending_generations:number; processing_generations:number; success_rate:number; credits_spent:number; credits_refunded:number; credit_balance:number }
export interface AnalyticsCreditBreakdown { generation_debits:number; refunds:number; admin_topups:number; signup_bonus:number }
export interface AnalyticsRecentGeneration { id:string; mode:GenerationMode; prompt:string; status:GenStatus; resolution:Resolution; aspect_ratio:AspectRatio; cost:number; error:string; created_at:string; completed_at:string|null }
export interface AnalyticsResponse { user:User; summary:AnalyticsSummary; status_distribution:AnalyticsDistributionItem[]; resolution_distribution:AnalyticsDistributionItem[]; aspect_ratio_distribution:AnalyticsDistributionItem[]; credit_breakdown:AnalyticsCreditBreakdown; recent_generations:AnalyticsRecentGeneration[] }

export type PresetCategory = 'photoreal'|'illustration'|'abstract'|'product'|'retro'|'portrait'
export interface Preset { id:string; title:string; description:string; prompt_seed:string; style:string; suggested_resolution:Resolution; suggested_aspect_ratio:AspectRatio; tags:PresetCategory[]; estimated_cost:number; preview:string }
```

### src/api/client.ts  (axios instance, default export `api` + named helpers)
```ts
// axios.create({ baseURL:'/api' }); request interceptor adds Bearer from localStorage 'sv_token';
// response interceptor: on 401 -> clear token + redirect '/login'.
export const api = /* AxiosInstance */
export const AuthAPI = {
  register(b:{username:string;email:string;password:string}): Promise<{token:string;user:User}>
  login(b:{email:string;password:string}): Promise<{token:string;user:User}>
  me(): Promise<{user:User}>
  stats(): Promise<Stats>
}
export const GenAPI = {
  cost(resolution:Resolution): Promise<number>          // unwraps {cost}
  create(input:CreateGenInput): Promise<Generation[]>
  list(params?:{status?:GenStatus;limit?:number}): Promise<Generation[]>  // unwraps {generations}
  get(id:string): Promise<Generation>
  remove(id:string): Promise<void>
}
export const AnalyticsAPI = {
  meAnalytics(): Promise<AnalyticsResponse>
}
export const AdminAPI = {
  users(p:{search?:string;page?:number;page_size?:number}): Promise<{users:AdminUser[];total:number;page:number;page_size:number}>
  inject(b:{target_public_id:string;amount:number}): Promise<{user:User}>
  cluster(): Promise<{status:string;load_percent:number}>
}
```

### src/stores/auth.ts  (Pinia, id 'auth')
```ts
// state: token:string|null (init from localStorage 'sv_token'), user:User|null
// getters: isAuthenticated, isAdmin
// actions: login(email,password), register(username,email,password), fetchMe(), logout(), setCredits(n), refresh()  // refresh = fetchMe; login/register persist token to localStorage 'sv_token' and set user
```

### src/stores/generations.ts  (Pinia, id 'generations')
```ts
// state: items:Generation[], active:Generation|null (the one being synthesized), loading:boolean
// actions:
//   fetchAll(params?), 
//   create(input):starts a generation, sets active, then polls get(id) every 1500ms until status completed|failed, updating active and items; refreshes auth credits after start AND on completion(refund),
//   remove(id), 
//   pollActive() (internal)
// getters: completed (items filter status==='completed')
```

### src/lib/format.ts
```ts
export function relativeTime(iso:string): string   // "2 mins ago","5 hours ago","1 day ago"
export function avatarUrl(seed:string): string     // dicebear bottts-neutral svg
export const COSTS: Record<Resolution,number>      // {'1K':5,'2K':15,'4K':40}
export const RESOLUTIONS: Resolution[]             // ['1K','2K','4K']
export const ASPECTS: {id:AspectRatio;label:string;w:number;h:number}[] // for the ratio grid icon boxes
```

### router/index.ts
Routes (createWebHistory):
- `/login` → Login (public; if already authed redirect `/`)
- `/` → Dashboard (auth) — inside AppShell
- `/gallery` → Gallery (auth)
- `/admin` → Admin (auth + admin; else redirect `/`)
- `/marketplace` → Marketplace (auth)
- `/analytics` → Analytics (auth)
- `/settings` → Settings (auth)
- `/support` → Support (auth)
beforeEach guard: if route needs auth and `!auth.isAuthenticated` → `/login`; if needs admin and `!isAdmin` → `/`. On first load with token but no user, call `fetchMe()`.
AppShell wraps authed views (sidebar + topbar + `<router-view>` in content). Login renders standalone (no shell).

### Components
- **AppShell.vue**: flex layout. Fixed `Sidebar` (w-72) + `TopBar` (h-16) + scrollable `<main>` with `<slot/>` or `<router-view/>`. Matches `_design/dashboard.html` shell. Mobile: sidebar hidden (`hidden md:flex`).
- **Sidebar.vue**: brand block (auto_awesome in primary-container rounded, "Synthetic Vision" / "V3.5 Engine"), nav links (router-link active = `bg-primary-container text-on-primary-container font-bold`), Admin link only if `auth.isAdmin`, footer Upgrade → `/settings?section=billing`, Support → `/support`, Settings → `/settings`, Logout (calls auth.logout → /login). Use exact classes from `_design`.
- **TopBar.vue**: right side — credits pill (`{{auth.user?.credits}} Credits`, secondary pulse dot), notifications icon with recent generation job dropdown, avatar (avatarUrl), **Generate** button (router push `/`). Use design classes.
- **ImageCard.vue** (FE-VIEWS): props `gen:Generation`. aspect `aspect-[4/5]`, `glass-panel glow-hover`, img `image_url`, hover overlay with `WxH • relativeTime`, truncated prompt, Download + Delete buttons (emit `delete`). Exactly per `_design/gallery.html` card.

### Views (FE-VIEWS) — match `_design/*.html` fidelity
- **Login.vue**: standalone. Atmospheric bg (glows + radial). Glass card. Toggle **Login ⇄ Register** ("Initialize Session" / "Request New Allocation"). Login form: email + access key. Register adds username. On submit → auth store → on success `router.push('/')`. Show error text on failure. Use `_design/login.html` markup.
- **Dashboard.vue**: 2-col workspace. Left `aside` (320px) glass "Generation Parameters": Resolution segmented (1K/2K/4K), Aspect ratio 2×2 grid (icon boxes per design), Energy Cost footer = `COSTS[resolution]` (live). Right: prompt glass textarea + toolbar (style chip = "Cinematic") + Generate button; below, Canvas glass panel showing: idle placeholder (radial dot bg) / synthesizing state (spinner ring + "Synthesizing Vision" + progress bar, from `gen.active`) / completed image. Generate calls `gen.create(input)`; disable while active pending/processing; if 402 show "Insufficient credits" toast/inline. After completion the image shows in canvas; credits pill updates.
- **Gallery.vue**: Profile header (avatar, username, plan "Pro/Free Member", Total Generations stat, Credit Balance stat from `/me/stats`). "Recent Output" grid of `ImageCard` over `gen.completed`. Filters by prompt text, generation mode, resolution, and aspect ratio. Delete removes via store. Empty state when none or when filters match nothing.
- **Admin.vue**: page title "User Directory". 12-col bento: left (col-span-8) glass table of `AdminAPI.users` (avatar initials chip, username/public_id, email, credits pill colored by amount: 0→error, low→secondary, high→primary, Recharge button prefills the form) plus backend-backed search over username/email/public_id. Footer "Showing 1-N of total" + pager. Right (col-span-4): "Manual Credit Injection" form (target_public_id + amount → AdminAPI.inject → refresh table + toast), "Compute Cluster" status card from AdminAPI.cluster. Per `_design/admin.html`.
- **Marketplace.vue**: static preset browser with category filtering, recommendation cards (title/description/tags/metadata), and preset application flow to Dashboard via query handoff.
- **Analytics.vue**: personal analytics dashboard using `/api/me/analytics`, including summary cards, status/resolution/aspect distributions, and recent generation activity; handles loading/error/empty states without placeholders.
- **Settings.vue**: local workspace preferences persisted in `localStorage` for default Dashboard mode/resolution/aspect/style, compact Gallery layout, and billing/upgrade guidance.
- **Support.vue**: authenticated help surface with generation/editing/credit troubleshooting and links back to Dashboard/Analytics.

### main.ts
createApp(App) + pinia + router; import './style.css'. App.vue: `<router-view/>` only (shell handled per-route or App decides shell vs login by route meta). Simplest: AppShell used inside each authed view OR App.vue checks `route.meta.shell`. **Decision: App.vue renders `<RouterView/>`; authed views import and wrap with `<AppShell>`** — NO, to avoid repetition: App.vue checks `route.meta.public`; if public → `<RouterView/>`, else `<AppShell><RouterView/></AppShell>`. FE-CORE implements this in App.vue; set `meta.public:true` on /login.

### vite.config.ts
Vue plugin; `server.proxy['/api']` → `http://localhost:8080`; `server.proxy['/images']` → `http://localhost:8080`. `build.outDir` default `dist`. base '/'.

### package.json
deps: vue, vue-router, pinia, axios. devDeps: vite, @vitejs/plugin-vue, typescript, vue-tsc, tailwindcss@3, postcss, autoprefixer, @types/node. scripts: `dev`,`build` (`vue-tsc -b && vite build` — or just `vite build` to avoid TS strictness blocking; use `vite build` for build script, keep `typecheck` separate), `preview`. `.npmrc`: `registry=https://registry.npmmirror.com`.

---

## 6. Mock image generator (BACKEND, mock.go) — make it look good

Deterministic from `Seed` + hash(Prompt). Steps:
1. RNG seeded by `Seed ^ fnv(Prompt)`.
2. Canvas = render dims. Base: vertical/diagonal linear gradient between two palette colors chosen by RNG from: `#7c3aed,#d2bbff,#00a2e6,#89ceff,#0b1326,#171f33,#a15100,#ffb784`.
3. Add 3–6 soft radial glow blobs (random center/radius, additive-ish blend toward primary/secondary/tertiary). 
4. Optional faint geometric overlay: a few translucent thin lines or concentric rings.
5. Subtle per-pixel noise (±small).
6. Vignette darkening toward edges (toward `#060e20`).
Encode PNG (`image/png`). Honor `cfg.MockDelayMs` sleep (split so context cancel works). Pure stdlib `image`,`image/color`,`image/draw`,`image/png`,`math`,`hash/fnv`.

---

## 7. Docker / ops (OPS agent)

**Dockerfile** (`<root>/Dockerfile`, context `<root>`):
```
# stage 1: frontend
FROM node:22-alpine AS fe
WORKDIR /fe
COPY frontend/package.json frontend/.npmrc ./
RUN npm install
COPY frontend/ ./
RUN npm run build         # → /fe/dist
# stage 2: backend build (embeds dist)
FROM golang:1.24-alpine AS be
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY backend/go.mod ./
RUN go mod download || true
COPY backend/ ./
COPY --from=fe /fe/dist ./web/dist
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/server ./cmd/server
# stage 3: runtime
FROM alpine:3.20
RUN adduser -D -u 1000 app && mkdir -p /data && chown app /data
COPY --from=be /out/server /usr/local/bin/server
USER app
ENV DATA_DIR=/data PORT=8080
EXPOSE 8080
VOLUME /data
CMD ["server"]
```
**docker-compose.yml**: service `app`, build `.`, ports `8080:8080`, `env_file: .env`, volume `./data:/data`, restart unless-stopped.
**.env.example**: all config keys from §3 with safe defaults; comment that JWT_SECRET must be stable; OPENAI_* blank (mock default).
**Makefile**: `dev-backend` (cd backend && go run ./cmd/server), `dev-frontend` (cd frontend && npm run dev), `build`, `up` (docker compose up --build), `down`. Note Windows: also document raw commands in README.
**.gitignore**: `data/`, `node_modules/`, `dist/` (except backend/web/dist placeholder? — ignore `frontend/dist`, keep `backend/web/dist/index.html`), `*.db`, `.env`.
**README.md**: what it is, screenshots note, quickstart (Docker: `docker compose up --build` → http://localhost:8080; local dev: backend `go run`, frontend `npm run dev` at :5173 proxying), default admin creds (admin@synthetic.ai / synthavision), env reference table, switching to real OpenAI-compatible provider, architecture diagram (text), credit costs table.

---

## 8. Acceptance / coherence checklist (build+verify step)
- `cd backend && go build ./...` succeeds (after `go mod tidy`).
- `cd frontend && npm install && npm run build` succeeds; outputs `dist`.
- Endpoints match §4 exactly; frontend `api/client.ts` paths match.
- Login → register → land on Dashboard → generate (mock) → image appears within ~MockDelay → Gallery shows it → credits decreased by cost → Admin (as admin) lists users + inject works.
- No symbol drift: every cross-package call uses the signatures in §3; every FE-VIEWS call uses §5 store/api signatures.
