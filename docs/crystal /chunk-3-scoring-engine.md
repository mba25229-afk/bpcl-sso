# Chunk 3 — Scoring Engine (Go)
**Status:** READY TO BUILD  
**Stack:** Go, pgx/v5  
**Depends on:** Chunk 1 (schema ✓), Chunk 2 (ingest API ✓)  
**Blocks:** Chunk 4 (Admin Portal), Chunk 5 (SSO Portal)

---

## What This Chunk Delivers

- Pure ranking engine — formula verified 40/40 against v4 template
- One SQL query fetches all actuals for all 40 dealers (no N+1)
- Idempotent — run mid-month for monitoring, end-of-month for final results
- One new DB migration (manual scores table for bonus + cleanliness)
- One API endpoint to trigger a scoring run

---

## DB Migration (append to Chunk 1)

```sql
-- dealer_manual_scores: admin-entered bonus marks and cleanliness grades
CREATE TABLE dealer_manual_scores (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cc_code            VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
    month_year         DATE NOT NULL,
    bonus_marks        NUMERIC(5,2) CHECK (bonus_marks BETWEEN 0 AND 10),
    cleanliness_grade  TEXT CHECK (cleanliness_grade IN (
                           'Excellent','Good','Average','Below Average','Poor'
                       )),
    entered_by         TEXT,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cc_code, month_year),
    FOREIGN KEY (month_year) REFERENCES competition_periods(month_year)
);
```

**Rollback:**
```sql
DROP TABLE IF EXISTS dealer_manual_scores;
```

---

## Directory Layout

```
internal/scoring/
├── engine.go       -- orchestrates full scoring run
├── actuals.go      -- single SQL query fetches all MTD actuals
├── ranker.go       -- assigns ranks 1–40 per metric
├── formula.go      -- pure scoring functions (no DB, fully testable)
├── repository.go   -- bulk upsert to dealer_scores
└── types.go        -- all structs
```

---

## internal/scoring/types.go

```go
package scoring

import "time"

// Actuals holds all MTD metric values for one dealer.
// Fetched in a single SQL query across all source tables.
type Actuals struct {
    CCCode           string
    MSVolKL          float64 // daily_ms + daily_speed MTD
    SpeedVolKL       float64 // daily_speed MTD only
    HSDVolKL         float64 // daily_hsd MTD
    MSLYKl           float64 // from monthly_targets.ms_ly
    HSDLYKl          float64 // from monthly_targets.hsd_ly
    QOCCount         float64 // daily_qoc MTD
    UFillCount       float64 // daily_ufill MTD
    MAKGESalesKL     float64 // mak_ge delta (last - first reading)
    MAKGEHasData     bool    // false = < 2 readings this month
    GoogleComposite  float64 // rating × review_count — used for ranking
    GoogleRating     float64 // raw rating — stored for display
    SangamCerts      float64 // sangam_data.cert_count
    BonusMarks       float64 // dealer_manual_scores.bonus_marks
    CleanlinessGrade string  // dealer_manual_scores.cleanliness_grade
}

// RankedActuals wraps Actuals with computed ranks per metric.
type RankedActuals struct {
    Actuals
    Ranks map[string]int // metric_key → rank (1 = best)
}

// ScoringParam mirrors scoring_params row for one metric.
type ScoringParam struct {
    MetricKey            string
    MaxMarks             float64
    NegativeScaleEnabled bool
    IsActive             bool
}

// ScoreRow is one row written to dealer_scores.
type ScoreRow struct {
    CCCode      string
    MonthYear   time.Time
    MetricKey   string
    ActualValue *float64 // nil for inactive metrics
    RankInGroup *int     // nil for grade/manual metrics
    MarksScored *float64 // nil if inactive or insufficient data
    MaxMarks    float64
}
```

---

## internal/scoring/formula.go

Pure functions — zero DB interaction. All tests run against this file.

