package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type cronPingRequest struct {
	Status     string `json:"status"`      // "ok" | "error"
	Detail     string `json:"detail"`
	DurationMs int    `json:"duration_ms"`
	Source     string `json:"source"`
}

type cronStatusResponse struct {
	LastRunAt  *time.Time `json:"last_run_at"`
	Status     *string    `json:"status"`
	Detail     *string    `json:"detail"`
	DurationMs *int       `json:"duration_ms"`
}

// POST /api/v1/health/cron-ping
// Called by the ETL script after each run (success or failure).
// No auth — secured by network boundary / internal call only.
func (h *Handler) CronPing(w http.ResponseWriter, r *http.Request) {
	var req cronPingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	if req.Status != "ok" && req.Status != "error" {
		writeError(w, 400, "status must be 'ok' or 'error'")
		return
	}
	if req.Source == "" {
		req.Source = "cron"
	}

	_, err := h.pool.Exec(r.Context(),
		`INSERT INTO cr_etl_log (status, detail, duration_ms, source) VALUES ($1, $2, $3, $4)`,
		req.Status, req.Detail, req.DurationMs, req.Source,
	)
	if err != nil {
		writeError(w, 500, "failed to record ETL ping")
		return
	}
	writeJSON(w, 200, map[string]string{"recorded": "ok"})
}

// GET /api/v1/health/cron-status (auth required — wired in router)
func (h *Handler) CronStatus(w http.ResponseWriter, r *http.Request) {
	row := h.pool.QueryRow(r.Context(),
		`SELECT run_at, status, detail, duration_ms
         FROM cr_etl_log
         ORDER BY run_at DESC
         LIMIT 1`,
	)

	var resp cronStatusResponse
	var runAt time.Time
	var status, detail string
	var durationMs int
	if err := row.Scan(&runAt, &status, &detail, &durationMs); err != nil {
		// No rows yet — return empty object rather than 500
		writeJSON(w, 200, resp)
		return
	}
	resp.LastRunAt = &runAt
	resp.Status = &status
	resp.Detail = &detail
	resp.DurationMs = &durationMs
	writeJSON(w, 200, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Ping is exported so router.go can call pool.Ping for health checks.
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
