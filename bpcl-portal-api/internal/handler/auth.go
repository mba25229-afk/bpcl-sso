package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bpcl/portal-api/internal/model"
)

type loginRequest struct {
	EmployeeID string `json:"employee_id"`
	Password   string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "INVALID_BODY", "")
		return
	}
	if req.EmployeeID == "" || req.Password == "" {
		writeError(w, http.StatusUnprocessableEntity, "employee_id and password are required", "MISSING_FIELDS", "")
		return
	}

	resp, err := h.Auth.Login(r.Context(), req.EmployeeID, req.Password)
	if err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, "invalid credentials", "UNAUTHORIZED", "")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error", "INTERNAL", "")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
