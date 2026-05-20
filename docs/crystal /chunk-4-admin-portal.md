# Chunk 4 — Admin Portal
**Status:** READY TO BUILD  
**Stack:** Go (backend), Next.js App Router (frontend)  
**Depends on:** Chunk 1 ✓, Chunk 2 ✓, Chunk 3 ✓  
**Blocks:** Chunk 5 (SSO Portal), Chunk 6 (Dashboard)

---

## What This Chunk Delivers

Six admin pages. Six new Go endpoints (read-only + one write group). No scoring logic here — scoring lives in Chunk 3. Admin portal is purely data entry + competition management.

---

## Pages

```
/admin
├── /dealers              dealer master — active/inactive toggle
├── /competition          create & activate competition periods
├── /targets              monthly targets bulk entry (all 40 dealers)
├── /ingest               daily data entry — tabs per metric
├── /mak-ge               weekly meter readings
└── /manual-scores        bonus marks + cleanliness grade entry
```

Nav: `bg: #003D66`, logo left, page links, user badge right.  
All pages: `bg: #F5F5FA`, content cards `bg: #FFFFFF`, `border-radius: 10px`.

---

## Backend — New Go Endpoints

### internal/admin/types.go

```go
package admin

import "time"

type Dealer struct {
    CCCode      string `json:"cc_code"`
    ROName      string `json:"ro_name"`
    Area        string `json:"area"`
    IsActive    bool   `json:"is_active"`
    DealerEmail string `json:"dealer_email,omitempty"`
}

type CompetitionPeriod struct {
    ID         string    `json:"id"`
    Name       string    `json:"name"`
    MonthYear  time.Time `json:"month_year"`
    TotalSlots int       `json:"total_slots"`
    IsActive   bool      `json:"is_active"`
}

type CreatePeriodRequest struct {
    Name      string `json:"name"`       // "Boost and Win May 2026"
    MonthYear string `json:"month_year"` // "2026-05-01"
}

type ToggleDealerRequest struct {
    IsActive bool `json:"is_active"`
}

type TargetRow struct {
    CCCode        string   `json:"cc_code"`
    ROName        string   `json:"ro_name"`
    UfillTarget   *int     `json:"ufill_target"`
    QOCTarget     *int     `json:"qoc_target"`
    SpeedKL       *float64 `json:"speed_kl"`
    MSKL          *float64 `json:"ms_kl"`
    HSDKL         *float64 `json:"hsd_kl"`
    MSLY          *float64 `json:"ms_ly"`
    HSDLY         *float64 `json:"hsd_ly"`
    MAKGETarget   *float64 `json:"mak_ge_target"`
    DarpanTarget  *int     `json:"darpan_target"`
    CoolantLubeKL *float64 `json:"coolant_lube_kl"`
    DSWAvailable  *bool    `json:"dsw_available"`
    Nitrogen      *bool    `json:"nitrogen"`
}

type IngestSummaryRow struct {
    CCCode  string  `json:"cc_code"`
    ROName  string  `json:"ro_name"`
    MTDSum  float64 `json:"mtd_sum"`
    DaysFilled int  `json:"days_filled"`
}

type MAKGERow struct {
    CCCode       string  `json:"cc_code"`
    ROName       string  `json:"ro_name"`
    ReadingDate  string  `json:"reading_date"`
    MeterReading float64 `json:"meter_reading"`
}

type ManualScoreRow struct {
    CCCode           string   `json:"cc_code"`
    ROName           string   `json:"ro_name"`
    BonusMarks       *float64 `json:"bonus_marks"`
    CleanlinessGrade *string  `json:"cleanliness_grade"`
}

type BulkManualScoresRequest struct {
    MonthYear string           `json:"month_year"`
    Rows      []ManualScoreRow `json:"rows"`
}
```

---

### internal/admin/repository.go

