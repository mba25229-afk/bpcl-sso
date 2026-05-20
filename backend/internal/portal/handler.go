package portal

import (
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct{ repo RepositoryInterface }

func NewHandler(repo RepositoryInterface) *Handler { return &Handler{repo: repo} }

// GET /api/v1/portal/{cc_code}/scorecard?month=2026-05-01
func (h *Handler) ScoreCard(w http.ResponseWriter, r *http.Request) {
	ccCode := r.PathValue("cc_code")

	monthStr := r.URL.Query().Get("month")
	var monthYear time.Time
	var err error
	if monthStr == "" {
		now := time.Now().UTC()
		monthYear = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		monthYear, err = time.Parse("2006-01-02", monthStr)
		if err != nil {
			writeError(w, 400, "invalid month")
			return
		}
	}

	card, err := h.repo.GetScoreCard(r.Context(), ccCode, monthYear)
	if err != nil {
		writeError(w, 500, err.Error())
		return
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
