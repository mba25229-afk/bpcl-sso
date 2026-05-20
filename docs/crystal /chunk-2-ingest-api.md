# Chunk 2 — Ingest API (Go)
**Status:** READY TO BUILD  
**Stack:** Go, pgx/v5, chi router  
**Depends on:** Chunk 1 migration applied ✓  
**Blocks:** Chunk 3 (Scoring Engine)

---

## Directory Layout

```
bpcl-sso/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── db/
│   │   └── db.go                  -- pgx pool setup
│   ├── dealer/
│   │   └── repository.go          -- dealer lookup / validation cache
│   ├── ingest/
│   │   ├── handler.go             -- HTTP handlers
│   │   ├── service.go             -- validation + business rules
│   │   ├── repository.go          -- DB writes (all upsert)
│   │   └── types.go               -- request/response structs
│   └── period/
│       └── repository.go          -- active competition period cache
├── go.mod
└── go.sum
```

---

## go.mod

```go
module github.com/bpcl/sso-backend

go 1.22

require (
    github.com/go-chi/chi/v5       v5.1.0
    github.com/jackc/pgx/v5        v5.5.5
    github.com/google/uuid         v1.6.0
)
```

---

## internal/db/db.go

```go
package db

import (
    "context"
    "fmt"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context) (*pgxpool.Pool, error) {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        return nil, fmt.Errorf("DATABASE_URL not set")
    }
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("parse db config: %w", err)
    }
    cfg.MaxConns = 20
    pool, err := pgxpool.NewWithConfig(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("open pool: %w", err)
    }
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping db: %w", err)
    }
    return pool, nil
}
```

---

## internal/ingest/types.go

```go
package ingest

import "time"

// Metric keys — matches scoring_params.metric_key and table names
type MetricKey string

const (
    MetricUfill    MetricKey = "ufill"
    MetricQOC      MetricKey = "qoc"
    MetricMS       MetricKey = "ms"
    MetricHSD      MetricKey = "hsd"
    MetricSpeed    MetricKey = "speed"
)

// DailyRow is one dealer's value for one day.
type DailyRow struct {
    CCCode string  `json:"cc_code"`
    Value  float64 `json:"value"` // count for ufill/qoc, KL for ms/hsd/speed
}

// BulkDailyRequest is the body for POST /ingest/daily-bulk
type BulkDailyRequest struct {
    Date   string     `json:"date"`   // "2026-05-07"
    Metric MetricKey  `json:"metric"` // "ufill" | "qoc" | "ms" | "hsd" | "speed"
    Rows   []DailyRow `json:"rows"`
}

// MAKGERequest is the body for POST /ingest/mak-ge
type MAKGERequest struct {
    CCCode        string  `json:"cc_code"`
    ReadingDate   string  `json:"reading_date"` // "2026-05-08"
    MeterReading  float64 `json:"meter_reading"`
}

// GoogleRatingRequest is the body for POST /ingest/google-rating
type GoogleRatingRequest struct {
    CCCode       string  `json:"cc_code"`
    SnapshotDate string  `json:"snapshot_date"` // "2026-05-08"
    Rating       float64 `json:"rating"`
    ReviewCount  int     `json:"review_count"`
}

// TargetRow is one dealer's monthly targets
type TargetRow struct {
    CCCode        string   `json:"cc_code"`
    UfillTarget   *int     `json:"ufill_target"`
    QOCTarget     *int     `json:"qoc_target"`
    SpeedKL       *float64 `json:"speed_kl"`
    MSKL          *float64 `json:"ms_kl"`
    HSDKL         *float64 `json:"hsd_kl"`
    MSLY          *float64 `json:"ms_ly"`
    HSDLY         *float64 `json:"hsd_ly"`
    DSWAvailable  *bool    `json:"dsw_available"`
    Nitrogen      *bool    `json:"nitrogen"`
    MAKGETarget   *float64 `json:"mak_ge_target"`
    DarpanTarget  *int     `json:"darpan_target"`
    CoolantLubeKL *float64 `json:"coolant_lube_kl"`
    Remarks       *string  `json:"remarks"`
}

// BulkTargetsRequest is the body for POST /targets/bulk
type BulkTargetsRequest struct {
    MonthYear string      `json:"month_year"` // "2026-05-01"
    Rows      []TargetRow `json:"rows"`
}

// IngestResult summarises what happened
type IngestResult struct {
    Inserted int      `json:"inserted"`
    Updated  int      `json:"updated"`
    Skipped  int      `json:"skipped"`
    Errors   []string `json:"errors,omitempty"`
}

// Period is a loaded competition period
type Period struct {
    MonthYear  time.Time
    TotalSlots int
    IsActive   bool
}
```

