package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HandlerRepoInterface is what the handler needs (subset of Repository).
type HandlerRepoInterface interface {
	GetDashboard(ctx context.Context, monthYear time.Time) (*DashboardResponse, error)
}

type Handler struct{ repo HandlerRepoInterface }

func NewHandler(repo HandlerRepoInterface) *Handler { return &Handler{repo: repo} }

// GET /api/v1/dashboard?month=2026-05-01
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	monthYear, err := parseMonth(r.URL.Query().Get("month"))
	if err != nil {
		writeError(w, 400, "invalid month")
		return
	}
	data, err := h.repo.GetDashboard(r.Context(), monthYear)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, data)
}

// GET /api/v1/dashboard/export?month=2026-05-01
// Export needs the full Repository (for ExportXLSX). Wire directly in router.
func ExportHandler(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		monthYear, err := parseMonth(r.URL.Query().Get("month"))
		if err != nil {
			writeError(w, 400, "invalid month")
			return
		}
		if err := repo.ExportXLSX(r.Context(), w, monthYear); err != nil {
			writeError(w, 500, err.Error())
		}
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