```go
package scoring

import "math"

// Standard rank-based score.
// Formula: ((n - rank) / (n - 1)) × maxMarks
// Verified: 40/40 dealers match v4 template output.
func StandardScore(rank, n int, maxMarks float64) float64 {
    if n <= 1 {
        return maxMarks
    }
    return round2((float64(n-rank) / float64(n-1)) * maxMarks)
}

// NegativeScaleScore for growth metrics where below-average = negative marks.
// Formula: ((n - rank) / (n - 1)) × 2 × maxMarks − maxMarks
// Range: [−maxMarks, +maxMarks]
// rank 1  (best)   → +maxMarks
// rank 20 (median) → ≈ 0
// rank 40 (worst)  → −maxMarks
func NegativeScaleScore(rank, n int, maxMarks float64) float64 {
    if n <= 1 {
        return 0
    }
    return round2(((float64(n-rank)/float64(n-1))*2*maxMarks) - maxMarks)
}

// CleanlinessScore maps grade string to fixed marks.
// Not rank-based — absolute lookup.
func CleanlinessScore(grade string) float64 {
    grades := map[string]float64{
        "Excellent":     5,
        "Good":          4,
        "Average":       3,
        "Below Average": 1,
        "Poor":          0,
    }
    if s, ok := grades[grade]; ok {
        return s
    }
    return 0
}

// GrowthPct computes year-over-year growth percentage.
// Returns 0 if LY is zero (no division by zero).
func GrowthPct(actual, ly float64) float64 {
    if ly == 0 {
        return 0
    }
    return round2(((actual - ly) / ly) * 100)
}

func round2(f float64) float64 {
    return math.Round(f*100) / 100
}
```

---

## internal/scoring/actuals.go

One query. All 40 dealers. All metrics. No N+1.

```go
package scoring

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type ActualsRepo struct {
    pool *pgxpool.Pool
}

func NewActualsRepo(pool *pgxpool.Pool) *ActualsRepo {
    return &ActualsRepo{pool: pool}
}

// FetchAll returns MTD actuals for every active dealer as of asOf date.
// asOf is inclusive. Set asOf = last day of month for final scoring.
func (r *ActualsRepo) FetchAll(ctx context.Context, monthYear, asOf time.Time) ([]Actuals, error) {
    sql := `