---

## internal/period/repository.go

```go
package period

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
    pool  *pgxpool.Pool
    mu    sync.RWMutex
    cache *activePeriod
}

type activePeriod struct {
    MonthYear  time.Time
    TotalSlots int
    loadedAt   time.Time
}

func New(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

// Active returns the current active competition period.
// Cached for 60s — admin changes take effect within a minute.
func (r *Repository) Active(ctx context.Context) (time.Time, int, error) {
    r.mu.RLock()
    if r.cache != nil && time.Since(r.cache.loadedAt) < 60*time.Second {
        t, s := r.cache.MonthYear, r.cache.TotalSlots
        r.mu.RUnlock()
        return t, s, nil
    }
    r.mu.RUnlock()

    var monthYear time.Time
    var totalSlots int
    err := r.pool.QueryRow(ctx,
        `SELECT month_year, total_slots FROM competition_periods WHERE is_active = TRUE LIMIT 1`,
    ).Scan(&monthYear, &totalSlots)
    if err != nil {
        return time.Time{}, 0, fmt.Errorf("no active competition period: %w", err)
    }

    r.mu.Lock()
    r.cache = &activePeriod{MonthYear: monthYear, TotalSlots: totalSlots, loadedAt: time.Now()}
    r.mu.Unlock()

    return monthYear, totalSlots, nil
}

// DateInActivePeriod returns true if date falls within the active period's calendar month.
func (r *Repository) DateInActivePeriod(ctx context.Context, d time.Time) (bool, error) {
    monthYear, _, err := r.Active(ctx)
    if err != nil {
        return false, err
    }
    return d.Year() == monthYear.Year() && d.Month() == monthYear.Month(), nil
}
```

---

## internal/dealer/repository.go

```go
package dealer

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
    pool     *pgxpool.Pool
    mu       sync.RWMutex
    cache    map[string]bool // cc_code -> active
    loadedAt time.Time
}

func New(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool, cache: make(map[string]bool)}
}

// IsActive returns true if cc_code exists and is active.
// Cache TTL: 5 minutes (dealer list changes rarely).
func (r *Repository) IsActive(ctx context.Context, ccCode string) (bool, error) {
    r.mu.RLock()
    if time.Since(r.loadedAt) < 5*time.Minute {
        active, ok := r.cache[ccCode]
        r.mu.RUnlock()
        if ok {
            return active, nil
        }
        return false, nil
    }
    r.mu.RUnlock()

    // Reload full cache
    rows, err := r.pool.Query(ctx, `SELECT cc_code, is_active FROM dealers`)
    if err != nil {
        return false, fmt.Errorf("load dealers: %w", err)
    }
    defer rows.Close()

    fresh := make(map[string]bool)
    for rows.Next() {
        var cc string
        var active bool
        if err := rows.Scan(&cc, &active); err != nil {
            return false, err
        }
        fresh[cc] = active
    }

    r.mu.Lock()
    r.cache = fresh
    r.loadedAt = time.Now()
    r.mu.Unlock()

    return fresh[ccCode], nil
}

// ValidateCodes returns a map of invalid cc_codes from the given slice.
func (r *Repository) ValidateCodes(ctx context.Context, codes []string) (map[string]bool, error) {
    invalid := make(map[string]bool)
    for _, c := range codes {
        ok, err := r.IsActive(ctx, c)
        if err != nil {
            return nil, err
        }
        if !ok {
            invalid[c] = true
        }
    }
    return invalid, nil
}
```

---

## internal/ingest/repository.go