```go
package admin

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) ListDealers(ctx context.Context) ([]Dealer, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT cc_code, ro_name, area, is_active, COALESCE(dealer_email,'')
         FROM dealers ORDER BY ro_name`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Dealer
    for rows.Next() {
        var d Dealer
        if err := rows.Scan(&d.CCCode, &d.ROName, &d.Area, &d.IsActive, &d.DealerEmail); err != nil {
            return nil, err
        }
        out = append(out, d)
    }
    return out, rows.Err()
}

func (r *Repository) ToggleDealer(ctx context.Context, ccCode string, active bool) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE dealers SET is_active = $1, updated_at = NOW() WHERE cc_code = $2`,
        active, ccCode)
    return err
}

func (r *Repository) ListPeriods(ctx context.Context) ([]CompetitionPeriod, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT id, name, month_year, total_slots, is_active
         FROM competition_periods ORDER BY month_year DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []CompetitionPeriod
    for rows.Next() {
        var p CompetitionPeriod
        if err := rows.Scan(&p.ID, &p.Name, &p.MonthYear, &p.TotalSlots, &p.IsActive); err != nil {
            return nil, err
        }
        out = append(out, p)
    }
    return out, rows.Err()
}

func (r *Repository) CreatePeriod(ctx context.Context, name string, monthYear time.Time) (*CompetitionPeriod, error) {
    var p CompetitionPeriod
    err := r.pool.QueryRow(ctx,
        `INSERT INTO competition_periods (name, month_year, total_slots, is_active)
         VALUES ($1, $2, 40, FALSE)
         RETURNING id, name, month_year, total_slots, is_active`,
        name, monthYear,
    ).Scan(&p.ID, &p.Name, &p.MonthYear, &p.TotalSlots, &p.IsActive)
    return &p, err
}

func (r *Repository) ActivatePeriod(ctx context.Context, monthYear time.Time) error {
    // Deactivate all, then activate target — in single transaction
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    if _, err := tx.Exec(ctx, `UPDATE competition_periods SET is_active = FALSE`); err != nil {
        return err
    }
    tag, err := tx.Exec(ctx,
        `UPDATE competition_periods SET is_active = TRUE WHERE month_year = $1`, monthYear)
    if err != nil {
        return err
    }
    if tag.RowsAffected() == 0 {
        return fmt.Errorf("period %s not found", monthYear.Format("2006-01"))
    }
    return tx.Commit(ctx)
}

func (r *Repository) GetTargets(ctx context.Context, monthYear time.Time) ([]TargetRow, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT d.cc_code, d.ro_name,
               t.ufill_target, t.qoc_target, t.speed_kl, t.ms_kl, t.hsd_kl,
               t.ms_ly, t.hsd_ly, t.mak_ge_target, t.darpan_target,
               t.coolant_lube_kl, t.dsw_available, t.nitrogen
        FROM dealers d
        LEFT JOIN monthly_targets t
            ON t.cc_code = d.cc_code AND t.month_year = $1
        WHERE d.is_active = TRUE
        ORDER BY d.ro_name`, monthYear)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []TargetRow
    for rows.Next() {
        var t TargetRow
        if err := rows.Scan(
            &t.CCCode, &t.ROName,
            &t.UfillTarget, &t.QOCTarget, &t.SpeedKL, &t.MSKL, &t.HSDKL,
            &t.MSLY, &t.HSDLY, &t.MAKGETarget, &t.DarpanTarget,
            &t.CoolantLubeKL, &t.DSWAvailable, &t.Nitrogen,
        ); err != nil {
            return nil, err
        }
        out = append(out, t)
    }
    return out, rows.Err()
}

