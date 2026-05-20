package ingest

import (
	"context"
	"encoding/json"
	"net/http"
)

// ServiceInterface allows handler tests to inject a mock without a DB.
type ServiceInterface interface {
	IngestDaily(ctx context.Context, req BulkDailyRequest) (*IngestResult, error)
	IngestMAKGE(ctx context.Context, req MAKGERequest) (*IngestResult, error)
	IngestGoogleRating(ctx context.Context, req GoogleRatingRequest) (*IngestResult, error)
	IngestTargets(ctx context.Context, req BulkTargetsRequest) (*IngestResult, error)
}

type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler {
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
