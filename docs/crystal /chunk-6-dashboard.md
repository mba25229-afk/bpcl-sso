# Chunk 6 — Dashboard
**Status:** READY AFTER CHUNK 5  
**Stack:** Go (backend), Next.js App Router (frontend)  
**Depends on:** Chunk 5 (SSO Portal ✓)  
**Audience:** Area Manager — sees all 40 dealers ranked

---

## What This Chunk Delivers

Single page. All 40 dealers ranked by total score. Traffic lights. Metric breakdown. Drilldown to Chunk 5 report card. Excel export.

One new Go package. Two endpoints.

---

## Backend — New Go Endpoints

### internal/dashboard/types.go

```go
package dashboard

import "time"

type DashboardRow struct {
    CCCode       string   `json:"cc_code"`
    ROName       string   `json:"ro_name"`
    Rank         int      `json:"rank"`
    TotalMarks   float64  `json:"total_marks"`
    MaxPossible  float64  `json:"max_possible"`
    AchievePct   float64  `json:"achieve_pct"` // total_marks / max_possible × 100
    Status       string   `json:"status"`       // "green" | "amber" | "red"
    MetricBreakdown []MetricBreakdownItem `json:"metric_breakdown"`
}

type MetricBreakdownItem struct {
    MetricKey   string   `json:"metric_key"`
    DisplayName string   `json:"display_name"`
    MarksScored *float64 `json:"marks_scored"`
    MaxMarks    float64  `json:"max_marks"`
    Rank        *int     `json:"rank"`
}

type DashboardResponse struct {
    MonthYear  string         `json:"month_year"`
    ComputedAt time.Time      `json:"computed_at"`
    Dealers    []DashboardRow `json:"dealers"`
    Summary    DashboardSummary `json:"summary"`
}

type DashboardSummary struct {
    TotalDealers int     `json:"total_dealers"`
    GreenCount   int     `json:"green_count"`   // ≥90%
    AmberCount   int     `json:"amber_count"`   // 70–89%
    RedCount     int     `json:"red_count"`     // <70%
    AvgScore     float64 `json:"avg_score"`
    TopScore     float64 `json:"top_score"`
}
```

---

### internal/dashboard/repository.go

```go
package dashboard

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) GetDashboard(ctx context.Context, monthYear time.Time) (*DashboardResponse, error) {
    // 1. Totals + rank from view
    rows, err := r.pool.Query(ctx, `
        SELECT
            dts.cc_code,
            d.ro_name,
            dts.rank_overall,
            dts.total_marks,
            SUM(sp.max_marks) FILTER (WHERE sp.is_active) AS max_possible,
            MAX(dts.computed_at)
        FROM dealer_total_scores dts
        JOIN dealers d ON d.cc_code = dts.cc_code
        JOIN scoring_params sp ON sp.month_year = dts.month_year
        WHERE dts.month_year = $1
        GROUP BY dts.cc_code, d.ro_name, dts.rank_overall, dts.total_marks, dts.computed_at
        ORDER BY dts.rank_overall`,
        monthYear,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    resp := &DashboardResponse{MonthYear: monthYear.Format("2006-01")}
    ccCodes := []string{}

    rowMap := map[string]*DashboardRow{}
    for rows.Next() {
        var dr DashboardRow
        var computedAt time.Time
        if err := rows.Scan(
            &dr.CCCode, &dr.ROName, &dr.Rank,
            &dr.TotalMarks, &dr.MaxPossible, &computedAt,
        ); err != nil {
            return nil, err
        }
        if dr.MaxPossible > 0 {
            dr.AchievePct = (dr.TotalMarks / dr.MaxPossible) * 100
        }
        dr.Status = trafficLight(dr.AchievePct)
        resp.Dealers = append(resp.Dealers, dr)
        rowMap[dr.CCCode] = &resp.Dealers[len(resp.Dealers)-1]
        ccCodes = append(ccCodes, dr.CCCode)
        if computedAt.After(resp.ComputedAt) {
            resp.ComputedAt = computedAt
        }
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }

    // 2. Per-metric breakdown for all dealers (one query)
    mRows, err := r.pool.Query(ctx, `
        SELECT ds.cc_code, ds.metric_key, sp.display_name,
               ds.marks_scored, ds.max_marks, ds.rank_in_group
        FROM dealer_scores ds
        JOIN scoring_params sp
            ON sp.metric_key = ds.metric_key AND sp.month_year = ds.month_year
        WHERE ds.month_year = $1 AND sp.is_active = TRUE
        ORDER BY ds.cc_code, sp.sort_order`,
        monthYear,
    )
    if err != nil {
        return nil, err
    }
    defer mRows.Close()

    for mRows.Next() {
        var ccCode string
        var item MetricBreakdownItem
        if err := mRows.Scan(
            &ccCode, &item.MetricKey, &item.DisplayName,
            &item.MarksScored, &item.MaxMarks, &item.Rank,
        ); err != nil {
            return nil, err
        }
        if dr, ok := rowMap[ccCode]; ok {
            dr.MetricBreakdown = append(dr.MetricBreakdown, item)
        }
    }

    // 3. Summary stats
    resp.Summary = buildSummary(resp.Dealers)
    return resp, nil
}

func trafficLight(pct float64) string {
    switch {
    case pct >= 90:
        return "green"
    case pct >= 70:
        return "amber"
    default:
        return "red"
    }
}

func buildSummary(dealers []DashboardRow) DashboardSummary {
    s := DashboardSummary{TotalDealers: len(dealers)}
    var total float64
    for _, d := range dealers {
        switch d.Status {
        case "green":
            s.GreenCount++
        case "amber":
            s.AmberCount++
        case "red":
            s.RedCount++
        }
        total += d.TotalMarks
        if d.TotalMarks > s.TopScore {
            s.TopScore = d.TotalMarks
        }
    }
    if s.TotalDealers > 0 {
        s.AvgScore = total / float64(s.TotalDealers)
    }
    return s
}
```