func (r *Repository) GetIngestSummary(ctx context.Context, metric, table, col string, monthYear time.Time) ([]IngestSummaryRow, error) {
    sql := fmt.Sprintf(`
        SELECT d.cc_code, d.ro_name,
               COALESCE(SUM(s.%s), 0)   AS mtd_sum,
               COALESCE(COUNT(s.%s), 0) AS days_filled
        FROM dealers d
        LEFT JOIN %s s ON s.cc_code = d.cc_code
            AND DATE_TRUNC('month', s.txn_date) = $1
        WHERE d.is_active = TRUE
        GROUP BY d.cc_code, d.ro_name
        ORDER BY d.ro_name`, col, col, table)
    rows, err := r.pool.Query(ctx, sql, monthYear)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []IngestSummaryRow
    for rows.Next() {
        var s IngestSummaryRow
        if err := rows.Scan(&s.CCCode, &s.ROName, &s.MTDSum, &s.DaysFilled); err != nil {
            return nil, err
        }
        out = append(out, s)
    }
    return out, rows.Err()
}

func (r *Repository) GetMAKGEReadings(ctx context.Context, monthYear time.Time) ([]MAKGERow, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT d.cc_code, d.ro_name,
               COALESCE(g.reading_date::text, ''), COALESCE(g.meter_reading, 0)
        FROM dealers d
        LEFT JOIN mak_ge_readings g ON g.cc_code = d.cc_code
            AND DATE_TRUNC('month', g.reading_date) = $1
        WHERE d.is_active = TRUE
        ORDER BY d.ro_name, g.reading_date`, monthYear)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []MAKGERow
    for rows.Next() {
        var m MAKGERow
        if err := rows.Scan(&m.CCCode, &m.ROName, &m.ReadingDate, &m.MeterReading); err != nil {
            return nil, err
        }
        out = append(out, m)
    }
    return out, rows.Err()
}

func (r *Repository) GetManualScores(ctx context.Context, monthYear time.Time) ([]ManualScoreRow, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT d.cc_code, d.ro_name,
               m.bonus_marks, m.cleanliness_grade
        FROM dealers d
        LEFT JOIN dealer_manual_scores m
            ON m.cc_code = d.cc_code AND m.month_year = $1
        WHERE d.is_active = TRUE
        ORDER BY d.ro_name`, monthYear)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []ManualScoreRow
    for rows.Next() {
        var s ManualScoreRow
        if err := rows.Scan(&s.CCCode, &s.ROName, &s.BonusMarks, &s.CleanlinessGrade); err != nil {
            return nil, err
        }
        out = append(out, s)
    }
    return out, rows.Err()
}

func (r *Repository) UpsertManualScores(ctx context.Context, monthYear time.Time, rows []ManualScoreRow) error {
    batch := &pgx.Batch{}
    for _, row := range rows {
        batch.Queue(`
            INSERT INTO dealer_manual_scores (cc_code, month_year, bonus_marks, cleanliness_grade)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (cc_code, month_year) DO UPDATE SET
                bonus_marks       = EXCLUDED.bonus_marks,
                cleanliness_grade = EXCLUDED.cleanliness_grade,
                updated_at        = NOW()`,
            row.CCCode, monthYear, row.BonusMarks, row.CleanlinessGrade)
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
```

---

### internal/admin/handler.go

```go
package admin

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
)

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

// GET /api/v1/admin/dealers
func (h *Handler) ListDealers(w http.ResponseWriter, r *http.Request) {
    dealers, err := h.repo.ListDealers(r.Context())
    if err != nil { writeError(w, 500, err.Error()); return }
    writeJSON(w, 200, dealers)
}

// PUT /api/v1/admin/dealers/:cc_code
func (h *Handler) ToggleDealer(w http.ResponseWriter, r *http.Request) {
    cc := chi.URLParam(r, "cc_code")
    var req ToggleDealerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, 400, "invalid JSON"); return
    }
    if err := h.repo.ToggleDealer(r.Context(), cc, req.IsActive); err != nil {
        writeError(w, 500, err.Error()); return
    }
    writeJSON(w, 200, map[string]bool{"ok": true})
}

// GET /api/v1/admin/competition-periods
func (h *Handler) ListPeriods(w http.ResponseWriter, r *http.Request) {
    periods, err := h.repo.ListPeriods(r.Context())
    if err != nil { writeError(w, 500, err.Error()); return }
    writeJSON(w, 200, periods)
}

