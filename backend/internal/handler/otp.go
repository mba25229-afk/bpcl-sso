package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/bpcl/portal-api/internal/model"
)

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "INVALID_BODY", "")
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		writeError(w, http.StatusUnprocessableEntity, "email is required", "MISSING_FIELDS", "")
		return
	}
	_ = h.OTP.GenerateAndSend(r.Context(), req.Email)
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "If that email is registered, an OTP has been sent.",
	})
}

func (h *Handler) ResetForgottenPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON body", "INVALID_BODY", "")
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.OTP) == "" || strings.TrimSpace(req.NewPassword) == "" {
		writeError(w, http.StatusUnprocessableEntity, "email, otp, and new_password are required", "MISSING_FIELDS", "")
		return
	}
	if err := h.OTP.ResetPassword(r.Context(), req.Email, req.OTP, req.NewPassword); err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, "invalid or expired OTP", "UNAUTHORIZED", "")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_REQUEST", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Password reset successful."})
}
