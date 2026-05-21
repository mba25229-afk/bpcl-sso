# Exec Plan: Crystal MVP
> Status: ACTIVE | Owner: Engineering | Updated: May 2026

## Definition of Done
- 41 dealers visible in dashboard with live MTD data
- JWT auth with 20-min inactivity logout working
- Cron runs nightly, pings health endpoint, failure alerts visible
- All backend + frontend tests passing in CI
- Playwright: 7 flows passing against staging

---

## Chunk 1 — Foundation (do first, everything depends on this)
**Goal:** DB + auth working locally. No ETL yet.

- [ ] Create PostgreSQL schema (from `docs/generated/db-schema.md`)
- [ ] Seed 40 dealers into `dealers` table
- [ ] Implement `src/middleware/auth.ts` (verifyToken, signAccessToken, signRefreshToken)
- [ ] POST `/api/auth/login` — bcrypt verify, return both tokens
- [ ] POST `/api/auth/refresh` — verify refresh token, return new access token
- [ ] GET `/api/health` — unauthenticated, returns `{status: ok}`
- [ ] Backend tests: auth routes (4 tests from test-strategy.md)
- [ ] Backend tests: dealers table (2 tests)

**Agent instruction:** Build only what's listed. Do not build the frontend yet. Do not build ETL yet.

---

## Chunk 2 — ETL Pipeline (depends on Chunk 1)
**Goal:** Google Sheet → PostgreSQL working with validation.

- [ ] `src/etl/schema.ts` — Zod schemas for DailyActualRow, TargetRow, excelSerialToDate
- [ ] `src/etl/ingest.ts` — fetch Sheet via googleapis, parse, validate, upsert
- [ ] `src/etl/healthcheck.ts` — POST ping to health endpoint on completion
- [ ] POST `/api/health/cron-ping` — receives ping, writes to `etl_log`
- [ ] Create `etl_log` table
- [ ] Setup crontab entry (see `docs/design-docs/etl.md`)
- [ ] Backend tests: Zod schema (5 tests), upsert idempotency (1 test), excelSerialToDate (2 tests)
- [ ] Add curl healthcheck smoke test to CI

**Agent instruction:** ETL must fail loudly. Silent failures are not acceptable. All errors must log to stdout AND ping the health endpoint with `status: error`.

---

## Chunk 3 — API Layer (depends on Chunk 2)
**Goal:** Frontend-ready REST endpoints.

- [ ] GET `/api/dealers` — list all 40 dealers (auth required)
- [ ] GET `/api/dealers/:cc_code/mtd?date=YYYY-MM-DD` — MTD for a dealer up to date
- [ ] GET `/api/dashboard?month=YYYY-MM` — all dealers, all MTD metrics for month
- [ ] GET `/api/health/cron-status` — last ETL run info (auth required)
- [ ] Backend tests: MTD computation (test-strategy.md §5)

---

## Chunk 4 — Frontend Auth (depends on Chunk 1)
**Goal:** Login, session management, inactivity logout.

- [ ] `/login` page — email + password form, calls POST `/api/auth/login`
- [ ] Store tokens in localStorage
- [ ] `src/lib/api.ts` — axios instance with auth interceptor + silent refresh
- [ ] `src/hooks/useInactivityLogout.ts` — 20-min timer, reset on activity
- [ ] Session warning banner at T-1 min
- [ ] `/login?reason=inactivity` message display
- [ ] Forgot password flow: email input → OTP input → new password
- [ ] Frontend tests: useInactivityLogout (2 tests), session warning (1 test)
- [ ] Playwright: login.spec, inactivity.spec, forgot-password.spec, logout.spec

---

## Chunk 5 — Dashboard UI (depends on Chunk 3 + Chunk 4)
**Goal:** 40 dealers visible, data correct, date filter working.

- [ ] Dashboard page — table of all dealers with MTD columns: UFILL, QOC, SPEED, MS, HSD
- [ ] Colour coding: green ≥ 100% target, amber 80–99%, red < 80%
- [ ] Date range selector (from–to, defaults to current month)
- [ ] Dealer drill-down → Report Card page (Target / Achievement / Gap)
- [ ] Cron status indicator in header
- [ ] Frontend tests: DealerCard badges (2 tests), Dashboard renders 40 rows (1 test)
- [ ] Playwright: dashboard-filter.spec, dealer-drilldown.spec, cron-status.spec

---

## Chunk 6 — Hardening (last, before go-live)
**Goal:** Production-safe.

- [ ] All env vars in `.env.example` (no secrets committed)
- [ ] Rate limiting on `/api/auth/*` (max 10 req/min per IP)
- [ ] POST `/api/auth/forgot-password` — OTP flow (see `docs/design-docs/auth.md`)
- [ ] Add UptimeRobot monitor on `/api/health` (free tier, alerts if no response in 25h)
- [ ] Final Playwright run against staging — all 7 flows green

---

## Invariants (enforced in CI, never skipped)
1. `npm run test:backend` must pass before any merge
2. `npm run test:frontend` must pass before any merge
3. ETL schema Zod tests must include all fields from `docs/generated/db-schema.md`
4. No hardcoded cc_codes in application code (always query DB)
5. The curl health ping smoke test must be in CI pipeline
