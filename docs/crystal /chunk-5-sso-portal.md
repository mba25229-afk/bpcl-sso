# Chunk 5 — SSO Portal / Dealer Report Card
**Status:** READY AFTER CHUNK 4  
**Stack:** Go (backend), Next.js App Router (frontend)  
**Depends on:** Chunk 3 (scoring engine ✓), Chunk 4 (admin portal ✓)  
**Blocks:** Chunk 6 (Dashboard)

---

## What This Chunk Delivers

Per-dealer read-only view. A dealer logs in, sees their own scorecard for the active competition month — scores per metric, rank among 40 peers, MTD actuals vs targets, day-wise trends, MAK GE log, Google rating.

One new Go package. Two new endpoints. Two Next.js pages.

---

## Pages

```
/portal
└── /portal/[cc_code]    dealer report card (SSO-protected)
```

---

## Backend — New Go Endpoints

### internal/portal/types.go

```go
package portal

import "time"

// ScoreCard is the full report card for one dealer for one month.
type ScoreCard struct {
    Dealer      DealerInfo    `json:"dealer"`
    MonthYear   string        `json:"month_year"`
    AsOf        string        `json:"as_of"`
    TotalMarks  float64       `json:"total_marks"`
    MaxPossible float64       `json:"max_possible"`  // sum of active max_marks
    RankOverall int           `json:"rank_overall"`
    TotalDealers int          `json:"total_dealers"`
    Metrics     []MetricCard  `json:"metrics"`
    DailyTrend  []DailyPoint  `json:"daily_trend"`   // last 30 days: ms + hsd KL
    MAKGELog    []MAKGEEntry  `json:"mak_ge_log"`
    GoogleLog   []GoogleEntry `json:"google_log"`
}

type DealerInfo struct {
    CCCode string `json:"cc_code"`
    ROName string `json:"ro_name"`
    Area   string `json:"area"`
}

type MetricCard struct {
    MetricKey    string   `json:"metric_key"`
    DisplayName  string   `json:"display_name"`
    ActualValue  *float64 `json:"actual_value"`   // nil if not scored
    Target       *float64 `json:"target"`         // nil if no target set
    AchievePct   *float64 `json:"achieve_pct"`    // actual/target × 100
    RankInGroup  *int     `json:"rank_in_group"`  // nil for manual/grade metrics
    MarksScored  *float64 `json:"marks_scored"`   // nil if inactive
    MaxMarks     float64  `json:"max_marks"`
    Unit         string   `json:"unit"`           // "KL" | "count" | "%" | "pts"
    Trend        string   `json:"trend"`          // "up" | "down" | "neutral" | ""
}

type DailyPoint struct {
    Date  string  `json:"date"`
    MSKL  float64 `json:"ms_kl"`
    HSDKL float64 `json:"hsd_kl"`
    QOC   int     `json:"qoc"`
    UFill int     `json:"ufill"`
}

type MAKGEEntry struct {
    ReadingDate  string  `json:"reading_date"`
    MeterReading float64 `json:"meter_reading"`
    Delta        float64 `json:"delta"` // vs previous reading
}

type GoogleEntry struct {
    SnapshotDate string  `json:"snapshot_date"`
    Rating       float64 `json:"rating"`
    ReviewCount  int     `json:"review_count"`
}
```

---

### internal/portal/repository.go

