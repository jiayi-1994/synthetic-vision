# syntax=docker/dockerfile:1

# ─────────────────────────────────────────────────────────────────────────────
# Stage 1: frontend — build the Vue 3 + Vite SPA into /fe/dist
# ─────────────────────────────────────────────────────────────────────────────
FROM node:22-alpine AS fe
WORKDIR /fe
COPY frontend/package.json frontend/package-lock.json frontend/.npmrc ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ─────────────────────────────────────────────────────────────────────────────
# Stage 2: backend build — embeds the frontend dist, builds a static binary
# ─────────────────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS be
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=fe /fe/dist ./web/dist
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# ─────────────────────────────────────────────────────────────────────────────
# Stage 3: runtime — minimal alpine, non-root uid 1000, /data volume
# ─────────────────────────────────────────────────────────────────────────────
FROM alpine:3.20
# ca-certificates: the CGO_ENABLED=0 binary has no system cert fallback, so the
# OpenAI provider's HTTPS calls fail without these. tzdata: correct timestamps.
RUN adduser -D -u 1000 app && mkdir -p /data && chown app /data \
 && apk add --no-cache ca-certificates tzdata
COPY --from=be /out/server /usr/local/bin/server
USER app
ENV DATA_DIR=/data PORT=8080
EXPOSE 8080
VOLUME /data
CMD ["server"]