// POST /api/v1/admin/competition-periods
func (h *Handler) CreatePeriod(w http.ResponseWriter, r *http.Request) {
    var req CreatePeriodRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, 400, "invalid JSON"); return
    }
    monthYear, err := time.Parse("2006-01-02", req.MonthYear)
    if err != nil || monthYear.Day() != 1 {
        writeError(w, 400, "month_year must be YYYY-MM-01"); return
    }
    period, err := h.repo.CreatePeriod(r.Context(), req.Name, monthYear)
    if err != nil { writeError(w, 500, err.Error()); return }
    writeJSON(w, 201, period)
}

// PATCH /api/v1/admin/competition-periods/:month_year/activate
func (h *Handler) ActivatePeriod(w http.ResponseWriter, r *http.Request) {
    monthYear, err := time.Parse("2006-01-02", chi.URLParam(r, "month_year"))
    if err != nil { writeError(w, 400, "invalid month_year"); return }
    if err := h.repo.ActivatePeriod(r.Context(), monthYear); err != nil {
        writeError(w, 500, err.Error()); return
    }
    writeJSON(w, 200, map[string]bool{"ok": true})
}

// GET /api/v1/admin/targets/:month_year
func (h *Handler) GetTargets(w http.ResponseWriter, r *http.Request) {
    monthYear, err := time.Parse("2006-01-02", chi.URLParam(r, "month_year"))
    if err != nil { writeError(w, 400, "invalid month_year"); return }
    rows, err := h.repo.GetTargets(r.Context(), monthYear)
    if err != nil { writeError(w, 500, err.Error()); return }
    writeJSON(w, 200, rows)
}

// GET /api/v1/admin/ingest/summary?month=2026-05-01&metric=ms
func (h *Handler) IngestSummary(w http.ResponseWriter, r *http.Request) {
    monthYear, err := time.Parse("2006-01-02", r.URL.Query().Get("month"))
    if err != nil { writeError(w, 400, "invalid month"); return }

    type metricMeta struct{ table, col string }
    metricMap := map[string]metricMeta{
        "ms":    {"daily_ms",    "kl"},
        "hsd":   {"daily_hsd",   "kl"},
        "speed": {"daily_speed", "kl"},
        "ufill": {"daily_ufill", "count"},
        "qoc":   {"daily_qoc",   "count"},
    }
    metric := r.URL.Query().Get("metric")
    meta, ok := metricMap[metric]
    if !ok { writeError(w, 400, "unknown metric"); return }

    rows, err := h.repo.GetIngestSummary(r.Context(), metric, meta.table, meta.col, monthYear)
    if err != nil { writeError(w, 500, err.Error()); return }
    writeJSON(w, 200, rows)
}

// GET /api/v1/admin/mak-ge/:month_year
func (h *Handler) GetMAKGE(w http.ResponseWriter, r *http.Request) {
    monthYear, err := time.Parse("2006-01-02", chi.URLParam(r, "month_year"))
    if err != nil { writeError(w, 400, "invalid month_year"); return }
    rows, err := h.repo.GetMAKGEReadings(r.Context(), monthYear)
    if err != nil { writeError(w, 500, err.Error()); return }
    writeJSON(w, 200, rows)
}

// GET /api/v1/admin/manual-scores/:month_year
func (h *Handler) GetManualScores(w http.ResponseWriter, r *http.Request) {
    monthYear, err := time.Parse("2006-01-02", chi.URLParam(r, "month_year"))
    if err != nil { writeError(w, 400, "invalid month_year"); return }
    rows, err := h.repo.GetManualScores(r.Context(), monthYear)
    if err != nil { writeError(w, 500, err.Error()); return }
    writeJSON(w, 200, rows)
}

