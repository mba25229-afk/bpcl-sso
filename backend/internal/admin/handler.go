package admin

import (
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct{ repo RepositoryInterface }

func NewHandler(repo RepositoryInterface) *Handler { return &Handler{repo: repo} }

// GET /api/v1/admin/dealers
func (h *Handler) ListDealers(w http.ResponseWriter, r *http.Request) {
	dealers, err := h.repo.ListDealers(r.Context())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, dealers)
}

// PUT /api/v1/admin/dealers/{cc_code}
func (h *Handler) ToggleDealer(w http.ResponseWriter, r *http.Request) {
	cc := r.PathValue("cc_code")
	var req ToggleDealerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	if err := h.repo.ToggleDealer(r.Context(), cc, req.IsActive); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// GET /api/v1/admin/competition-periods
func (h *Handler) ListPeriods(w http.ResponseWriter, r *http.Request) {
	periods, err := h.repo.ListPeriods(r.Context())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, periods)
}

// POST /api/v1/admin/competition-periods
func (h *Handler) CreatePeriod(w http.ResponseWriter, r *http.Request) {
	var req CreatePeriodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	monthYear, err := time.Parse("2006-01-02", req.MonthYear)
	if err != nil || monthYear.Day() != 1 {
		writeError(w, 400, "month_year must be YYYY-MM-01")
		return
	}
	period, err := h.repo.CreatePeriod(r.Context(), req.Name, monthYear)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, period)
}

// PATCH /api/v1/admin/competition-periods/{month_year}/activate
func (h *Handler) ActivatePeriod(w http.ResponseWriter, r *http.Request) {
	monthYear, err := time.Parse("2006-01-02", r.PathValue("month_year"))
	if err != nil {
		writeError(w, 400, "invalid month_year")
		return
	}
	if err := h.repo.ActivatePeriod(r.Context(), monthYear); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// GET /api/v1/admin/targets/{month_year}
func (h *Handler) GetTargets(w http.ResponseWriter, r *http.Request) {
	monthYear, err := time.Parse("2006-01-02", r.PathValue("month_year"))
	if err != nil {
		writeError(w, 400, "invalid month_year")
		return
	}
	rows, err := h.repo.GetTargets(r.Context(), monthYear)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, rows)
}

// GET /api/v1/admin/ingest/summary?month=2026-05-01&metric=ms
func (h *Handler) IngestSummary(w http.ResponseWriter, r *http.Request) {
	monthYear, err := time.Parse("2006-01-02", r.URL.Query().Get("month"))
	if err != nil {
		writeError(w, 400, "invalid month")
		return
	}

	metric := r.URL.Query().Get("metric")
	table, col, ok := MetricTable(metric)
	if !ok {
		writeError(w, 400, "unknown metric")
		return
	}

	rows, err := h.repo.GetIngestSummary(r.Context(), metric, table, col, monthYear)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, rows)
}

// GET /api/v1/admin/mak-ge/{month_year}
func (h *Handler) GetMAKGE(w http.ResponseWriter, r *http.Request) {
	monthYear, err := time.Parse("2006-01-02", r.PathValue("month_year"))
	if err != nil {
		writeError(w, 400, "invalid month_year")
		return
	}
	rows, err := h.repo.GetMAKGEReadings(r.Context(), monthYear)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, rows)
}

// GET /api/v1/admin/manual-scores/{month_year}
func (h *Handler) GetManualScores(w http.ResponseWriter, r *http.Request) {
	monthYear, err := time.Parse("2006-01-02", r.PathValue("month_year"))
	if err != nil {
		writeError(w, 400, "invalid month_year")
		return
	}
	rows, err := h.repo.GetManualScores(r.Context(), monthYear)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, rows)
}

// POST /api/v1/admin/manual-scores
func (h *Handler) SaveManualScores(w http.ResponseWriter, r *http.Request) {
	var req BulkManualScoresRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	monthYear, err := time.Parse("2006-01-02", req.MonthYear)
	if err != nil {
		writeError(w, 400, "invalid month_year")
		return
	}
	if err := h.repo.UpsertManualScores(r.Context(), monthYear, req.Rows); err != nil {
		writeError(w, 500, err.Error())
		return
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

type ETLTriggerResponse struct {
	Status  string `json:"status"`
	Detail  string `json:"detail,omitempty"`
	Message string `json:"message"`
}

func (h *Handler) TriggerETL(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, ETLTriggerResponse{
		Status:  "ok",
		Message: "ETL is managed via standalone binary. Use 'go run etl/cmd/etl/main.go' or set up cron.",
	})
}

func (h *Handler) GetETLStatus(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.GetLastETLRun(r.Context())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, rows)
}