```go
package portal

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// GetScoreCard fetches all data for a dealer's report card in 4 queries.
func (r *Repository) GetScoreCard(ctx context.Context, ccCode string, monthYear time.Time) (*ScoreCard, error) {
    card := &ScoreCard{}

    // 1. Dealer info
    err := r.pool.QueryRow(ctx,
        `SELECT cc_code, ro_name, area FROM dealers WHERE cc_code = $1`, ccCode,
    ).Scan(&card.Dealer.CCCode, &card.Dealer.ROName, &card.Dealer.Area)
    if err != nil {
        return nil, err
    }

    // 2. Scores + metric metadata + targets in one JOIN
    rows, err := r.pool.Query(ctx, `
        SELECT
            ds.metric_key,
            sp.display_name,
            ds.actual_value,
            ds.rank_in_group,
            ds.marks_scored,
            ds.max_marks,
            -- target value per metric (mapped from monthly_targets columns)
            CASE ds.metric_key
                WHEN 'ms_absolute_vol'  THEN t.ms_kl
                WHEN 'hsd_absolute_vol' THEN t.hsd_kl
                WHEN 'speed_vol'        THEN t.speed_kl
                WHEN 'ufill_txns'       THEN t.ufill_target::numeric
                WHEN 'oil_change_count' THEN t.qoc_target::numeric
                WHEN 'mak_ge_sales'     THEN t.mak_ge_target
                ELSE NULL
            END AS target_value
        FROM dealer_scores ds
        JOIN scoring_params sp
            ON sp.metric_key = ds.metric_key AND sp.month_year = ds.month_year
        LEFT JOIN monthly_targets t
            ON t.cc_code = ds.cc_code AND t.month_year = ds.month_year
        WHERE ds.cc_code = $1 AND ds.month_year = $2
        ORDER BY sp.sort_order`,
        ccCode, monthYear,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    unitMap := map[string]string{
        "ms_absolute_vol": "KL", "ms_growth_pct": "%",
        "hsd_absolute_vol": "KL", "hsd_growth_pct": "%",
        "oil_change_count": "count", "mak_ge_sales": "KL",
        "ufill_txns": "count", "speed_vol": "KL",
        "cleanliness_audit": "grade", "sangam_certs": "count",
        "google_rating": "stars", "ips_pct": "%", "bonus": "pts",
    }

    var totalMarks, maxPossible float64
    for rows.Next() {
        var m MetricCard
        var target *float64
        if err := rows.Scan(
            &m.MetricKey, &m.DisplayName,
            &m.ActualValue, &m.RankInGroup, &m.MarksScored, &m.MaxMarks,
            &target,
        ); err != nil {
            return nil, err
        }
        m.Target = target
        if target != nil && *target > 0 && m.ActualValue != nil {
            pct := (*m.ActualValue / *target) * 100
            m.AchievePct = &pct
        }
        m.Unit = unitMap[m.MetricKey]
        if m.MarksScored != nil {
            totalMarks += *m.MarksScored
            maxPossible += m.MaxMarks
        }
        card.Metrics = append(card.Metrics, m)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    card.TotalMarks = totalMarks
    card.MaxPossible = maxPossible

    // 3. Overall rank from view
    r.pool.QueryRow(ctx,
        `SELECT rank_overall, COUNT(*) OVER ()
         FROM dealer_total_scores
         WHERE cc_code = $1 AND month_year = $2`,
        ccCode, monthYear,
    ).Scan(&card.RankOverall, &card.TotalDealers)

    // 4. Daily trend (all metrics, current month)
    dRows, err := r.pool.Query(ctx, `
        WITH dates AS (
            SELECT generate_series(
                DATE_TRUNC('month', $2::date),
                $2::date,
                '1 day'::interval
            )::date AS d
        )
        SELECT
            dates.d::text,
            COALESCE(ms.kl,0) + COALESCE(sp.kl,0) AS ms_kl,
            COALESCE(hsd.kl,0) AS hsd_kl,
            COALESCE(qoc.count,0) AS qoc,
            COALESCE(uf.count,0)  AS ufill
        FROM dates
        LEFT JOIN daily_ms    ms  ON ms.cc_code = $1 AND ms.txn_date = dates.d
        LEFT JOIN daily_speed sp  ON sp.cc_code = $1 AND sp.txn_date = dates.d
        LEFT JOIN daily_hsd   hsd ON hsd.cc_code = $1 AND hsd.txn_date = dates.d
        LEFT JOIN daily_qoc   qoc ON qoc.cc_code = $1 AND qoc.txn_date = dates.d
        LEFT JOIN daily_ufill uf  ON uf.cc_code  = $1 AND uf.txn_date = dates.d
        ORDER BY dates.d`,
        ccCode, monthYear,
    )
    if err != nil {
        return nil, err
    }
    defer dRows.Close()
    for dRows.Next() {
        var p DailyPoint
        if err := dRows.Scan(&p.Date, &p.MSKL, &p.HSDKL, &p.QOC, &p.UFill); err != nil {
            return nil, err
        }
        card.DailyTrend = append(card.DailyTrend, p)
    }

    // 5. MAK GE log
    gRows, err := r.pool.Query(ctx, `
        SELECT reading_date::text, meter_reading,
               meter_reading - LAG(meter_reading) OVER (ORDER BY reading_date) AS delta
        FROM mak_ge_readings
        WHERE cc_code = $1 AND DATE_TRUNC('month', reading_date) = $2
        ORDER BY reading_date`,
        ccCode, monthYear,
    )
    if err != nil {
        return nil, err
    }
    defer gRows.Close()
    for gRows.Next() {
        var e MAKGEEntry
        if err := gRows.Scan(&e.ReadingDate, &e.MeterReading, &e.Delta); err != nil {
            return nil, err
        }
        card.MAKGELog = append(card.MAKGELog, e)
    }

    // 6. Google rating log
    grRows, err := r.pool.Query(ctx, `
        SELECT snapshot_date::text, rating, review_count
        FROM google_ratings
        WHERE cc_code = $1 AND DATE_TRUNC('month', snapshot_date) = $2
        ORDER BY snapshot_date`,
        ccCode, monthYear,
    )
    if err != nil {
        return nil, err
    }
    defer grRows.Close()
    for grRows.Next() {
        var e GoogleEntry
        if err := grRows.Scan(&e.SnapshotDate, &e.Rating, &e.ReviewCount); err != nil {
            return nil, err
        }
        card.GoogleLog = append(card.GoogleLog, e)
    }

    card.MonthYear = monthYear.Format("2006-01")
    return card, nil
}
```

---

### internal/portal/handler.go

