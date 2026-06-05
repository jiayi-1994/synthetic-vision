# Synthetic Vision

> An AI **image generation and editing SaaS**, credit-based, wrapped in a dark glassmorphism UI
> ("Midnight Spectrum"). A single self-contained Go binary serves both the REST
> API and the embedded Vue 3 single-page app. No external database or services
> required — it ships with pure-Go SQLite and a built-in **mock image provider**,
> so it runs end-to-end out of the box.

- **Backend:** Go 1.24 — chi router, GORM, pure-Go SQLite (`glebarez/sqlite`),
  JWT auth, a background generation worker pool.
- **Frontend:** Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS 3 + axios,
  built and **embedded** into the Go binary via `go:embed`.
- **Image generation/editing:** a deterministic, on-brand **mock** generator by
  default; switchable to OpenAI-compatible `/v1/images/generations` and
  `/v1/images/edits` endpoints.

---

## The six routes

The UI is built to match the design references in [`_design/`](./_design)
(`*.html` Tailwind markup + `*.png` screenshots).

| Screen | Route | What it does |
|---|---|---|
| **Login / Register** | `/login` | Atmospheric glass card. Toggles between *Initialize Session* (login) and *Request New Allocation* (register). New accounts get a **1,250-credit** signup bonus. |
| **Dashboard** | `/` | The generation/editing workspace. Left panel: **Generation Parameters** plus mode switch (**文生图 / 图生图 / 局部修图**), resolution, aspect, and live **Energy Cost**. Right: prompt editor, reference image upload, lightweight mask brush for local edits, and the **Canvas** that shows the source preview, progress state, then the finished image. |
| **Gallery** | `/gallery` | Profile header (avatar, plan, total generations, credit balance from `/api/me/stats`) plus a **Recent Output** grid of your completed images, with download and delete. |
| **Admin** | `/admin` | *(admins only)* **User Directory** table with credit pills, a **Manual Credit Injection** form (top up any user by their `public_id`), and a **Compute Cluster** status card. |
| **Marketplace** | `/marketplace` | Curated preset browser with reusable templates, quick **Apply** to prefill Dashboard settings, and no backend commerce dependencies. |
| **Analytics** | `/analytics` | Personal dashboard aggregating generation counts, success rate, resolution/aspect distribution, refunds/spend trends, and recent activity. |

> All six routes now have functional front-end surfaces; `/marketplace` is a curated static preset catalog and `/analytics` shows personal usage metrics from `/api/me/analytics`.

---

## Quickstart

### Option A — Docker (recommended)

Builds the frontend, embeds it into the Go binary, and runs everything in one
container. Data persists in `./data`.

```bash
cp .env.example .env          # adjust if you like (defaults work as-is)
docker compose up --build
```

Then open **http://localhost:8080**.

Sign in with the default admin (or register a fresh account):

```
email:    admin@synthetic.ai
password: synthavision
```

> Change `ADMIN_PASSWORD` (and `JWT_SECRET`) in `.env` before exposing this
> anywhere real.

### Option B — Local development (hot reload)

Run the backend and the Vite dev server separately. The frontend dev server on
**:5173** proxies `/api` and `/images` to the backend on **:8080**.

**Backend** (terminal 1):

```bash
cd backend
go run ./cmd/server          # serves the API on http://localhost:8080
```

**Frontend** (terminal 2):

```bash
cd frontend
npm install
npm run dev                  # serves the SPA on http://localhost:5173
```

Open **http://localhost:5173** during development. (The backend on `:8080`
also serves the *built* SPA from `backend/web/dist` once you run a production
build.)

### Makefile shortcuts (Unix-like shells)

On Linux/macOS, WSL, or Git Bash you can use:

```bash
make dev-backend     # cd backend && go run ./cmd/server
make dev-frontend    # cd frontend && npm run dev
make build           # build frontend, embed into backend, build the binary
make up              # docker compose up --build
make down            # docker compose down
```

> **Windows note:** the `Makefile` targets are a convenience and assume a
> Unix-like shell. There is no `make` requirement — every target maps to the
> raw commands shown above (the `cd ... && ...` lines), which you can run
> directly in **PowerShell**. For example:
>
> ```powershell
> cd backend ; go run ./cmd/server          # backend
> cd frontend ; npm install ; npm run dev   # frontend
> docker compose up --build                 # full stack
> ```

---

## Configuration (environment variables)

All configuration is via environment variables (see [`.env.example`](./.env.example)).
Sensible defaults make the app run with zero configuration.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port. |
| `DATA_DIR` | `./data` | Directory for the SQLite database, generated images, and private source/mask references. In Docker this is the `/data` volume. |
| `JWT_SECRET` | `dev-insecure-change-me` | Signing secret for session tokens. **Keep this stable across restarts** — changing it invalidates every existing session and forces all users to re-login. Use a strong value in production (`openssl rand -hex 32`). |
| `JWT_TTL_HOURS` | `168` | Session token lifetime, in hours (default 7 days). |
| `IMAGE_PROVIDER` | `mock` | Image backend: `mock` or `openai`. |
| `OPENAI_BASE_URL` | *(blank)* | Base URL of an OpenAI-compatible API (used only when `IMAGE_PROVIDER=openai`). |
| `OPENAI_API_KEY` | *(blank)* | Bearer key for the OpenAI-compatible API. |
| `IMAGE_MODEL` | `dall-e-3` | Model name sent to the image API. |
| `MOCK_DELAY_MS` | `2500` | Artificial synthesis delay (ms) for the mock provider. |
| `ADMIN_EMAIL` | `admin@synthetic.ai` | Seeded admin account email. |
| `ADMIN_PASSWORD` | `synthavision` | Seeded admin account password. |
| `GEN_WORKERS` | `4` | Number of background generation worker goroutines. |
| `SEED_DEMO_USERS` | `true` | Seed demo users (matching the screenshots) on an empty database. |

