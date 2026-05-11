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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
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

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required", "MISSING_FIELDS", "")
		return
	}
	resp, err := h.Auth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token", "UNAUTHORIZED", "")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
