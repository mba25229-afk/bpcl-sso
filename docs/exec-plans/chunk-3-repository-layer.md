# Chunk 3 — Repository Layer
## All SQL query functions. No business logic here.

---

## Context
Models and DB pool exist (Chunk 2 complete). This chunk wires SQL to Go.
Repository functions are pure data access: take context + params, return (T, error).
No access checks, no KPI computation — that's the service layer's job.
Read AGENTS.md before starting.

## Rules (non-negotiable)
- Every function: `func (r *XRepo) Method(ctx context.Context, ...) (T, error)`
- `pgx.ErrNoRows` → return `nil, model.ErrNotFound`
- No `fmt.Sprintf` in SQL. Parameters only ($1, $2...)
- No business logic — just translate DB rows to model structs

---

## Task 1 — UserRepository (`internal/repository/user.go`)
Functions needed:
- `GetByEmployeeID(ctx, employeeID string) (*model.User, error)`
- `GetByID(ctx, id uuid.UUID) (*model.User, error)`
- `UpdateLastLogin(ctx, id uuid.UUID) error`

---

## Task 2 — OutletRepository (`internal/repository/outlet.go`)
Functions needed:
- `GetByCC(ctx, cc string) (*model.RetailOutlet, error)`
- `ListByTerritory(ctx, territoryCode string) ([]*model.RetailOutlet, error)`
- `ListAll(ctx) ([]*model.RetailOutlet, error)`

---

## Task 3 — PerformanceRepository (`internal/repository/performance.go`)
Functions needed:
- `GetByPeriod(ctx, cc string, period time.Time) ([]*model.PerformanceRecord, error)`
- `GetDateRange(ctx, cc string, from, to time.Time) ([]*model.PerformanceRecord, error)`
- `Upsert(ctx, rec *model.PerformanceRecord) error`
- `UpsertBatch(ctx, pool *pgxpool.Pool, recs []*model.PerformanceRecord) error` — use pgx batch

---

## Task 4 — TargetRepository (`internal/repository/target.go`)
Functions needed:
- `GetByPeriod(ctx, cc string, period time.Time) ([]*model.Target, error)`
- `Upsert(ctx, t *model.Target) error`
- `UpsertBatch(ctx, targets []*model.Target) error`

---

## Task 5 — UploadRepository (`internal/repository/upload.go`)
Functions needed:
- `Create(ctx, f *model.UploadedFile) error`
- `GetByID(ctx, id uuid.UUID) (*model.UploadedFile, error)`
- `ListByUser(ctx, userID uuid.UUID, cc string, limit, offset int) ([]*model.UploadedFile, int, error)`
- `UpdateStatus(ctx, id uuid.UUID, status, errMsg string, rowCount int) error`
- `SoftDelete(ctx, id uuid.UUID) error`

---

## Task 6 — AuditRepository (`internal/repository/audit.go`)
Functions needed:
- `Log(ctx, userID uuid.UUID, action, cc string, payload any) error`
  Note: payload marshalled to JSONB. Non-blocking — if audit write fails, log it but don't return error to caller.

---

## Task 7 — CompetitionRepository (`internal/repository/competition.go`)
Functions needed:
- `GetActivePeriod(ctx, territoryCode string) (*model.CompetitionPeriod, error)`
- `GetScores(ctx, competitionID uuid.UUID) ([]*model.CompetitionScore, error)`
- `GetDealerScore(ctx, competitionID uuid.UUID, cc string) (*model.CompetitionScore, error)`
- `GetAuditScore(ctx, cc string, period time.Time) (*model.DealerAuditScore, error)`
- `GetBonus(ctx, competitionID uuid.UUID, cc string) (*model.CompetitionBonus, error)`
- `UpsertBonus(ctx, b *model.CompetitionBonus) error`

---

## Task 8 — Repository Tests

For each repository, write a test file (`*_test.go`) that:
- Uses a real DB (postgres running locally — not mocks)
- Inserts test data, calls the function, asserts result, cleans up
- Runs with: `go test ./internal/repository/... -v`

**Verify:**
```bash
go test ./internal/repository/... -v
```
All tests pass. No SQL injection vectors (sqlvet or manual review).

---

## Done When
- `go build ./internal/repository/...` succeeds
- All repository tests pass against the seeded DB
- No pgx.ErrNoRows leaking out of any function (all converted to model.ErrNotFound)

## Next Chunk
→ `docs/exec-plans/chunk-4-service-layer.md`