WITH
ms_mtd AS (
    SELECT cc_code, COALESCE(SUM(kl), 0) AS kl
    FROM daily_ms
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
speed_mtd AS (
    SELECT cc_code, COALESCE(SUM(kl), 0) AS kl
    FROM daily_speed
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
hsd_mtd AS (
    SELECT cc_code, COALESCE(SUM(kl), 0) AS kl
    FROM daily_hsd
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
qoc_mtd AS (
    SELECT cc_code, COALESCE(SUM(count), 0) AS cnt
    FROM daily_qoc
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
ufill_mtd AS (
    SELECT cc_code, COALESCE(SUM(count), 0) AS cnt
    FROM daily_ufill
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
mak_ge_delta AS (
    SELECT
        cc_code,
        MAX(meter_reading) - MIN(meter_reading) AS sales_kl,
        COUNT(*) AS reading_count
    FROM mak_ge_readings
    WHERE DATE_TRUNC('month', reading_date) = $1 AND reading_date <= $2
    GROUP BY cc_code
),
google_latest AS (
    SELECT DISTINCT ON (cc_code)
        cc_code,
        rating,
        review_count,
        rating * review_count AS composite
    FROM google_ratings
    WHERE DATE_TRUNC('month', snapshot_date) = $1 AND snapshot_date <= $2
    ORDER BY cc_code, snapshot_date DESC
),
sangam_mtd AS (
    SELECT cc_code, COALESCE(cert_count, 0) AS cert_count
    FROM sangam_data
    WHERE month_year = $1
),
manual AS (
    SELECT cc_code,
           COALESCE(bonus_marks, 0)    AS bonus_marks,
           COALESCE(cleanliness_grade, '') AS cleanliness_grade
    FROM dealer_manual_scores
    WHERE month_year = $1
)
SELECT
    d.cc_code,
    COALESCE(ms.kl,  0)                        AS ms_kl,
    COALESCE(sp.kl,  0)                        AS speed_kl,
    COALESCE(hsd.kl, 0)                        AS hsd_kl,
    COALESCE(t.ms_ly,  0)                      AS ms_ly,
    COALESCE(t.hsd_ly, 0)                      AS hsd_ly,
    COALESCE(qoc.cnt, 0)                       AS qoc_count,
    COALESCE(uf.cnt,  0)                       AS ufill_count,
    COALESCE(ge.sales_kl, 0)                   AS mak_ge_sales,
    COALESCE(ge.reading_count, 0) >= 2         AS mak_ge_has_data,
    COALESCE(gr.rating, 0)                     AS google_rating,
    COALESCE(gr.composite, 0)                  AS google_composite,
    COALESCE(sg.cert_count, 0)                 AS sangam_certs,
    COALESCE(m.bonus_marks, 0)                 AS bonus_marks,
    COALESCE(m.cleanliness_grade, '')          AS cleanliness_grade
FROM dealers d
LEFT JOIN ms_mtd       ms  ON ms.cc_code  = d.cc_code
LEFT JOIN speed_mtd    sp  ON sp.cc_code  = d.cc_code
LEFT JOIN hsd_mtd      hsd ON hsd.cc_code = d.cc_code
LEFT JOIN qoc_mtd      qoc ON qoc.cc_code = d.cc_code
LEFT JOIN ufill_mtd    uf  ON uf.cc_code  = d.cc_code
LEFT JOIN mak_ge_delta ge  ON ge.cc_code  = d.cc_code
LEFT JOIN google_latest gr ON gr.cc_code  = d.cc_code
LEFT JOIN sangam_mtd   sg  ON sg.cc_code  = d.cc_code
LEFT JOIN monthly_targets t ON t.cc_code  = d.cc_code AND t.month_year = $1
LEFT JOIN manual        m  ON m.cc_code   = d.cc_code
WHERE d.is_active = TRUE
    `

    rows, err := r.pool.Query(ctx, sql, monthYear, asOf)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Actuals
    for rows.Next() {
        var a Actuals
        if err := rows.Scan(
            &a.CCCode,
            &a.MSVolKL, &a.SpeedVolKL, &a.HSDVolKL,
            &a.MSLYKl, &a.HSDLYKl,
            &a.QOCCount, &a.UFillCount,
            &a.MAKGESalesKL, &a.MAKGEHasData,
            &a.GoogleRating, &a.GoogleComposite,
            &a.SangamCerts,
            &a.BonusMarks, &a.CleanlinessGrade,
        ); err != nil {
            return nil, err
        }
        // MS total = regular petrol + speed (business rule: ms-aggregation.md)
        a.MSVolKL = a.MSVolKL + a.SpeedVolKL
        out = append(out, a)
    }
    return out, rows.Err()
}
```

---

## internal/scoring/ranker.go

```go
package scoring

import "sort"

// metricValue extracts the rankable float64 for a given metric from Actuals.
// Growth metrics are converted to % before ranking.
func metricValue(a Actuals, metricKey string) float64 {
    switch metricKey {
    case "ms_absolute_vol":
        return a.MSVolKL
    case "ms_growth_pct":
        return GrowthPct(a.MSVolKL, a.MSLYKl)
    case "hsd_absolute_vol":
        return a.HSDVolKL
    case "hsd_growth_pct":
        return GrowthPct(a.HSDVolKL, a.HSDLYKl)
    case "oil_change_count":
        return a.QOCCount
    case "mak_ge_sales":
        if !a.MAKGEHasData {
            return -1 // sentinel: insufficient readings → ranks last
        }
        return a.MAKGESalesKL
    case "ufill_txns":
        return a.UFillCount
    case "speed_vol":
        return a.SpeedVolKL
    case "sangam_certs":
        return a.SangamCerts
    case "google_rating":
        return a.GoogleComposite // rank by rating × reviews
    default:
        return 0
    }
}

// Rank assigns rank 1..n to each dealer for a given metric.
// Rank 1 = highest value. Ties broken by stable sort order (first occurrence).
// Returns map[cc_code]rank.
func Rank(actuals []Actuals, metricKey string, n int) map[string]int {
    type entry struct {
        ccCode string
        value  float64
    }

    entries := make([]entry, len(actuals))
    for i, a := range actuals {
        entries[i] = entry{a.CCCode, metricValue(a, metricKey)}
    }

    // Stable sort descending — preserves insertion order for equal values
    sort.SliceStable(entries, func(i, j int) bool {
        return entries[i].value > entries[j].value
    })

    ranks := make(map[string]int, len(entries))
    for i, e := range entries {
        ranks[e.ccCode] = i + 1 // 1-indexed
    }

    // Dealers not in actuals (empty slots up to n) get rank n
    // This handles the fixed n=40 requirement
    for rank := len(entries) + 1; rank <= n; rank++ {
        // empty slots — no cc_code, don't add to map
        _ = rank
    }

    return ranks
}
```

---

## internal/scoring/engine.go

Orchestrates: fetch → rank → score → write. Idempotent.

```go
package scoring

import (
    "context"
    "fmt"
    "time"
)

type Engine struct {
    actuals    *ActualsRepo
    paramsRepo *ParamsRepo
    scoreRepo  *ScoreRepository
}

func NewEngine(actuals *ActualsRepo, params *ParamsRepo, scores *ScoreRepository) *Engine {
    return &Engine{actuals: actuals, paramsRepo: params, scoreRepo: scores}
}

// Run computes scores for all active dealers for monthYear, using data up to asOf.
// Idempotent: re-running overwrites previous results via UPSERT.
func (e *Engine) Run(ctx context.Context, monthYear, asOf time.Time) (int, error) {
    // 1. Load scoring params for this period
    params, err := e.paramsRepo.FetchActive(ctx, monthYear)
    if err != nil {
        return 0, fmt.Errorf("load scoring params: %w", err)
    }
    if len(params) == 0 {
        return 0, fmt.Errorf("no active scoring params for %s", monthYear.Format("2006-01"))
    }

    // 2. Fetch all MTD actuals (one query, all dealers)
    actuals, err := e.actuals.FetchAll(ctx, monthYear, asOf)
    if err != nil {
        return 0, fmt.Errorf("fetch actuals: %w", err)
    }
    if len(actuals) == 0 {
        return 0, fmt.Errorf("no active dealers found")
    }

    n := 40 // fixed competition slots — business rule: scoring-formula.md

    // 3. Build score rows
    var scoreRows []ScoreRow
    for _, param := range params {
        if !param.IsActive {
            continue
        }

        switch param.MetricKey {
        case "bonus":
            // Direct value — not rank-based
            for _, a := range actuals {
                v := a.BonusMarks
                m := min(v, param.MaxMarks) // cap at max_marks
                scoreRows = append(scoreRows, ScoreRow{
                    CCCode:      a.CCCode,
                    MonthYear:   monthYear,
                    MetricKey:   param.MetricKey,
                    ActualValue: &v,
                    MarksScored: &m,
                    MaxMarks:    param.MaxMarks,
                })
            }

        case "cleanliness_audit":
            // Grade lookup — not rank-based
            for _, a := range actuals {
                m := CleanlinessScore(a.CleanlinessGrade)
                scoreRows = append(scoreRows, ScoreRow{
                    CCCode:      a.CCCode,
                    MonthYear:   monthYear,
                    MetricKey:   param.MetricKey,
                    MarksScored: &m,
                    MaxMarks:    param.MaxMarks,
                })
            }

        case "mak_ge_sales":
            // Rank-based but NULL if insufficient readings
            ranks := Rank(actuals, param.MetricKey, n)
            for _, a := range actuals {
                v := a.MAKGESalesKL
                row := ScoreRow{
                    CCCode:      a.CCCode,
                    MonthYear:   monthYear,
                    MetricKey:   param.MetricKey,
                    ActualValue: &v,
                    MaxMarks:    param.MaxMarks,
                }
                if a.MAKGEHasData {
                    rank := ranks[a.CCCode]
                    m := StandardScore(rank, n, param.MaxMarks)
                    row.RankInGroup = &rank
                    row.MarksScored = &m
                }
                // else: MarksScored stays nil — excluded from total
                scoreRows = append(scoreRows, row)
            }

        default:
            // Standard or negative-scale rank-based metrics
            ranks := Rank(actuals, param.MetricKey, n)
            for _, a := range actuals {
                v := metricValue(a, param.MetricKey)
                rank := ranks[a.CCCode]
                var m float64
                if param.NegativeScaleEnabled {
                    m = NegativeScaleScore(rank, n, param.MaxMarks)
                } else {
                    m = StandardScore(rank, n, param.MaxMarks)
                }
                scoreRows = append(scoreRows, ScoreRow{
                    CCCode:      a.CCCode,
                    MonthYear:   monthYear,
                    MetricKey:   param.MetricKey,
                    ActualValue: &v,
                    RankInGroup: &rank,
                    MarksScored: &m,
                    MaxMarks:    param.MaxMarks,
                })
            }
        }
    }

    // 4. Bulk upsert to dealer_scores
    if err := e.scoreRepo.BulkUpsert(ctx, scoreRows); err != nil {
        return 0, fmt.Errorf("write scores: %w", err)
    }
    return len(scoreRows), nil
}

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}
```

---

## internal/scoring/repository.go

```go
package scoring

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type ScoreRepository struct {
    pool *pgxpool.Pool
}

func NewScoreRepository(pool *pgxpool.Pool) *ScoreRepository {
    return &ScoreRepository{pool: pool}
}

// BulkUpsert writes all score rows to dealer_scores.
// ON CONFLICT DO UPDATE — fully idempotent.
func (r *ScoreRepository) BulkUpsert(ctx context.Context, rows []ScoreRow) error {
    sql := `
        INSERT INTO dealer_scores
            (cc_code, month_year, metric_key, actual_value, rank_in_group, marks_scored, max_marks)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (cc_code, month_year, metric_key) DO UPDATE SET
            actual_value  = EXCLUDED.actual_value,
            rank_in_group = EXCLUDED.rank_in_group,
            marks_scored  = EXCLUDED.marks_scored,
            max_marks     = EXCLUDED.max_marks,
            computed_at   = NOW()
    `
    batch := &pgx.Batch{}
    for _, r := range rows {
        batch.Queue(sql, r.CCCode, r.MonthYear, r.MetricKey,
            r.ActualValue, r.RankInGroup, r.MarksScored, r.MaxMarks)
    }
    br := r.pool.SendBatch(ctx, batch)
    defer br.Close()

    for range rows {
        if _, err := br.Exec(); err != nil {
            return err
        }
    }
    return nil
}

// ParamsRepo loads scoring_params from DB.
type ParamsRepo struct {
    pool *pgxpool.Pool
}

func NewParamsRepo(pool *pgxpool.Pool) *ParamsRepo {
    return &ParamsRepo{pool: pool}
}

func (r *ParamsRepo) FetchActive(ctx context.Context, monthYear time.Time) ([]ScoringParam, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT metric_key, max_marks, negative_scale_enabled, is_active
         FROM scoring_params
         WHERE month_year = $1
         ORDER BY sort_order`,
        monthYear,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []ScoringParam
    for rows.Next() {
        var p ScoringParam
        if err := rows.Scan(&p.MetricKey, &p.MaxMarks, &p.NegativeScaleEnabled, &p.IsActive); err != nil {
            return nil, err
        }
        out = append(out, p)
    }
    return out, rows.Err()
}
```

---

## HTTP Trigger Endpoint

Add to `cmd/api/main.go` router and wire new handler:

```go
// internal/scoring/handler.go
package scoring

import (
    "encoding/json"
    "net/http"
    "time"
)

type Handler struct {
    engine *Engine
}

func NewHandler(engine *Engine) *Handler {
    return &Handler{engine: engine}
}

type ComputeRequest struct {
    MonthYear string `json:"month_year"` // "2026-05-01"
    AsOf      string `json:"as_of"`      // "2026-05-08" — omit for today
}

type ComputeResponse struct {
    RowsWritten int    `json:"rows_written"`
    MonthYear   string `json:"month_year"`
    AsOf        string `json:"as_of"`
}

// POST /api/v1/scoring/compute
func (h *Handler) Compute(w http.ResponseWriter, r *http.Request) {
    var req ComputeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }

    monthYear, err := time.Parse("2006-01-02", req.MonthYear)
    if err != nil || monthYear.Day() != 1 {
        writeError(w, http.StatusBadRequest, "month_year must be YYYY-MM-01")
        return
    }

    asOf := time.Now()
    if req.AsOf != "" {
        asOf, err = time.Parse("2006-01-02", req.AsOf)
        if err != nil {
            writeError(w, http.StatusBadRequest, "invalid as_of date")
            return
        }
    }

    n, err := h.engine.Run(r.Context(), monthYear, asOf)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, ComputeResponse{
        RowsWritten: n,
        MonthYear:   req.MonthYear,
        AsOf:        asOf.Format("2006-01-02"),
    })
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, map[string]string{"error": msg})
}
```

**Router addition in main.go:**
```go
scoringActuals := scoring.NewActualsRepo(pool)
scoringParams  := scoring.NewParamsRepo(pool)
scoringRepo    := scoring.NewScoreRepository(pool)
scoringEngine  := scoring.NewEngine(scoringActuals, scoringParams, scoringRepo)
scoringHandler := scoring.NewHandler(scoringEngine)

r.Route("/api/v1", func(r chi.Router) {
    // existing routes...
    r.Post("/scoring/compute", scoringHandler.Compute)
})
```

---

## Tests — formula.go (zero DB, fully isolated)

```go
// internal/scoring/formula_test.go
package scoring_test

import (
    "testing"
    "github.com/bpcl/sso-backend/internal/scoring"
)

func TestStandardScore(t *testing.T) {
    cases := []struct {
        rank, n  int
        max      float64
        expected float64
    }{
        {1, 40, 10, 10.0},  // rank 1 always gets full marks
        {40, 40, 10, 0.0},  // rank 40 always gets zero
        {4, 40, 10, 9.23},  // MANN — verified against v4 template
        {2, 40, 10, 9.74},  // AUTO CARE — verified against v4 template
    }
    for _, c := range cases {
        got := scoring.StandardScore(c.rank, c.n, c.max)
        if got != c.expected {
            t.Errorf("StandardScore(%d,%d,%.0f) = %.2f, want %.2f",
                c.rank, c.n, c.max, got, c.expected)
        }
    }
}

func TestNegativeScaleScore(t *testing.T) {
    cases := []struct {
        rank, n  int
        max      float64
        expected float64
    }{
        {1,  40, 5,  5.0},   // best growth → full +5
        {40, 40, 5, -5.0},   // worst growth → full -5
        {20, 40, 5,  0.13},  // median → near zero
    }
    for _, c := range cases {
        got := scoring.NegativeScaleScore(c.rank, c.n, c.max)
        if got != c.expected {
            t.Errorf("NegativeScaleScore(%d,%d,%.0f) = %.2f, want %.2f",
                c.rank, c.n, c.max, got, c.expected)
        }
    }
}

func TestCleanlinessScore(t *testing.T) {
    if scoring.CleanlinessScore("Excellent") != 5 { t.Fail() }
    if scoring.CleanlinessScore("Good")      != 4 { t.Fail() }
    if scoring.CleanlinessScore("Average")   != 3 { t.Fail() }
    if scoring.CleanlinessScore("")          != 0 { t.Fail() }
}

func TestGrowthPct(t *testing.T) {
    if scoring.GrowthPct(552.5, 507.5) != 8.87 {
        t.Errorf("growth pct mismatch")
    }
    if scoring.GrowthPct(100, 0) != 0 { // no division by zero
        t.Fail()
    }
}
```

**Run with:** `go test ./internal/scoring/... -v`

---

## API Reference

### POST /api/v1/scoring/compute

Trigger a full scoring run. Safe to call multiple times.

```json
// Request — mid-month snapshot
{ "month_year": "2026-05-01", "as_of": "2026-05-08" }

// Request — final end-of-month
{ "month_year": "2026-05-01", "as_of": "2026-05-31" }

// Response
{ "rows_written": 520, "month_year": "2026-05-01", "as_of": "2026-05-08" }
```

`rows_written = active_dealers × active_metrics = 40 × 13 = 520` for a full run.

---

## Invariants Enforced

| Rule | Where |
|---|---|
| n = 40 fixed | `engine.go` hardcoded constant with doc reference |
| MS = daily_ms + daily_speed | `actuals.go` line after Scan |
| MAK GE NULL if < 2 readings | `actuals.go` HAVING COUNT(*) >= 2 + engine nil guard |
| Bonus capped at max_marks | `engine.go` min() |
| Inactive metrics skipped | `engine.go` param.IsActive check |
| Idempotent writes | `repository.go` ON CONFLICT DO UPDATE |
| Growth % = 0 when LY = 0 | `formula.go` GrowthPct guard |

---

## Acceptance Criteria

- [ ] `go test ./internal/scoring/... -v` — all pass
- [ ] `POST /scoring/compute` with May data → rows_written = 520
- [ ] Re-run same request → rows_written = 520, no errors (idempotent)
- [ ] VAIBHAV (112458) ms_absolute_vol marks = 10.00 ± 0.01
- [ ] MANN (cc from data) ms_absolute_vol marks = 9.23 ± 0.01
- [ ] Dealer with 0 MAK GE readings → mak_ge_sales marks_scored = NULL in DB
- [ ] `dealer_total_scores` view ranks match v4 template order for March data

---

## What Is NOT in This Chunk

| Feature | Chunk |
|---|---|
| Admin UI to enter bonus/cleanliness | Chunk 4 |
| Scheduled auto-compute (cron) | Chunk 4b |
| Score display in SSO portal | Chunk 5 |
| MS Market Share Gain % (inactive) | Future |
