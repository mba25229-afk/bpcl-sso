package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
)

func (h *Handler) GetOutlet(w http.ResponseWriter, r *http.Request) {
	cc := r.PathValue("cc")
	if cc == "" {
		writeError(w, http.StatusUnprocessableEntity, "cc path param required", "MISSING_CC", "")
		return
	}
	userID := userIDFromCtx(r)
	outlet, err := h.Outlet.GetOutlet(r.Context(), cc, userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, outlet)
}

func (h *Handler) ListOutlets(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)
	outlets, err := h.Outlet.ListOutlets(r.Context(), userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, outlets)
}

func (h *Handler) GetTerritorySummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = time.Now().Format("2006-01")
	}

	summary, err := h.Outlet.GetTerritorySummary(r.Context(), period)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) GetTerritoryOutlets(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = time.Now().Format("2006-01")
	}

	summary, err := h.Outlet.GetTerritorySummary(r.Context(), period)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// userIDFromCtx extracts the user ID from the JWT claims in the context.
func userIDFromCtx(r *http.Request) uuid.UUID {
	claims, err := service.ExtractClaims(r.Context())
	if err != nil {
		return uuid.Nil
	}
	return claims.UserID
}

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found", "NOT_FOUND", "")
	case errors.Is(err, model.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "FORBIDDEN", "")
	case errors.Is(err, model.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", "")
	default:
		writeError(w, http.StatusInternalServerError, "internal error", "INTERNAL", "")
	}
}
