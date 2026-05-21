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
	MonthYear string `json:"month_year"`
	AsOf      string `json:"as_of"`
}

type ComputeResponse struct {
	RowsWritten int    `json:"rows_written"`
	MonthYear   string `json:"month_year"`
	AsOf        string `json:"as_of"`
}

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