```go
package ingest

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

// UpsertDaily writes daily rows into the appropriate table.
// Uses UPSERT — duplicate (cc_code, txn_date) updates the value.
func (r *Repository) UpsertDaily(ctx context.Context, metric MetricKey, date time.Time, rows []DailyRow) (inserted, updated int, err error) {
    table, col := dailyTable(metric)

    batch := &pgx.Batch{}
    sql := fmt.Sprintf(`
        INSERT INTO %s (cc_code, txn_date, %s)
        VALUES ($1, $2, $3)
        ON CONFLICT (cc_code, txn_date) DO UPDATE SET %s = EXCLUDED.%s
    `, table, col, col, col)

    for _, row := range rows {
        batch.Queue(sql, row.CCCode, date, row.Value)
    }

    br := r.pool.SendBatch(ctx, batch)
    defer br.Close()

    for range rows {
        tag, err := br.Exec()
        if err != nil {
            return inserted, updated, fmt.Errorf("upsert %s: %w", table, err)
        }
        switch tag.String() {
        case "INSERT 0 1":
            inserted++
        default:
            updated++
        }
    }
    return inserted, updated, nil
}

// UpsertMAKGE inserts or updates a single MAK GE reading.
func (r *Repository) UpsertMAKGE(ctx context.Context, req MAKGERequest, date time.Time) (inserted bool, err error) {
    sql := `
        INSERT INTO mak_ge_readings (cc_code, reading_date, meter_reading)
        VALUES ($1, $2, $3)
        ON CONFLICT (cc_code, reading_date) DO UPDATE SET meter_reading = EXCLUDED.meter_reading
        RETURNING (xmax = 0) AS was_inserted
    `
    err = r.pool.QueryRow(ctx, sql, req.CCCode, date, req.MeterReading).Scan(&inserted)
    return inserted, err
}

// UpsertGoogleRating inserts or updates a Google rating snapshot.
func (r *Repository) UpsertGoogleRating(ctx context.Context, req GoogleRatingRequest, date time.Time) (inserted bool, err error) {
    sql := `
        INSERT INTO google_ratings (cc_code, snapshot_date, rating, review_count)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (cc_code, snapshot_date) DO UPDATE
            SET rating = EXCLUDED.rating, review_count = EXCLUDED.review_count
        RETURNING (xmax = 0) AS was_inserted
    `
    err = r.pool.QueryRow(ctx, sql, req.CCCode, date, req.Rating, req.ReviewCount).Scan(&inserted)
    return inserted, err
}

// UpsertTargets bulk-upserts monthly targets for all dealers.
func (r *Repository) UpsertTargets(ctx context.Context, monthYear time.Time, rows []TargetRow) (inserted, updated int, err error) {
    sql := `
        INSERT INTO monthly_targets (
            cc_code, month_year,
            ufill_target, qoc_target, speed_kl, ms_kl, hsd_kl,
            ms_ly, hsd_ly, dsw_available, nitrogen,
            mak_ge_target, darpan_target, coolant_lube_kl, remarks
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
        )
        ON CONFLICT (cc_code, month_year) DO UPDATE SET
            ufill_target    = EXCLUDED.ufill_target,
            qoc_target      = EXCLUDED.qoc_target,
            speed_kl        = EXCLUDED.speed_kl,
            ms_kl           = EXCLUDED.ms_kl,
            hsd_kl          = EXCLUDED.hsd_kl,
            ms_ly           = EXCLUDED.ms_ly,
            hsd_ly          = EXCLUDED.hsd_ly,
            dsw_available   = EXCLUDED.dsw_available,
            nitrogen        = EXCLUDED.nitrogen,
            mak_ge_target   = EXCLUDED.mak_ge_target,
            darpan_target   = EXCLUDED.darpan_target,
            coolant_lube_kl = EXCLUDED.coolant_lube_kl,
            remarks         = EXCLUDED.remarks
    `

    batch := &pgx.Batch{}
    for _, t := range rows {
        batch.Queue(sql,
            t.CCCode, monthYear,
            t.UfillTarget, t.QOCTarget, t.SpeedKL, t.MSKL, t.HSDKL,
            t.MSLY, t.HSDLY, t.DSWAvailable, t.Nitrogen,
            t.MAKGETarget, t.DarpanTarget, t.CoolantLubeKL, t.Remarks,
        )
    }

    br := r.pool.SendBatch(ctx, batch)
    defer br.Close()

    for range rows {
        tag, e := br.Exec()
        if e != nil {
            return inserted, updated, fmt.Errorf("upsert target: %w", e)
        }
        if tag.String() == "INSERT 0 1" {
            inserted++
        } else {
            updated++
        }
    }
    return inserted, updated, nil
}

func dailyTable(m MetricKey) (table, col string) {
    switch m {
    case MetricUfill:
        return "daily_ufill", "count"
    case MetricQOC:
        return "daily_qoc", "count"
    case MetricMS:
        return "daily_ms", "kl"
    case MetricHSD:
        return "daily_hsd", "kl"
    case MetricSpeed:
        return "daily_speed", "kl"
    default:
        panic("unknown metric: " + string(m))
    }
}
```

