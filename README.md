# BPCL SSO Portal

Bharat Petroleum Corporation Limited — Delhi Retail Operations Portal

A two-part platform for BPCL Delhi retail operations:
1. **SSO Dashboard** — Sales Officers view per-outlet KPIs (fuel + non-fuel performance vs. targets)
2. **Boost and Win** — Monthly dealer competition with a 100-point scoring engine and leaderboard
3. **Crystal Subsystem** — ETL ingestion, scoring, admin portal, dealer scorecards, MTD dashboards

## Quick Start

```bash
# Start the entire stack (PostgreSQL + Backend + Frontend)
cd docker
docker compose up

# Or in detached mode
docker compose up -d
```

### Access Points

| Service | URL | Description |
|---------|-----|-------------|
| Frontend | http://localhost:5173 | React dev server with hot reload |
| Backend API | http://localhost:8080 | Go API server |
| PostgreSQL | localhost:5432 | Database |
| Health Check | http://localhost:8080/health | API health endpoint |

### Default Test Users

All users have password: **`password123`**

| Employee ID | Role | Access |
|------------|------|--------|
| `EMP10001` | admin | All outlets, admin panel, set targets |
| `EMP10002` | territory_manager | Territory outlets (DELHI-W), set targets |
| `EMP10005` | ro_manager | Own outlet only (112847) |

## Project Structure

```
├── backend/                 # Go backend API
│   ├── cmd/api/            # Entry point
│   ├── internal/           # Business logic (handler, service, repository)
│   ├── migrations/         # Database migrations (copied from root)
│   ├── Dockerfile          # Multi-stage build with migration runner
│   └── docker-entrypoint.sh # Runs migrations, then starts server
├── frontend/               # React frontend (Vite + TypeScript)
│   ├── src/
│   │   ├── api/            # API client
│   │   ├── components/     # UI components (ui/, shared/, outlet/, sso/)
│   │   ├── hooks/          # React hooks (useAuth)
│   │   ├── layouts/        # Layout components (OutletLayout, SSOLayout)
│   │   ├── pages/          # Page components (outlet/, sso/, admin/)
│   │   └── lib/            # Utilities (cn)
│   └── vite.config.ts      # Vite config with API proxy
├── docker/
│   ├── docker-compose.yml  # Orchestrates postgres, backend, frontend
│   ├── .env                # Environment variables for compose
│   └── init/               # SQL seed data (runs on first DB startup)
├── migrations/             # Database migrations (source of truth)
├── etl/                    # Standalone ETL service (Microsoft Graph)
└── docs/                   # Design documentation
```

## Architecture

### Three-Layer Backend
```
HTTP Handler  →  Service Layer  →  Repository Layer  →  PostgreSQL
   (Chi)         (Business logic)   (Raw pgx SQL)
```

### Middleware Chain
```
SecurityHeaders → RequestTimeout(30s) → RequestID → Logger → CORS → RateLimit → [Auth]
```

### Role-Based Access Control

| Role | Own Outlet | Territory Outlets | All Outlets | Set Targets |
|------|-----------|-------------------|-------------|-------------|
| ro_manager | ✅ | ❌ | ❌ | ❌ |
| territory_manager | ✅ | ✅ | ❌ | ✅ |
| admin | ✅ | ✅ | ✅ | ✅ |

## Development

### Backend (Go)

```bash
cd backend
go run ./cmd/api
```

### Frontend (React)

```bash
cd frontend
npm install
npm run dev
```

### Database Migrations

```bash
# Run all pending migrations
migrate -path docker/migrations -database "postgresql://bpcl:bpcl_secret@localhost:5432/bpcl_portal?sslmode=disable" up

# Rollback last migration
migrate -path docker/migrations -database "postgresql://bpcl:bpcl_secret@localhost:5432/bpcl_portal?sslmode=disable" down
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26, Chi router, pgx/v5 |
| Frontend | React 19, TypeScript, Vite, Tailwind v4, shadcn/ui, MUI |
| Database | PostgreSQL 16 |
| Auth | JWT (Bearer tokens) |
| ETL | Microsoft Graph API, robfig/cron |
| Testing | Vitest (frontend), Go testing (backend), Playwright (E2E) |

## Environment Variables

See `backend/.env.example` for all available configuration options.

Key variables:
- `BPCL_DB_URL` — PostgreSQL connection string
- `JWT_SECRET` — Secret for signing JWT tokens
- `CORS_ORIGINS` — Allowed origins for CORS
- `SMTP_*` — SMTP configuration for OTP emails
