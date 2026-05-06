package handler

import (
	"fmt"
	"net/http"
	"time"
)

func (h *Handler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	cc := r.PathValue("cc")
	if cc == "" {
		writeError(w, http.StatusUnprocessableEntity, "cc path param required", "MISSING_CC", "")
		return
	}

	period, err := parsePeriod(r.URL.Query().Get("period"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "period must be YYYY-MM", "INVALID_PERIOD", "")
		return
	}

	userID := userIDFromCtx(r)
	resp, err := h.Performance.GetPerformance(r.Context(), cc, period, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	cc := r.PathValue("cc")
	if cc == "" {
		writeError(w, http.StatusUnprocessableEntity, "cc path param required", "MISSING_CC", "")
		return
	}

	q := r.URL.Query()
	from, err := parsePeriod(q.Get("from"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "from must be YYYY-MM", "INVALID_FROM", "")
		return
	}
	to, err := parsePeriod(q.Get("to"))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "to must be YYYY-MM", "INVALID_TO", "")
		return
	}

	userID := userIDFromCtx(r)
	resp, err := h.Performance.GetAnalysis(r.Context(), cc, from, to, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// parsePeriod parses a "YYYY-MM" string into the first day of that month.
func parsePeriod(s string) (time.Time, error) {
	return time.Parse("2006-01", s)
}

func (h *Handler) GetTrend(w http.ResponseWriter, r *http.Request) {
	cc := r.PathValue("cc")
	if cc == "" {
		writeError(w, http.StatusUnprocessableEntity, "cc path param required", "MISSING_CC", "")
		return
	}

	q := r.URL.Query()
	months := 6
	if m := q.Get("months"); m != "" {
		if _, err := fmt.Sscanf(m, "%d", &months); err != nil || months < 1 || months > 24 {
			writeError(w, http.StatusUnprocessableEntity, "months must be 1-24", "INVALID_MONTHS", "")
			return
		}
	}

	resp, err := h.Performance.GetTrend(r.Context(), cc, months)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
