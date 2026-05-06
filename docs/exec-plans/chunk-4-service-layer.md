# Chunk 4 — Service + Handler Layer
## Business logic, KPI computation, HTTP handlers

---

## Context
Repository layer is complete and tested (Chunk 3 done).
This chunk adds the business logic (service) and HTTP layer (handlers).
Read AGENTS.md and docs/design-docs/scoring-engine.md before starting.

## Rules
- Service first call for any outlet endpoint: verify user can see this CC
- Access check: `outlet.territory_code == user.territory_code OR user.role == 'admin'`
- Never call repository directly from handler
- Handlers: parse → validate → call service → write response. Nothing else.

---

## Task 1 — Auth Service (`internal/service/auth.go`)

Functions:
- `Login(ctx, employeeID, password string) (*LoginResponse, error)`
  - Fetch user by employee_id
  - Compare bcrypt hash (golang.org/x/crypto/bcrypt)
  - Generate JWT (HS256, claims: sub=userID, role, territory_code, exp)
  - Update last_login_at
  - Return token + user info
- `ValidateToken(tokenStr string) (*Claims, error)` — parse and verify JWT
- `ExtractUserID(ctx) (uuid.UUID, error)` — get from context (set by middleware)

---

## Task 2 — Outlet Service (`internal/service/outlet.go`)

Functions:
- `GetOutlet(ctx, cc string, userID uuid.UUID) (*model.RetailOutlet, error)`
  - Access check first
  - Audit log: "view_outlet"
  - Return outlet

---

## Task 3 — Performance Service (`internal/service/performance.go`)

Functions:
- `GetPerformance(ctx, cc string, period time.Time, userID uuid.UUID) (*PerformanceResponse, error)`
  - Access check
  - Fetch performance records + targets for the period
  - Compute KPIs:
    - total_revenue_cr = SUM(achieved) / 10,000,000
    - target_achievement_pct = SUM(achieved) / SUM(target) * 100
    - yoy_growth_pct = (SUM(achieved) - SUM(last_year)) / SUM(last_year) * 100
    - fuel_vs_nonfuel_ratio = fuel_achieved : nonfuel_achieved
  - Separate fuel[] and non_fuel[] arrays per CONTRACTS.md shape
  - Each row: product, target, achieved, last_year, volume_kl, achievement_pct, yoy_pct

- `GetAnalysis(ctx, cc string, from, to time.Time, userID uuid.UUID) (*AnalysisResponse, error)`
  - Access check
  - Fetch date range records
  - Build: fuel_mix, non_fuel_mix, target_vs_achieved, monthly_growth, category_growth
  - weekly_trend: divide each month's data into 4 equal weeks by volume proportion

---

## Task 4 — Target Service (`internal/service/target.go`)

Functions:
- `GetTargets(ctx, cc string, period time.Time, userID uuid.UUID) (*TargetsResponse, error)`
- `SetTargets(ctx, cc string, period time.Time, targets TargetInput, userID uuid.UUID) (*TargetsResponse, error)`
  - Role check: role must be territory_manager or admin → ErrForbidden
  - Upsert all products
  - Audit log: "set_target" with payload={period, products changed}

---

## Task 5 — Upload Service (`internal/service/upload.go`)

Functions:
- `HandleUpload(ctx, file multipart.File, header *multipart.FileHeader, cc string, period time.Time, userID uuid.UUID) (*UploadResponse, error)`
  - Validate: file type (xlsx/xls/csv), size ≤ max
  - Save to BPCL_UPLOAD_DIR
  - Create uploaded_files record (status=pending)
  - Launch background goroutine: ProcessAsync(ctx, uploadID)
  - Return immediately with upload ID

- `ProcessAsync(uploadID uuid.UUID)` — background goroutine
  - Parse Excel file (call parser.Parse)
  - Upsert performance_records
  - Update status to done/failed
  - goroutine must respect context cancellation

- `GetHistory(ctx, userID uuid.UUID, cc string, limit, offset int) (*UploadListResponse, error)`

---

## Task 6 — Competition Service (`internal/service/competition.go`)

Functions:
- `GetLeaderboard(ctx, territoryCode string, userID uuid.UUID) (*LeaderboardResponse, error)`
  - Fetch active competition period
  - Fetch all scores for that period
  - Filter: outlet_type = 'regular' only (Bug #5 fix)
  - Sort by total_score DESC, add rank
  - Return top 10 highlighted, full list available

- `GetDealerScorecard(ctx, cc string, competitionID uuid.UUID, userID uuid.UUID) (*ScorecardResponse, error)`
  - Access check
  - Return full parameter breakdown for one dealer

---

## Task 7 — Excel Parser (`internal/parser/excel.go`)

IMPORTANT: Read docs/design-docs/data-pipeline.md before implementing.

Two parse modes (detect from column headers):
1. `ParsePerformance(filePath string) ([]RawPerformanceRow, error)` — standard upload
2. `ParseDelhiMaster(filePath string) ([]RawMarketShareRow, error)` — Delhi Master format

Each returns raw string values. Type conversion happens in service layer.
Invalid rows: skip and count (do not fail the whole file).

---

## Task 8 — All Handlers

For each service, create corresponding handler. Pattern:
```go
func (h *Handler) Method(w http.ResponseWriter, r *http.Request) {
    // 1. Parse URL params + query params
    // 2. Validate (return 422 with field errors if invalid)  
    // 3. Call service
    // 4. Map service errors to HTTP status:
    //    ErrNotFound → 404, ErrForbidden → 403, ErrUnauthorized → 401
    // 5. writeJSON(w, statusCode, response)
}
```

Create `internal/handler/response.go` with helpers:
- `writeJSON(w, status int, v any)`
- `writeError(w, status int, msg, code string, requestID string)`

---

## Task 9 — Service + Handler Tests

Write unit tests for each service using repository mocks (use `testify/mock`).
Write handler tests using `httptest.NewRecorder`.

**Verify:**
```bash
go test ./internal/service/... ./internal/handler/... -v
```

---

## Done When
- `go build ./internal/service/... ./internal/handler/...` succeeds
- All service + handler tests pass
- KPI computation produces correct values for the seeded March 2026 data

## Next Chunk
→ `docs/exec-plans/chunk-5-router-main.md`
