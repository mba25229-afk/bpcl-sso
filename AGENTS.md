# AGENTS.md — BPCL SSO Portal
## Read this before doing anything. This is your map, not your manual.

---

## What This System Is

A two-part platform for BPCL Delhi retail operations:

1. **SSO Dashboard** — Sales Officers view per-outlet KPIs (fuel + non-fuel performance vs. targets)
2. **Boost and Win** — Monthly dealer competition with a 100-point scoring engine and leaderboard
3. **Crystal Subsystem** — ETL ingestion, scoring, admin portal, dealer scorecards, MTD dashboards

Backend: Go 1.22 + Chi router + pgx/v5 + PostgreSQL 16 (no ORM)
Frontend: React 18 + Vite + MUI + shadcn/ui + Tailwind v4 + Vitest (at `Bpclssoportal-main/`)
ETL: Separate Go module at `etl/`

---

## Project Structure

```
bpcl-portal-api/          # Go backend API
  cmd/api/main.go         # Entry point
  internal/
    config/               # Viper-based config (.env)
    repository/           # Raw pgx SQL (db.go, helpers.go, *_repo.go)
    service/              # Business logic layer
    handler/              # HTTP handlers (auth, outlets, performance, etc.)
    router/               # Chi router + middleware chain
    middleware/           # auth, cors, ratelimit, logger, requestid, security, timeout
    model/                # Domain structs
    scoring/              # 100-point scoring engine (formula, ranker, actuals)
    dashboard/            # Crystal Chunk 6: dashboard handler + export
    portal/               # Crystal Chunk 5: SSO portal handler
    admin/                # Crystal Chunk 4: admin portal handler
    crystal/              # Crystal subsystem
      ingest/             # Daily bulk, MAK-GE, Google rating ingestion
      health/             # Cron monitoring (ping/status)
      mtd/                # Month-to-date dealer actuals
      period/             # Competition period management
      dealer/             # Dealer repository
    parser/               # Excel file parser (excelize)
Bpclssoportal-main/       # React frontend (Vite + MUI + shadcn/ui)
etl/                      # ETL Go module (Google Sheets sync)
migrations/               # 13 migration files (000001–000013)
tests/                    # Playwright e2e + API + functional + security tests
scripts/                  # migrate.sh, seed_dev.sql, etc.
```

---

## Execution Order — Follow This. Do Not Skip Phases.

| Phase | Doc | What gets built |
|---|---|---|
| 1 | `docs/exec-plans/chunk-1-database-schema.md` | All migration files, schema only |
| 2 | `docs/exec-plans/chunk-2-backend-foundation.md` | go.mod, config, db, models |
| 3 | `docs/exec-plans/chunk-3-repository-layer.md` | All repository SQL functions |
| 4 | `docs/exec-plans/chunk-4-service-layer.md` | Business logic + HTTP handlers |
| 5 | `docs/exec-plans/chunk-5-router-main.md` | Router, middleware, main.go |
| 6 | `docs/exec-plans/chunk-6-frontend-connection.md` | CORS, env, frontend API wiring |

**Crystal subsystem** (built after core):
| Chunk | What | Key files |
|---|---|---|
| 1 | Database schema | `migrations/000010_crystal_schema.up.sql`, `000011_crystal_etl_log.up.sql` |
| 2 | Ingest API | `internal/crystal/ingest/` — daily-bulk, MAK-GE, Google rating, targets-bulk |
| 3 | Scoring engine | `internal/scoring/` — formula, ranker, engine, actuals |
| 4 | Admin portal | `internal/admin/` — dealers, periods, targets, manual scores, ETL trigger |
| 5 | SSO portal | `internal/portal/` — dealer scorecard |
| 6 | Dashboard | `internal/dashboard/` — dashboard + CSV export |
| + | Health/MTD | `internal/crystal/health/`, `internal/crystal/mtd/` |

---

## Critical Design Rules (enforced in every phase)

- Handler → Service → Repository. Never skip a layer. Never call DB from handler.
- No ORM. Raw pgx SQL only. No gorm, no ent.
- No fmt.Sprintf in SQL. Parameterized queries only ($1, $2...).
- Blank ≠ Zero in scoring. See `docs/design-docs/scoring-engine.md` Bug #2.
- Access check is the FIRST call in every outlet-scoped service function.
- Period is always stored as DATE = first day of month (2025-04 → 2025-04-01).
- JWT auth via `Authorization: Bearer <token>` header. Token contains `user_id`, `role`, `outlet_code`.
- Middleware chain (outermost → innermost): SecurityHeaders → RequestTimeout → RequestID → Logger → CORS → RateLimit → [Auth].

---

## API Endpoints

### Public
- `GET /health` — Health check (includes DB ping)
- `POST /api/v1/auth/login` — Login
- `POST /api/v1/auth/refresh` — Token refresh
- `POST /api/v1/auth/forgot-password` — Request password reset OTP
- `POST /api/v1/auth/reset-password` — Reset password with OTP

### Protected (all require JWT)
- `GET /api/v1/auth/me` — Current user info
- `GET /api/v1/outlets` — List outlets
- `GET /api/v1/outlets/{cc}` — Single outlet
- `GET /api/v1/outlets/{cc}/performance` — Dashboard KPIs
- `GET /api/v1/outlets/{cc}/analysis` — Trend analysis
- `GET /api/v1/outlets/{cc}/trend` — Performance trend chart
- `GET /api/v1/outlets/{cc}/targets` — Get targets
- `PUT /api/v1/outlets/{cc}/targets` — Set targets
- `GET /api/v1/territory/outlets` — Territory outlets
- `GET /api/v1/territory/summary` — Territory overview
- `POST /api/v1/outlets/{cc}/uploads` — Upload performance data
- `GET /api/v1/uploads` — Upload history
- `POST /api/v1/uploads/delhi-master` — Upload Delhi Master data