// POST /api/v1/admin/manual-scores
func (h *Handler) SaveManualScores(w http.ResponseWriter, r *http.Request) {
    var req BulkManualScoresRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, 400, "invalid JSON"); return
    }
    monthYear, err := time.Parse("2006-01-02", req.MonthYear)
    if err != nil { writeError(w, 400, "invalid month_year"); return }
    if err := h.repo.UpsertManualScores(r.Context(), monthYear, req.Rows); err != nil {
        writeError(w, 500, err.Error()); return
    }
    writeJSON(w, 200, map[string]bool{"ok": true})
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

**Router additions in main.go:**
```go
adminRepo    := admin.NewRepository(pool)
adminHandler := admin.NewHandler(adminRepo)

r.Route("/api/v1/admin", func(r chi.Router) {
    r.Get("/dealers",                              adminHandler.ListDealers)
    r.Put("/dealers/{cc_code}",                    adminHandler.ToggleDealer)
    r.Get("/competition-periods",                  adminHandler.ListPeriods)
    r.Post("/competition-periods",                 adminHandler.CreatePeriod)
    r.Patch("/competition-periods/{month_year}/activate", adminHandler.ActivatePeriod)
    r.Get("/targets/{month_year}",                 adminHandler.GetTargets)
    r.Get("/ingest/summary",                       adminHandler.IngestSummary)
    r.Get("/mak-ge/{month_year}",                  adminHandler.GetMAKGE)
    r.Get("/manual-scores/{month_year}",           adminHandler.GetManualScores)
    r.Post("/manual-scores",                       adminHandler.SaveManualScores)
})
```

---

## Frontend — Next.js Pages

### Layout: `app/admin/layout.tsx`

```
AdminLayout
├── Sidebar nav (fixed left, bg: #003D66)
│   ├── BPCL logo + gold circle mark
│   ├── NavLink: Dealers
│   ├── NavLink: Competition
│   ├── NavLink: Targets
│   ├── NavLink: Daily Ingest
│   ├── NavLink: MAK GE
│   └── NavLink: Manual Scores
└── Main content area (bg: #F5F5FA)
```

---

### Page: `/admin/dealers`

**Data:** `GET /api/v1/admin/dealers`

```
DealersPage
├── Header: "Dealer Master" + active/total count badge (bg: #007BC9)
├── SearchInput (filter by name/cc_code)
└── DealerTable
    ├── columns: CC Code | RO Name | Area | Status | Action
    ├── status chip: active = bg:#DAFBE1 text:#00875A | inactive = bg:#FFEAE6 text:#CC3333
    └── toggle switch → PUT /api/v1/admin/dealers/:cc_code
```

---

### Page: `/admin/competition`

**Data:** `GET /api/v1/admin/competition-periods`

```
CompetitionPage
├── CreatePeriodForm
│   ├── TextInput: name ("Boost and Win May 2026")
│   ├── DateInput: month_year (first of month picker)
│   └── Button: "Create Period" → POST /api/v1/admin/competition-periods
└── PeriodsTable
    ├── columns: Month | Name | Slots | Status | Action
    ├── active row: left border 3px #007BC9
    └── "Activate" button → PATCH .../activate
        confirmation modal: "This will deactivate all other periods."
```

---

### Page: `/admin/targets`

**Data:** `GET /api/v1/admin/targets/:month_year`  
**Write:** `POST /api/v1/targets/bulk` (Chunk 2 endpoint)

```
TargetsPage
├── MonthSelector (dropdown of competition periods)
├── TargetsGrid (40 rows × 10 columns — inline editable)
│   ├── sticky first column: RO Name + CC Code
│   ├── editable cells: UFILL | QOC | Speed KL | MS KL | HSD KL |
│   │                   MS LY | HSD LY | MAK GE | Darpan | Coolant
│   ├── unsaved cell: bg #FFF3B0 (light gold)
│   └── DSW / Nitrogen: checkbox columns
├── SaveButton (bottom sticky): "Save All Targets"
│   → POST /api/v1/targets/bulk with all rows
└── StatusBar: "X of 40 dealers have targets set"
```

