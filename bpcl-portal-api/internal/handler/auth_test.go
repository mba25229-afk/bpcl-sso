package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLoginHandler_Success(t *testing.T) {
	authSvc := &mockAuthSvc{}
	h := newHandler(authSvc, nil, nil, nil, nil, nil, nil)

	body := `{"employee_id":"EMP001","password":"secret123"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc := "DELHI-01"
	loginResp := &service.LoginResponse{
		Token: "tok.en.here",
		User:  &model.User{ID: uuid.New(), EmployeeID: "EMP001", TerritoryCode: &tc},
	}
	authSvc.On("Login", mock.Anything, "EMP001", "secret123").Return(loginResp, nil)

	h.Login(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "tok.en.here", got["token"])
}

func TestLoginHandler_WrongCredentials(t *testing.T) {
	authSvc := &mockAuthSvc{}
	h := newHandler(authSvc, nil, nil, nil, nil, nil, nil)

	body := `{"employee_id":"EMP001","password":"wrong"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authSvc.On("Login", mock.Anything, "EMP001", "wrong").Return(nil, model.ErrUnauthorized)

	h.Login(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_MissingFields(t *testing.T) {
	authSvc := &mockAuthSvc{}
	h := newHandler(authSvc, nil, nil, nil, nil, nil, nil)

	body := `{"employee_id":"EMP001"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