---

## internal/ingest/service.go

Validation lives here. Handlers stay thin. Rules enforced at boundary — never let bad data into the repository layer.

```go
package ingest

import (
    "context"
    "fmt"
    "time"

    "github.com/bpcl/sso-backend/internal/dealer"
    "github.com/bpcl/sso-backend/internal/period"
)

type Service struct {
    repo    *Repository
    dealers *dealer.Repository
    periods *period.Repository
}

func NewService(repo *Repository, dealers *dealer.Repository, periods *period.Repository) *Service {
    return &Service{repo: repo, dealers: dealers, periods: periods}
}

// IngestDaily validates and writes one metric's daily data.
func (s *Service) IngestDaily(ctx context.Context, req BulkDailyRequest) (*IngestResult, error) {
    result := &IngestResult{}

    // Parse date
    date, err := time.Parse("2006-01-02", req.Date)
    if err != nil {
        return nil, fmt.Errorf("invalid date %q: use YYYY-MM-DD", req.Date)
    }

    // Date must be within active competition period
    inPeriod, err := s.periods.DateInActivePeriod(ctx, date)
    if err != nil {
        return nil, err
    }
    if !inPeriod {
        return nil, fmt.Errorf("date %s is not within the active competition period", req.Date)
    }

    // Validate metric key
    switch req.Metric {
    case MetricUfill, MetricQOC, MetricMS, MetricHSD, MetricSpeed:
    default:
        return nil, fmt.Errorf("unknown metric %q: must be one of ufill, qoc, ms, hsd, speed", req.Metric)
    }

    // Validate all cc_codes exist
    codes := make([]string, len(req.Rows))
    for i, r := range req.Rows {
        codes[i] = r.CCCode
    }
    invalid, err := s.dealers.ValidateCodes(ctx, codes)
    if err != nil {
        return nil, err
    }

    // Separate valid rows, collect invalid errors
    valid := make([]DailyRow, 0, len(req.Rows))
    for _, row := range req.Rows {
        if invalid[row.CCCode] {
            result.Errors = append(result.Errors, fmt.Sprintf("unknown cc_code: %s", row.CCCode))
            result.Skipped++
            continue
        }
        if row.Value < 0 {
            result.Errors = append(result.Errors, fmt.Sprintf("cc_code %s: value cannot be negative", row.CCCode))
            result.Skipped++
            continue
        }
        valid = append(valid, row)
    }

    if len(valid) == 0 {
        return result, nil
    }

    ins, upd, err := s.repo.UpsertDaily(ctx, req.Metric, date, valid)
    if err != nil {
        return nil, err
    }
    result.Inserted = ins
    result.Updated = upd
    return result, nil
}

// IngestMAKGE validates and writes a single MAK GE meter reading.
func (s *Service) IngestMAKGE(ctx context.Context, req MAKGERequest) (*IngestResult, error) {
    result := &IngestResult{}

    date, err := time.Parse("2006-01-02", req.ReadingDate)
    if err != nil {
        return nil, fmt.Errorf("invalid reading_date %q", req.ReadingDate)
    }
    inPeriod, err := s.periods.DateInActivePeriod(ctx, date)
    if err != nil {
        return nil, err
    }
    if !inPeriod {
        return nil, fmt.Errorf("reading_date %s is not within the active competition period", req.ReadingDate)
    }

    active, err := s.dealers.IsActive(ctx, req.CCCode)
    if err != nil {
        return nil, err
    }
    if !active {
        return nil, fmt.Errorf("unknown cc_code: %s", req.CCCode)
    }
    if req.MeterReading < 0 {
        return nil, fmt.Errorf("meter_reading cannot be negative")
    }

    inserted, err := s.repo.UpsertMAKGE(ctx, req, date)
    if err != nil {
        return nil, err
    }
    if inserted {
        result.Inserted = 1
    } else {
        result.Updated = 1
    }
    return result, nil
}

// IngestGoogleRating validates and writes a Google rating snapshot.
func (s *Service) IngestGoogleRating(ctx context.Context, req GoogleRatingRequest) (*IngestResult, error) {
    result := &IngestResult{}

    date, err := time.Parse("2006-01-02", req.SnapshotDate)
    if err != nil {
        return nil, fmt.Errorf("invalid snapshot_date %q", req.SnapshotDate)
    }
    inPeriod, err := s.periods.DateInActivePeriod(ctx, date)
    if err != nil {
        return nil, err
    }
    if !inPeriod {
        return nil, fmt.Errorf("snapshot_date %s is not within the active competition period", req.SnapshotDate)
    }

    active, err := s.dealers.IsActive(ctx, req.CCCode)
    if err != nil {
        return nil, err
    }
    if !active {
        return nil, fmt.Errorf("unknown cc_code: %s", req.CCCode)
    }
    if req.Rating < 1.0 || req.Rating > 5.0 {
        return nil, fmt.Errorf("rating must be between 1.0 and 5.0, got %.1f", req.Rating)
    }
    if req.ReviewCount < 0 {
        return nil, fmt.Errorf("review_count cannot be negative")
    }

    inserted, err := s.repo.UpsertGoogleRating(ctx, req, date)
    if err != nil {
        return nil, err
    }
    if inserted {
        result.Inserted = 1
    } else {
        result.Updated = 1
    }
    return result, nil
}

// IngestTargets validates and bulk-writes monthly targets.
func (s *Service) IngestTargets(ctx context.Context, req BulkTargetsRequest) (*IngestResult, error) {
    result := &IngestResult{}

    monthYear, err := time.Parse("2006-01-02", req.MonthYear)
    if err != nil {
        return nil, fmt.Errorf("invalid month_year %q: use YYYY-MM-01", req.MonthYear)
    }
    if monthYear.Day() != 1 {
        return nil, fmt.Errorf("month_year must be the first day of the month, e.g. 2026-05-01")
    }

    codes := make([]string, len(req.Rows))
    for i, r := range req.Rows {
        codes[i] = r.CCCode
    }
    invalid, err := s.dealers.ValidateCodes(ctx, codes)
    if err != nil {
        return nil, err
    }

    valid := make([]TargetRow, 0, len(req.Rows))
    for _, row := range req.Rows {
        if invalid[row.CCCode] {
            result.Errors = append(result.Errors, fmt.Sprintf("unknown cc_code: %s", row.CCCode))
            result.Skipped++
            continue
        }
        valid = append(valid, row)
    }

    if len(valid) == 0 {
        return result, nil
    }

    ins, upd, err := s.repo.UpsertTargets(ctx, monthYear, valid)
    if err != nil {
        return nil, err
    }
    result.Inserted = ins
    result.Updated = upd
    return result, nil
}
```

