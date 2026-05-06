# Chunk 1 — Database Schema
## Execute this first. Nothing else runs until the DB is clean.

---

## Context
You are setting up a fresh PostgreSQL database for the BPCL SSO Portal.
No backend code exists yet. This chunk only creates SQL schema files and runs migrations.
Read AGENTS.md before starting.

## Prerequisites
- Docker installed and running
- PostgreSQL image available (postgres:16)
- golang-migrate CLI installed (`brew install golang-migrate` or download binary)

---

## Task 1 — Start PostgreSQL

Create `docker-compose.yml` in project root:
```yaml
version: '3.9'
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: bpcl
      POSTGRES_PASSWORD: bpcl
      POSTGRES_DB: bpcl_portal
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bpcl -d bpcl_portal"]
      interval: 5s
      timeout: 5s
      retries: 5
volumes:
  pgdata:
```

Run: `docker-compose up -d postgres`

**Verify:** `docker-compose ps` shows postgres as healthy

---

## Task 2 — Create .env.example and .env

```bash
BPCL_DB_URL=postgres://bpcl:bpcl@localhost:5432/bpcl_portal
BPCL_JWT_SECRET=change-this-to-32-plus-random-chars-before-production
BPCL_JWT_EXPIRY=24h
BPCL_PORT=8080
BPCL_UPLOAD_DIR=/tmp/bpcl-uploads
BPCL_CORS_ORIGINS=http://localhost:5173
BPCL_MAX_UPLOAD_MB=10
BPCL_LOG_LEVEL=info
BPCL_ENV=development
BPCL_RATE_LIMIT_RPM=100
```

Copy to `.env`. Add `.env` to `.gitignore`. Commit `.env.example`.

---

## Task 3 — Create Migration Script

Create `scripts/migrate.sh`:
```bash
#!/bin/bash
set -euo pipefail
source .env
migrate -path migrations -database "$BPCL_DB_URL" "$@"
```
`chmod +x scripts/migrate.sh`

Create `Makefile` with just the migrate target for now:
```makefile
.PHONY: migrate migrate-down migrate-status

migrate:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down 1

migrate-status:
	./scripts/migrate.sh version
```

---

## Task 4 — Run All Migrations

The migrations are already in `migrations/` directory.
Run them in order (golang-migrate does this automatically):

```bash
make migrate
```

Expected output:
```
1/u create_users (done)
2/u create_products (done)
3/u create_outlets (done)
4/u create_performance_records (done)
5/u create_targets (done)
6/u create_uploads (done)
7/u create_audit_log (done)
8/u seed_products (done)
9/u create_competition (done)
```

**Verify:**
```bash
psql postgres://bpcl:bpcl@localhost:5432/bpcl_portal -c "\dt"
```
Expected: 14 tables listed (users, products, trading_areas, retail_outlets,
performance_records, targets, uploaded_files, audit_log, competition_periods,
competition_scores, dealer_audit_scores, competition_bonus, market_share_data,
trading_area_totals)

---

## Task 5 — Run Seed Data

```bash
psql postgres://bpcl:bpcl@localhost:5432/bpcl_portal < scripts/seed_dev.sql
```

**Verify — expected row counts:**
```
users                   9
trading_areas          27
retail_outlets         40   (39 dealers + MAHADEV)
products                8
performance_records  7680   (40 outlets × 8 products × 24 months)
targets               ~280
competition_periods     1
competition_scores     40
dealer_audit_scores    40
competition_bonus       1
```

If counts differ, check the PL/pgSQL block output for errors.

---

## Done When
- `make migrate` runs clean with 9 migrations applied
- `make migrate-status` shows version 9
- Seed data row counts match above
- No errors in psql output

## Next Chunk
→ `docs/exec-plans/chunk-2-backend-foundation.md`