```go
package portal

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
)

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

// GET /api/v1/portal/:cc_code/scorecard?month=2026-05-01
func (h *Handler) ScoreCard(w http.ResponseWriter, r *http.Request) {
    ccCode := chi.URLParam(r, "cc_code")

    monthStr := r.URL.Query().Get("month")
    var monthYear time.Time
    var err error
    if monthStr == "" {
        // default to active competition period's month
        monthYear = time.Now().UTC().Truncate(24 * time.Hour)
        monthYear = time.Date(monthYear.Year(), monthYear.Month(), 1, 0, 0, 0, 0, time.UTC)
    } else {
        monthYear, err = time.Parse("2006-01-02", monthStr)
        if err != nil {
            writeError(w, 400, "invalid month"); return
        }
    }

    card, err := h.repo.GetScoreCard(r.Context(), ccCode, monthYear)
    if err != nil {
        writeError(w, 500, err.Error()); return
    }
    writeJSON(w, 200, card)
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

**Router addition:**
```go
portalRepo    := portal.NewRepository(pool)
portalHandler := portal.NewHandler(portalRepo)

r.Get("/api/v1/portal/{cc_code}/scorecard", portalHandler.ScoreCard)
```

---

## Frontend — Next.js Pages

### Page: `/portal/[cc_code]`

**Data:** `GET /api/v1/portal/:cc_code/scorecard?month=`

```
ScoreCardPage
├── Header bar (bg: #003D66)
│   ├── BPCL gold circle mark
│   ├── "[RO Name] — [CC Code]"
│   └── MonthSelector (dropdown)
│
├── HeroRow (3 KPI cards)
│   ├── TotalScore: "[X.XX] / [max]"  bg: #007BC9, color: #FFF
│   ├── Rank: "#[N] of 40"            bg: #003D66, color: #FFF
│   └── Achievement %: "[X]%"         bg: green/amber/red based on %
│
├── MetricsGrid (2-column)
│   └── MetricCard (per active metric)
│       ├── header: display_name + unit
│       ├── actual value (large, #007BC9)
│       ├── target value (small, #6B6B7B) if available
│       ├── achieve % bar: fill=#007BC9, track=#EBF5FD
│       ├── rank badge: "#N"
│       └── marks: "[X.XX] / [max_marks]" pts
│           color: ≥70% → #00875A | 40–70% → #FFB800 | <40% → #CC3333
│
├── DailyTrendChart (line chart — last 30 days)
│   ├── MS KL line: #007BC9
│   ├── HSD KL line: #FFB800
│   └── x-axis: dates, y-axis: KL
│
├── MAKGETable (if data exists)
│   ├── columns: Date | Reading | Delta
│   └── delta color: positive=#00875A, zero=#6B6B7B
│
└── GoogleRatingLog (if data exists)
    ├── columns: Date | Rating | Reviews
    └── star display for rating
```

### Achievement % color thresholds

| % | Background | Text color | Meaning |
|---|---|---|---|
| ≥ 90% | `#DAFBE1` | `#00875A` | On track |
| 70–89% | `#FFF3B0` | `#C4A800` | Watch |
| < 70% | `#FFEAE6` | `#CC3333` | Behind |

Apply to both the hero achievement card and per-metric marks color.

---

## Negative Score Display

Growth metrics can return negative `marks_scored`. Show these as:
- Value: `−2.34 pts` in `#CC3333`
- Rank badge: `#N` with `bg: #FFEAE6`
- No achievement bar (replace with "Below average growth" label in `#CC3333`)

---

## SSO Auth Note

The `/portal/[cc_code]` page sits behind SSO. The logged-in dealer should only see their own cc_code. Enforce at middleware level:

```typescript
// middleware.ts
if (session.cc_code !== params.cc_code && !session.is_admin) {
    redirect('/portal/' + session.cc_code)
}
```

---

## Design Tokens Applied

| Element | Token |
|---|---|
| Page bg | `#F5F5FA` |
| Header bar | `#003D66` |
| Total score card | `#007BC9` |
| Rank card | `#003D66` |
| Metric card bg | `#FFFFFF` |
| Achievement bar fill | `#007BC9` |
| Achievement bar track | `#EBF5FD` |
| On-track marks | `#00875A` |
| Watch marks | `#FFB800` |
| Behind marks | `#CC3333` |
| MS trend line | `#007BC9` |
| HSD trend line | `#FFB800` |
| MAK GE positive delta | `#00875A` |

---

## Acceptance Criteria

- [ ] `GET /api/v1/portal/112458/scorecard?month=2026-05-01` → returns ScoreCard JSON
- [ ] TotalMarks matches `dealer_total_scores` view for same dealer/month
- [ ] RankOverall matches view rank
- [ ] Dealer with no data for a metric: `marks_scored = null`, card shows "—"
- [ ] Negative growth marks: shown in red, not as negative bar
- [ ] Daily trend chart: no data days show 0, not gaps
- [ ] Non-owner dealer trying to access another dealer's cc_code → redirect
- [ ] Page loads in under 1s (single API call)