---

## internal/ingest/handler.go

```go
package ingest

import (
    "encoding/json"
    "net/http"
)

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

// POST /api/v1/ingest/daily-bulk
func (h *Handler) DailyBulk(w http.ResponseWriter, r *http.Request) {
    var req BulkDailyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
        return
    }
    if len(req.Rows) == 0 {
        writeError(w, http.StatusBadRequest, "rows cannot be empty")
        return
    }
    if len(req.Rows) > 100 {
        writeError(w, http.StatusBadRequest, "max 100 rows per request")
        return
    }

    result, err := h.svc.IngestDaily(r.Context(), req)
    if err != nil {
        writeError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, result)
}

// POST /api/v1/ingest/mak-ge
func (h *Handler) MAKGE(w http.ResponseWriter, r *http.Request) {
    var req MAKGERequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }

    result, err := h.svc.IngestMAKGE(r.Context(), req)
    if err != nil {
        writeError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, result)
}

// POST /api/v1/ingest/google-rating
func (h *Handler) GoogleRating(w http.ResponseWriter, r *http.Request) {
    var req GoogleRatingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }

    result, err := h.svc.IngestGoogleRating(r.Context(), req)
    if err != nil {
        writeError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, result)
}

// POST /api/v1/targets/bulk
func (h *Handler) TargetsBulk(w http.ResponseWriter, r *http.Request) {
    var req BulkTargetsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    if len(req.Rows) == 0 {
        writeError(w, http.StatusBadRequest, "rows cannot be empty")
        return
    }

    result, err := h.svc.IngestTargets(r.Context(), req)
    if err != nil {
        writeError(w, http.StatusUnprocessableEntity, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, result)
}

// --- helpers ---

type errorResponse struct {
    Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, errorResponse{Error: msg})
}
```

