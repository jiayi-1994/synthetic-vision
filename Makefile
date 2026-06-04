# Synthetic Vision — developer convenience targets.
#
# These are conveniences for Unix-like shells (Linux/macOS, WSL, Git Bash).
# On Windows PowerShell, run the equivalent raw commands documented in README.md.

.PHONY: dev-backend dev-frontend build up down

## dev-backend: run the Go API server on :8080 (mock provider by default)
dev-backend:
	cd backend && go run ./cmd/server

## dev-frontend: run the Vite dev server on :5173 (proxies /api and /images to :8080)
dev-frontend:
	cd frontend && npm run dev

## build: build the frontend, embed it, and build the backend binary
build:
	cd frontend && npm ci && npm run build
	mkdir -p backend/web/dist
	find backend/web/dist -mindepth 1 -delete
	cp -r frontend/dist/. backend/web/dist/
	cd backend && go build -o server ./cmd/server

## up: build and start the full stack via Docker Compose
up:
	docker compose up --build

## down: stop and remove the Docker Compose stack
down:
	docker compose down
