package handler

import (
	"net/http"

	"github.com/bpcl/portal-api/internal/service"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := service.ExtractClaims(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":        claims.UserID,
		"role":           claims.Role,
		"territory_code": claims.TerritoryCode,
	})
}