---

## cmd/api/main.go

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"

    "github.com/bpcl/sso-backend/internal/db"
    "github.com/bpcl/sso-backend/internal/dealer"
    "github.com/bpcl/sso-backend/internal/ingest"
    "github.com/bpcl/sso-backend/internal/period"
)

func main() {
    ctx := context.Background()

    pool, err := db.New(ctx)
    if err != nil {
        log.Fatalf("db: %v", err)
    }
    defer pool.Close()

    dealerRepo := dealer.New(pool)
    periodRepo := period.New(pool)
    ingestRepo := ingest.NewRepository(pool)
    ingestSvc  := ingest.NewService(ingestRepo, dealerRepo, periodRepo)
    ingestH    := ingest.NewHandler(ingestSvc)

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/ingest/daily-bulk",   ingestH.DailyBulk)
        r.Post("/ingest/mak-ge",       ingestH.MAKGE)
        r.Post("/ingest/google-rating", ingestH.GoogleRating)
        r.Post("/targets/bulk",        ingestH.TargetsBulk)
    })

    addr := ":" + getEnv("PORT", "8080")
    log.Printf("listening on %s", addr)
    if err := http.ListenAndServe(addr, r); err != nil {
        log.Fatalf("server: %v", err)
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

## API Reference

### POST /api/v1/ingest/daily-bulk

**One call per metric per day.** Max 100 rows. Speed + Speed100 must be summed client-side before sending.

```json
{
  "date": "2026-05-07",
  "metric": "ufill",
  "rows": [
    { "cc_code": "112385", "value": 52 },
    { "cc_code": "112386", "value": 46 }
  ]
}
```

```json
{ "inserted": 2, "updated": 0, "skipped": 0 }
```

---

### POST /api/v1/ingest/mak-ge

```json
{
  "cc_code": "112390",
  "reading_date": "2026-05-08",
  "meter_reading": 14523.50
}
```

---

### POST /api/v1/ingest/google-rating

```json
{
  "cc_code": "112461",
  "snapshot_date": "2026-05-08",
  "rating": 4.4,
  "review_count": 1469
}
```

---

### POST /api/v1/targets/bulk

```json
{
  "month_year": "2026-05-01",
  "rows": [
    {
      "cc_code": "112385",
      "ufill_target": 3000,
      "qoc_target": 100,
      "speed_kl": 20,
      "ms_kl": 456,
      "hsd_kl": 50,
      "ms_ly": 415,
      "hsd_ly": 65,
      "dsw_available": false,
      "nitrogen": false,
      "mak_ge_target": 200,
      "darpan_target": 60,
      "coolant_lube_kl": 100
    }
  ]
}
```

---

## Invariants Enforced (mechanically, not by convention)

| Rule | Enforced where |
|---|---|
| `cc_code` must exist in `dealers` | `service.go` → dealer cache lookup |
| `date` within active competition period | `service.go` → period cache lookup |
| `value` / `meter_reading` cannot be negative | `service.go` validation |
| `rating` between 1.0–5.0 | `service.go` validation |
| Duplicate date+cc_code = upsert, not error | `repository.go` → ON CONFLICT DO UPDATE |
| Speed + Speed100 combined before insert | Caller contract — documented in API ref above |
| Max 100 rows per bulk request | `handler.go` — prevents runaway payloads |

---

## Acceptance Criteria

- [ ] `go build ./...` — zero errors
- [ ] `go vet ./...` — zero warnings
- [ ] Seed 6 days of MAY_DATA via API — all rows insert
- [ ] Re-submit same 6 days — all rows update, zero errors
- [ ] Submit unknown cc_code — returns 422 with error message, valid rows still insert
- [ ] Submit date outside active period — returns 422
- [ ] MTD sums via DB query match MAY_DATA Excel MTD column ±0.01

---

## What Is NOT in This Chunk

| Feature | Chunk |
|---|---|
| Auth / JWT middleware | Chunk 2b (wire after ingest works) |
| Scoring engine | Chunk 3 |
| Admin portal UI | Chunk 4 |
| Read endpoints (GET actuals, GET dashboard) | Chunk 4 |
| Sangam data ingest | Chunk 4 |
