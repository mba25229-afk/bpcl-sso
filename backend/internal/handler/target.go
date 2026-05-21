package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bpcl/portal-api/internal/service"
)

func (h *Handler) GetTargets(w http.ResponseWriter, r *http.Request) {
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
	resp, err := h.Target.GetTargets(r.Context(), cc, period, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) SetTargets(w http.ResponseWriter, r *http.Request) {
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

	var input service.TargetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "INVALID_BODY", "")
		return
	}

	userID := userIDFromCtx(r)
	resp, err := h.Target.SetTargets(r.Context(), cc, period, input, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
