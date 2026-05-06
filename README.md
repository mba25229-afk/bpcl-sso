# BPCL SSO Portal — Backend API

## Quick Start (Development)

```bash
docker-compose up -d postgres
cp configs/development.env.example .env
make migrate
psql $BPCL_DB_URL < scripts/seed_dev.sql
make run
```

## Production Deployment

```bash
cp configs/production.env.example .env
# Edit .env — set BPCL_DB_URL, BPCL_JWT_SECRET (use: make generate-secret)
docker-compose -f docker-compose.prod.yml build
docker-compose -f docker-compose.prod.yml up -d
make migrate
```

## Architecture

- Go 1.22 + Chi router + pgx/v5 + PostgreSQL 16
- 3-layer: Handler → Service → Repository
- JWT auth, role-based access control
- See docs/ARCHITECTURE.md for full design

## Test Users (dev only)

| Employee ID | Password   | Role               |
|------------|------------|--------------------|
| EMP10001    | Bpcl@2026  | admin              |
| EMP10002    | Bpcl@2026  | territory_manager  |
| EMP10005    | Bpcl@2026  | ro_manager         |

## API Endpoints

See architecture/CONTRACTS.md for full contract specification.

### Admin Users

- `GET /api/v1/admin/users` — List users (admin only)
- `POST /api/v1/admin/users` — Create user (admin only)
- `PUT /api/v1/admin/users/:id` — Update user (admin only)
- `PUT /api/v1/admin/users/:id/reset-password` — Reset password (admin only)

### Territory

- `GET /api/v1/territory/summary` — Territory overview (territory_manager/admin)

### Performance

- `GET /api/v1/outlets/:cc/performance` — Dashboard KPIs
- `GET /api/v1/outlets/:cc/analysis` — Trend analysis
- `GET /api/v1/outlets/:cc/trend` — Performance trend chart

## Build

```bash
cd bpcl-portal-api && go build ./...
```

## Frontend

```bash
cd Bpclssoportal-main && npm run build
```