---

## Switching to a real (OpenAI-compatible) image provider

By default `IMAGE_PROVIDER=mock` generates deterministic, on-brand abstract art
locally — no API key, no network. It also supports source-image and mask-backed
edit modes for end-to-end demos. To produce **real** images from OpenAI or a
compatible gateway/proxy, set these in your `.env`:

```dotenv
IMAGE_PROVIDER=openai
OPENAI_BASE_URL=https://api.openai.com      # or your compatible gateway base URL
OPENAI_API_KEY=sk-...                       # your bearer key
IMAGE_MODEL=gpt-image-1                       # accepts arbitrary WxH sizes
```

The backend POSTs text-only jobs to `{OPENAI_BASE_URL}/v1/images/generations`
with `{model, prompt, size:"WxH", n:1, response_format:"b64_json"}`. When the
Dashboard sends a reference image or mask, the backend stores the upload under
`DATA_DIR/references` and POSTs multipart form data to
`{OPENAI_BASE_URL}/v1/images/edits` with `image[]`, optional `mask`, `model`,
`prompt`, and `size`. The returned image is decoded and stored like any other
generation result.

> **Model size constraints.** `gpt-image-1` accepts arbitrary `WxH` sizes, so
> every resolution/aspect combination works. **`dall-e-3` only accepts
> `1024x1024`, `1792x1024`, and `1024x1792`** — with it, any non-square (or 2K/4K
> non-1:1) request sends an unsupported size and the API returns 400, marking
> that generation `failed`. If you must use `dall-e-3`, restrict the UI to the
> 1:1 aspect or add a size-mapping step in `internal/provider/openai.go`.

---

## Credit costs

Generation cost is charged up front and **refunded automatically** if the
provider fails. Cost depends only on resolution; aspect ratio does not change it.

| Resolution | Longest side | Credits |
|---|---|---|
| **1K** | 1024 px | **5** |
| **2K** | 2048 px | **15** |
| **4K** | 4096 px | **40** |

New accounts start with a **1,250-credit** signup bonus. Admins can top up any
user from the Admin screen (*Manual Credit Injection*).

---

## Architecture overview

```
                         ┌──────────────────────────────────────────────┐
   Browser  ───HTTP──▶   │            Go binary (single process)         │
                         │                                              │
                         │  chi router                                  │
                         │   ├─ /api/*      → JSON REST handlers         │
                         │   │     ├─ auth (register/login/me)           │
                         │   │     ├─ generations (create/list/get/del)  │
                         │   │     ├─ me/stats                           │
                         │   │     └─ admin/* (users, credits, cluster)  │
                         │   ├─ /images/*   → generated PNG file server  │
                         │   └─ /*          → embedded Vue SPA (go:embed) │
                         │                                              │
                         │  Service layer                               │
                         │   ├─ credit ledger (atomic deduct/refund)    │
                         │   └─ worker pool ──┐                          │
                         │                    ▼                          │
                         │            Provider interface                │
                         │             ├─ mock   (deterministic art)    │
                         │             └─ openai (HTTP image API)        │
                         └───────┬──────────────────────────┬───────────┘
                                 │                          │
                          GORM + SQLite              PNG files
                         (DATA_DIR/*.db)         (DATA_DIR/images/*.png)
```

- **Single deployable.** The Vue app is built by Vite and **embedded** into the
  Go binary at compile time (`backend/web/web.go` `go:embed all:dist`). One
  process serves the API, the generated images, and the SPA — with client-side
  history fallback for unknown non-`/api`, non-`/images` paths.
- **Async generation/editing.** Creating a generation deducts credits in a single DB
  transaction, records the generation as `pending`, and enqueues it. A pool of
  `GEN_WORKERS` goroutines moves it `pending → processing → completed`, writing
  the PNG to `DATA_DIR/images/<id>.png`. On provider failure it becomes `failed`
  and the cost is **refunded** as a credit transaction.
- **Credit ledger.** Every balance change (`signup_bonus`, `generation`,
  `refund`, `admin_injection`) is recorded as a `CreditTransaction`, keeping the
  user balance auditable.
- **Pluggable provider.** `IMAGE_PROVIDER` selects the implementation behind a
  small `Provider` interface, so the mock and the real OpenAI-compatible backend
  are interchangeable without touching the rest of the app.
- **Persistence.** SQLite database, generated PNGs, and private source/mask
  reference uploads all live under
  `DATA_DIR` (the `/data` volume in Docker), so all state survives restarts.

---

## Project layout

```
synthetic-vision/
├─ Dockerfile            multi-stage: frontend build → embed → go build → alpine runtime
├─ docker-compose.yml    one service (app), port 8080, ./data volume
├─ .env.example          all config keys with safe defaults
├─ Makefile              dev/build/up/down convenience targets
├─ backend/              Go API + embedded SPA (module "syntheticvision", entry ./cmd/server)
└─ frontend/             Vue 3 + TS + Vite SPA
```