### Competition
- `GET /api/v1/competition/leaderboard` — Leaderboard
- `GET /api/v1/competition/{competition_id}/dealers/{cc}` — Dealer scorecard
- `GET /api/v1/competition/{id}/market-share-status` — Market share status
- `POST /api/v1/competition/{id}/recompute` — Recompute scores

### Admin
- `GET/POST/PUT /api/v1/admin/users` — User management
- `PUT /api/v1/admin/users/{id}/reset-password` — Reset user password
- `GET /api/v1/admin/dealers` — List dealers
- `PUT /api/v1/admin/dealers/{cc_code}` — Toggle dealer
- `GET/POST /api/v1/admin/competition-periods` — Period management
- `PATCH /api/v1/admin/competition-periods/{month_year}/activate` — Activate period
- `GET /api/v1/admin/targets/{month_year}` — Admin targets view
- `GET /api/v1/admin/ingest/summary` — Ingest summary
- `GET /api/v1/admin/mak-ge/{month_year}` — MAK-GE data
- `GET/POST /api/v1/admin/manual-scores` — Manual score entry
- `POST /api/v1/admin/etl/trigger` — Trigger ETL sync
- `GET /api/v1/admin/etl/status` — ETL status

### Crystal
- `POST /api/v1/ingest/daily-bulk` — Bulk daily ingestion
- `POST /api/v1/ingest/mak-ge` — MAK-GE ingestion
- `POST /api/v1/ingest/google-rating` — Google rating ingestion
- `POST /api/v1/targets/bulk` — Bulk target upload
- `POST /api/v1/scoring/compute` — Compute scores
- `GET /api/v1/portal/{cc_code}/scorecard` — Dealer scorecard
- `GET /api/v1/dashboard` — Dashboard data
- `GET /api/v1/dashboard/export` — CSV export
- `POST /api/v1/health/cron-ping` — Cron heartbeat (no auth)
- `GET /api/v1/health/cron-status` — Cron status
- `GET /api/v1/crystal/dealers` — List dealers
- `GET /api/v1/crystal/dealers/{cc_code}/mtd` — MTD actuals
- `GET /api/v1/crystal/dashboard` — MTD dashboard

---

## Domain Knowledge

| Topic | Where to read |
|---|---|
| 100-point scoring system + 5 bugs | `docs/design-docs/scoring-engine.md` |
| Delhi Master market share logic | `docs/design-docs/data-pipeline.md` |
| Crystal DB schema | `migrations/000010_crystal_schema.up.sql` |
| Scoring formula | `docs/crystal /scoring-formula.md` |
| MAK-GE delta logic | `docs/crystal /mak-ge-delta.md` |
| MS aggregation | `docs/crystal /ms-aggregation.md` |
| Crystal test strategy | `docs/crystal /test-strategy.md` |

---

## Access Control (memorize this)

| Role | Own outlet | Territory outlets | All outlets | Set targets |
|---|---|---|---|---|
| ro_manager | ✅ | ❌ | ❌ | ❌ |
| territory_manager | ✅ | ✅ | ❌ | ✅ |
| admin | ✅ | ✅ | ✅ | ✅ |

Check: `outlet.territory_code == user.territory_code OR user.role = 'admin'`

---

## Products Reference

| Code | Category | Unit |
|---|---|---|
| MS | fuel | kL |
| HSD | fuel | kL |
| SPEED | fuel | kL |
| QOC | non_fuel | Nos |
| Lubricants | non_fuel | ₹ |
| UFill | non_fuel | Nos |
| SBI | non_fuel | Nos |
| BeCafe | non_fuel | ₹ |

---

## Test Users (dev only)

| Employee ID | Password   | Role               |
|------------|------------|--------------------|
| EMP10001    | Bpcl@2026  | admin              |
| EMP10002    | Bpcl@2026  | territory_manager  |
| EMP10005    | Bpcl@2026  | ro_manager         |

---

## Quick Commands

```bash
# Start dev environment
docker-compose up -d postgres
make migrate
psql $BPCL_DB_URL < scripts/seed_dev.sql
make run

# Build
make build          # Go backend → bin/api
cd Bpclssoportal-main && npm run build   # Frontend

# Test
make test           # Go unit tests
cd Bpclssoportal-main && npm test        # Frontend Vitest tests
npx playwright test  # E2E tests

# Migrations
make migrate        # Run all pending migrations
make migrate-down   # Rollback last migration
make migrate-status # Check current version

# Lint
make lint           # golangci-lint
```

---

## When Something Breaks

1. Check the relevant exec-plan doc for the verification step
2. Check `docs/design-docs/scoring-engine.md` for scoring logic errors
3. Check `architecture/CONTRACTS.md` for shape mismatches
4. Never "try harder" — identify what capability is missing and add it to docs first

---

## What Agents Cannot See (encode it or it doesn't exist)

- Slack discussions about architecture decisions → encode in docs/design-docs/
- Excel column format of actual upload files → document in parser/excel.go header comment
- BPCL business rules not in these docs → ask the human, then write it down before building
