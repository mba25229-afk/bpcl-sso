# Chunk 2 — Backend Foundation
## go.mod, config, db pool, models

---

## Context
Database is running with seed data (Chunk 1 complete).
This chunk builds the Go skeleton: module init, config loader, DB connection, and all structs.
No HTTP endpoints yet. Just the foundation that everything else depends on.
Read AGENTS.md before starting.

## Prerequisites
- Chunk 1 complete and verified
- Go 1.22+ installed (`go version`)

---

## Task 1 — Initialize Go Module

```bash
mkdir bpcl-portal-api && cd bpcl-portal-api
go mod init github.com/bpcl/portal-api
```

Add all dependencies at once:
```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/golang-migrate/migrate/v4@latest
go get github.com/golang-migrate/migrate/v4/database/postgres@latest
go get github.com/golang-migrate/migrate/v4/source/file@latest
go get github.com/go-playground/validator/v10@latest
go get github.com/golang-jwt/jwt/v5@latest
go get github.com/qax-os/excelize/v2@latest
go get github.com/spf13/viper@latest
go get github.com/prometheus/client_golang@latest
go get github.com/stretchr/testify@latest
go get github.com/google/uuid@latest
go get golang.org/x/crypto@latest
go get golang.org/x/time@latest
```

**Verify:** `go mod tidy` runs without errors, `go.sum` is generated.

---

## Task 2 — Config Loader

Create `internal/config/config.go`:
- Load from `.env` file using viper
- Validate required vars: BPCL_DB_URL, BPCL_JWT_SECRET, BPCL_PORT
- Fail fast with clear error if any required var missing
- Expose typed struct: `Config{DBUrl, JWTSecret, JWTExpiry, Port, UploadDir, CORSOrigins, MaxUploadMB, LogLevel, Env, RateLimitRPM}`

**Verify:** `go build ./internal/config/...` succeeds

---

## Task 3 — Database Pool

Create `internal/repository/db.go`:
- Connect using pgxpool with config from Config struct
- Ping on startup, return error if unreachable
- Pool settings: MaxConns=20, MinConns=2, MaxConnIdleTime=30m
- Expose `Connect(ctx, dbURL) (*pgxpool.Pool, error)` and `Close(pool)`

**Verify:** `go build ./internal/repository/...` succeeds

---

## Task 4 — All Models

Create these files (pure structs, no methods, no DB calls):

`internal/model/user.go` — User struct, UserRole constants (ro_manager, territory_manager, admin)
`internal/model/outlet.go` — RetailOutlet struct matching retail_outlets table
`internal/model/product.go` — Product struct, category constants
`internal/model/performance.go` — PerformanceRecord, PerformanceSummary, KPIData structs
`internal/model/target.go` — Target struct, TargetsResponse (fuel map + non_fuel map)
`internal/model/upload.go` — UploadedFile struct
`internal/model/competition.go` — CompetitionPeriod, CompetitionScore, DealerAuditScore, CompetitionBonus structs
`internal/model/errors.go` — Sentinel errors: ErrNotFound, ErrForbidden, ErrUnauthorized, ErrBonusRemarksRequired

**Verify:** `go build ./internal/model/...` succeeds

---

## Done When
- `go build ./...` succeeds (will fail on missing handler/service files — that's OK)
- `go build ./internal/config/... ./internal/repository/... ./internal/model/...` all pass

## Next Chunk
→ `docs/exec-plans/chunk-3-repository-layer.md`
