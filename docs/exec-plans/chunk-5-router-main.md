# Chunk 5 — Middleware, Router, Main
## Wire everything together. First running server.

---

## Context
Services and handlers are built and tested (Chunk 4 done).
This chunk creates middleware, the router, and main.go.
After this chunk, `make run` starts a live API.

---

## Task 1 — Middleware (`internal/middleware/`)

`auth.go` — JWT validation middleware:
- Extract `Authorization: Bearer <token>` header
- Call auth.ValidateToken
- Inject userID + role + territory_code into context using typed context keys
- Return 401 if missing/invalid

`logger.go` — Request logging:
- Log: method, path, status, duration, request_id using slog
- Format: structured JSON when ENV=production, human-readable in development

`requestid.go` — UUID per request:
- Generate UUID, add to context and `X-Request-ID` response header

`cors.go` — CORS from env:
- Parse BPCL_CORS_ORIGINS (comma-separated)
- Allow GET/POST/PUT/DELETE/OPTIONS
- Allow Content-Type, Authorization headers

`ratelimit.go` — Per-user rate limiting:
- Use `golang.org/x/time/rate` limiter
- Key: userID from JWT (or IP for unauthenticated)
- BPCL_RATE_LIMIT_RPM config value
- Return 429 with Retry-After header when exceeded

---

## Task 2 — Router (`internal/router/router.go`)

Mount all routes per CONTRACTS.md:
```
POST   /api/v1/auth/login         → authHandler.Login       [public]
GET    /api/v1/auth/me            → authHandler.Me           [jwt]

GET    /api/v1/outlets/:cc        → outletHandler.Get        [jwt]
GET    /api/v1/outlets/:cc/performance → perfHandler.Get     [jwt] ?period=YYYY-MM
GET    /api/v1/outlets/:cc/analysis   → perfHandler.Analysis [jwt] ?from=&to=
GET    /api/v1/outlets/:cc/targets    → targetHandler.Get    [jwt] ?period=YYYY-MM
PUT    /api/v1/outlets/:cc/targets    → targetHandler.Set    [jwt, role: tm|admin]

GET    /api/v1/competition/leaderboard       → compHandler.Leaderboard  [jwt]
GET    /api/v1/competition/dealers/:cc/scorecard → compHandler.Scorecard [jwt]

POST   /api/v1/uploads            → uploadHandler.Create     [jwt]
GET    /api/v1/uploads            → uploadHandler.List       [jwt]
GET    /api/v1/uploads/:id        → uploadHandler.Get        [jwt]
DELETE /api/v1/uploads/:id        → uploadHandler.Delete     [jwt]

GET    /health                    → healthHandler.Health     [public]
GET    /metrics                   → promhttp.Handler()       [public or internal]
```

Middleware chain: requestid → logger → cors → ratelimit → auth (on protected routes)

---

## Task 3 — Main (`cmd/api/main.go`)

Wire order:
1. Load config (viper)
2. Set up slog structured logger
3. Connect pgxpool
4. Run pending migrations (golang-migrate, source=file, db=postgres)
5. Create all repositories
6. Create all services (inject repos)
7. Create all handlers (inject services)
8. Create router (inject handlers)
9. Create `http.Server{Addr: ":"+cfg.Port, Handler: router, ReadTimeout: 15s, WriteTimeout: 30s}`
10. Start server in goroutine
11. Wait for SIGTERM/SIGINT
12. Graceful shutdown: `srv.Shutdown(ctx 30s timeout)`, then `pool.Close()`

---

## Task 4 — Complete Makefile

```makefile
.PHONY: run build test migrate migrate-down lint docker

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v -race

migrate:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down 1

lint:
	golangci-lint run ./...

docker:
	docker-compose up -d

health:
	curl -s http://localhost:8080/health | jq .
```

---

## Task 5 — Dockerfile

Multi-stage build:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bin/api ./cmd/api

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/api .
EXPOSE 8080
CMD ["./api"]
```

---

## Verify — Full Integration Test

```bash
# 1. Start everything
make docker
make migrate

# 2. Run API
make run &

# 3. Health check
curl http://localhost:8080/health
# Expected: {"status":"ok","db":"ok","version":"1.0.0"}

# 4. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"employee_id":"EMP10001","password":"Bpcl@2026"}'
# Expected: {"token":"eyJ...","user":{"role":"admin",...}}

# 5. Get outlet (use token from step 4)
export TOKEN="<token from step 4>"
curl http://localhost:8080/api/v1/outlets/112847 \
  -H "Authorization: Bearer $TOKEN"
# Expected: {"cc_number":"112847","name":"M.L. SETHI SERVICE STATION",...}

# 6. Get performance
curl "http://localhost:8080/api/v1/outlets/112847/performance?period=2026-03" \
  -H "Authorization: Bearer $TOKEN"
# Expected: fuel[] and non_fuel[] arrays with real data

# 7. Leaderboard
curl http://localhost:8080/api/v1/competition/leaderboard \
  -H "Authorization: Bearer $TOKEN"
# Expected: M.L. SETHI rank 1 with score 56.37
```

---

## Done When
- `make run` starts without errors
- All 7 curl checks return expected responses
- `go test ./... -race` passes

## Next Chunk
→ `docs/exec-plans/chunk-6-frontend-connection.md`