Key UX: save entire grid in one POST. No row-by-row saves. The Chunk 2 bulk endpoint handles upsert.

---

### Page: `/admin/ingest`

**Data:** `GET /api/v1/admin/ingest/summary?month=&metric=`  
**Write:** `POST /api/v1/ingest/daily-bulk` (Chunk 2 endpoint)

```
IngestPage
├── MonthSelector + DatePicker (select date within active month)
├── MetricTabs: UFILL | QOC | MS | HSD | SPEED
└── ActiveTab
    ├── IngestSummaryTable (40 rows)
    │   ├── columns: RO Name | MTD Sum | Days Filled | Today's Value (editable)
    │   ├── MTD sum shown for context — read only
    │   └── "Today's Value" input: numeric, validates ≥ 0
    ├── BulkPasteButton: paste from Excel (parse tab-separated cc_code + value)
    └── SaveButton: "Submit [metric] for [date]"
        → POST /api/v1/ingest/daily-bulk
        → on success: reload MTD sums
```

Speed tab note: renders as single input per dealer (Speed + Speed100 already summed — per ms-aggregation.md).

---

### Page: `/admin/mak-ge`

**Data:** `GET /api/v1/admin/mak-ge/:month_year`  
**Write:** `POST /api/v1/ingest/mak-ge` (Chunk 2, one at a time)

```
MAKGEPage
├── MonthSelector
└── MAKGEGrid (40 dealers)
    ├── columns: RO Name | Reading 1 | Reading 2 | Reading 3 | Reading 4 | Delta
    ├── readings are date-stamped inputs (one per week)
    ├── Delta column: auto-calculated last - first, color:
    │   positive → #00875A, zero/null → #6B6B7B
    └── per-cell save on blur → POST /api/v1/ingest/mak-ge
```

---

### Page: `/admin/manual-scores`

**Data:** `GET /api/v1/admin/manual-scores/:month_year`  
**Write:** `POST /api/v1/admin/manual-scores`

```
ManualScoresPage
├── MonthSelector
├── "Trigger Scoring Run" button (top right)
│   → POST /api/v1/scoring/compute { month_year, as_of: today }
│   → shows spinner + "520 rows written" on success
└── ManualScoresGrid (40 rows)
    ├── columns: RO Name | Bonus Marks (0–10) | Cleanliness Grade
    ├── BonusMarks: numeric input, validates 0–10
    ├── CleanlinessGrade: select (Excellent/Good/Average/Below Average/Poor/—)
    └── SaveAll button → POST /api/v1/admin/manual-scores
```

---

## Design Tokens Applied

| Element | Token |
|---|---|
| Sidebar bg | `#003D66` |
| Active nav item | `#007BC9` left border |
| Page bg | `#F5F5FA` |
| Card bg | `#FFFFFF` |
| Primary button | `bg: #007BC9, color: #FFFFFF` |
| Unsaved cell | `bg: #FFF3B0` |
| Active status chip | `bg: #DAFBE1, color: #00875A` |
| Inactive status chip | `bg: #FFEAE6, color: #CC3333` |
| Table header | `bg: #003D66, color: #FFFFFF` |
| Alternating rows | `bg: #F5F5FA` |
| Border accent | `#FFE000` 3px for section dividers |

---

## Acceptance Criteria

- [ ] All 6 admin pages render without error
- [ ] Dealers page: toggle dealer inactive → reloads list, status chip updates
- [ ] Competition page: create May 2026 period → appears in list; activate → other periods show inactive
- [ ] Targets page: edit 3 cells → save → reload page → values persist
- [ ] Ingest page: enter UFill for 5 dealers for a date → MTD sums update on reload
- [ ] MAK GE page: enter 2 readings for one dealer → delta column shows correct value
- [ ] Manual scores: save bonus marks → trigger scoring run → 520 rows written toast
- [ ] No page makes more than 2 API calls on initial load
