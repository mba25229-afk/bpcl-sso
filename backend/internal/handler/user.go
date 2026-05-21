package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := service.ExtractClaims(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", "")
		return
	}
	if model.UserRole(claims.Role) != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "admin role required", "FORBIDDEN", "")
		return
	}

	q := r.URL.Query()
	role := q.Get("role")
	territory := q.Get("territory")
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	resp, err := h.Users.ListUsers(ctx, service.ListUsersParams{
		Role:      role,
		Territory: territory,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := service.ExtractClaims(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", "")
		return
	}
	if model.UserRole(claims.Role) != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "admin role required", "FORBIDDEN", "")
		return
	}

	var input service.CreateUserInput
	if err := decodeJSON(r.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_BODY", "")
		return
	}

	user, err := h.Users.CreateUser(ctx, input)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id", "INVALID_ID", "")
		return
	}

	claims, err := service.ExtractClaims(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", "")
		return
	}
	if model.UserRole(claims.Role) != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "admin role required", "FORBIDDEN", "")
		return
	}

	var input service.UpdateUserInput
	if err := decodeJSON(r.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_BODY", "")
		return
	}

	user, err := h.Users.UpdateUser(ctx, id, input)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id", "INVALID_ID", "")
		return
	}

	claims, err := service.ExtractClaims(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", "")
		return
	}
	if model.UserRole(claims.Role) != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "admin role required", "FORBIDDEN", "")
		return
	}

	var input service.ResetPasswordInput
	if err := decodeJSON(r.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_BODY", "")
		return
	}

	if err := h.Users.ResetPassword(ctx, id, input); err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func getClaims(ctx context.Context) (*service.Claims, error) {
	key := struct{}{}
	val := ctx.Value(key)
	if val == nil {
		return nil, model.ErrUnauthorized
	}
	if claims, ok := val.(*service.Claims); ok {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid claims type")
}

func decodeJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
