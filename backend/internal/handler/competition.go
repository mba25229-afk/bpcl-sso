package handler

import (
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	territory := r.URL.Query().Get("territory")
	if territory == "" {
		writeError(w, http.StatusUnprocessableEntity, "territory query param required", "MISSING_TERRITORY", "")
		return
	}
	userID := userIDFromCtx(r)
	resp, err := h.Competition.GetLeaderboard(r.Context(), territory, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetDealerScorecard(w http.ResponseWriter, r *http.Request) {
	cc := r.PathValue("cc")
	if cc == "" {
		writeError(w, http.StatusUnprocessableEntity, "cc path param required", "MISSING_CC", "")
		return
	}

	competitionIDStr := r.PathValue("competition_id")
	competitionID, err := uuid.Parse(competitionIDStr)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid competition_id", "INVALID_UUID", "")
		return
	}

	userID := userIDFromCtx(r)
	resp, err := h.Competition.GetDealerScorecard(r.Context(), cc, competitionID, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