---

### internal/dashboard/export.go

Excel export using `github.com/xuri/excelize/v2`.

```go
package dashboard

import (
    "context"
    "net/http"
    "time"

    "github.com/xuri/excelize/v2"
)

// ExportXLSX writes the dashboard as an Excel file to the response writer.
func (r *Repository) ExportXLSX(ctx context.Context, w http.ResponseWriter, monthYear time.Time) error {
    data, err := r.GetDashboard(ctx, monthYear)
    if err != nil {
        return err
    }

    f := excelize.NewFile()
    sheet := "Dealer Rankings"
    f.SetSheetName("Sheet1", sheet)

    // Header row styling: #003D66 bg, white text
    headerStyle, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"003D66"}, Pattern: 1},
        Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
    })
    greenStyle, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"DAFBE1"}, Pattern: 1},
    })
    amberStyle, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"FFF3B0"}, Pattern: 1},
    })
    redStyle, _ := f.NewStyle(&excelize.Style{
        Fill: excelize.Fill{Type: "pattern", Color: []string{"FFEAE6"}, Pattern: 1},
    })

    // Headers
    headers := []string{"Rank", "CC Code", "RO Name", "Total Marks", "Max Possible", "Achievement %", "Status"}
    for i, h := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue(sheet, cell, h)
        f.SetCellStyle(sheet, cell, cell, headerStyle)
    }

    // Data rows
    for i, d := range data.Dealers {
        row := i + 2
        vals := []interface{}{
            d.Rank, d.CCCode, d.ROName,
            d.TotalMarks, d.MaxPossible,
            d.AchievePct,
            map[string]string{"green": "On Track", "amber": "Watch", "red": "Behind"}[d.Status],
        }
        rowStyle := map[string]int{"green": greenStyle, "amber": amberStyle, "red": redStyle}[d.Status]
        for j, v := range vals {
            cell, _ := excelize.CoordinatesToCellName(j+1, row)
            f.SetCellValue(sheet, cell, v)
            f.SetCellStyle(sheet, cell, cell, rowStyle)
        }
    }

    // Column widths
    f.SetColWidth(sheet, "A", "A", 6)
    f.SetColWidth(sheet, "B", "B", 12)
    f.SetColWidth(sheet, "C", "C", 30)
    f.SetColWidth(sheet, "D", "F", 14)

    filename := "BPCL_Rankings_" + monthYear.Format("2006_01") + ".xlsx"
    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition", "attachment; filename="+filename)
    return f.Write(w)
}
```

**go.mod addition:** `github.com/xuri/excelize/v2 v2.8.1`

---

### internal/dashboard/handler.go

```go
package dashboard

import (
    "encoding/json"
    "net/http"
    "time"
)

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

// GET /api/v1/dashboard?month=2026-05-01
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    monthYear, err := parseMonth(r.URL.Query().Get("month"))
    if err != nil {
        writeError(w, 400, "invalid month"); return
    }
    data, err := h.repo.GetDashboard(r.Context(), monthYear)
    if err != nil {
        writeError(w, 500, err.Error()); return
    }
    writeJSON(w, 200, data)
}

// GET /api/v1/dashboard/export?month=2026-05-01
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
    monthYear, err := parseMonth(r.URL.Query().Get("month"))
    if err != nil {
        writeError(w, 400, "invalid month"); return
    }
    if err := h.repo.ExportXLSX(r.Context(), w, monthYear); err != nil {
        writeError(w, 500, err.Error())
    }
}

func parseMonth(s string) (time.Time, error) {
    if s == "" {
        now := time.Now().UTC()
        return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
    }
    return time.Parse("2006-01-02", s)
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

**Router additions:**
```go
dashRepo    := dashboard.NewRepository(pool)
dashHandler := dashboard.NewHandler(dashRepo)

r.Get("/api/v1/dashboard",        dashHandler.GetDashboard)
r.Get("/api/v1/dashboard/export", dashHandler.Export)
```

---

## Frontend — `/dashboard`

**Data:** `GET /api/v1/dashboard?month=`

```
DashboardPage
├── Header (bg: #003D66)
│   ├── "Central Delhi — [Month Year]"
│   ├── MonthSelector dropdown
│   └── ExportButton: "Export Excel" → GET /api/v1/dashboard/export?month=
│
├── SummaryStrip (3 stat chips)
│   ├── Green chip: "[N] On Track ≥90%"   bg:#DAFBE1 text:#00875A
│   ├── Amber chip: "[N] Watch 70–89%"    bg:#FFF3B0 text:#C4A800
│   └── Red chip:   "[N] Behind <70%"     bg:#FFEAE6 text:#CC3333
│
├── RankingsTable
│   ├── sticky header: bg:#003D66 text:#FFF
│   ├── columns:
│   │   Rank | RO Name | Score | Achievement% | [metric columns...] | Action
│   ├── traffic light left border per row:
│   │   green→#00875A | amber→#FFB800 | red→#CC3333
│   ├── achievement % pill: colored by threshold
│   ├── metric columns: marks/max_marks for each active metric
│   │   (collapsed by default, expandable via "Show Detail" toggle)
│   └── Action: "View →" → /portal/[cc_code]
│
└── LastUpdated: "Scores computed at [timestamp]" in #6B6B7B
```

### Metric Columns (collapsed by default)

Show 3 summary columns by default:
- Score (total_marks)
- Achievement %
- Rank

Toggle "Show Detail" expands to all 13 active metric columns. Each cell:
- Value: `marks_scored / max_marks`
- Color: top-third `#007BC9`, mid-third `#FFB800`, bottom-third `#CC3333`

### Row Coloring

```
status === 'green' → left border 3px #00875A, bg on hover #DAFBE1
status === 'amber' → left border 3px #FFB800, bg on hover #FFF3B0
status === 'red'   → left border 3px #CC3333, bg on hover #FFEAE6
```

---

## Design Tokens Applied

| Element | Token |
|---|---|
| Page header | `#003D66` |
| Page bg | `#F5F5FA` |
| Table header | `#003D66` + white text |
| Alternating rows | `#F5F5FA` |
| Green status | `#00875A` / `#DAFBE1` |
| Amber status | `#FFB800` / `#FFF3B0` |
| Red status | `#CC3333` / `#FFEAE6` |
| Top metric marks | `#007BC9` |
| Export button | `bg:#FFE000, color:#003D66` (gold CTA) |
| Last updated | `#6B6B7B` |

---

## Acceptance Criteria

- [ ] `GET /api/v1/dashboard?month=2026-05-01` returns all 40 dealers ranked
- [ ] Rank 1 dealer has highest `total_marks`
- [ ] `summary.green_count + amber_count + red_count = total_dealers`
- [ ] Export endpoint returns valid `.xlsx` file with BPCL blue headers
- [ ] Exported file: rank order matches API response
- [ ] Dashboard page: clicking "View →" navigates to `/portal/[cc_code]`
- [ ] MonthSelector: changing month refetches and re-ranks
- [ ] `computed_at` timestamp visible — area manager knows if scores are stale
- [ ] Page renders in under 1.5s with 40 dealers of data
