package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bpcl/portal-api/internal/handler"
	"github.com/bpcl/portal-api/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOTPSvc struct{ mock.Mock }

func (m *mockOTPSvc) GenerateAndSend(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockOTPSvc) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	return m.Called(ctx, email, otp, newPassword).Error(0)
}

func TestForgotPassword_MissingEmail(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ForgotPassword(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestForgotPassword_Success(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	otp.On("GenerateAndSend", mock.Anything, "emp@bpcl.com").Return(nil)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ForgotPassword(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResetForgottenPassword_MissingFields(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ResetForgottenPassword(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestResetForgottenPassword_InvalidOTP(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	otp.On("ResetPassword", mock.Anything, "emp@bpcl.com", "000000", "newpass123").
		Return(model.ErrUnauthorized)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com","otp":"000000","new_password":"newpass123"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ResetForgottenPassword(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestResetForgottenPassword_Success(t *testing.T) {
	otp := &mockOTPSvc{}
	h := &handler.Handler{OTP: otp}

	otp.On("ResetPassword", mock.Anything, "emp@bpcl.com", "123456", "newpass123").Return(nil)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password",
		bytes.NewBufferString(`{"email":"emp@bpcl.com","otp":"123456","new_password":"newpass123"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ResetForgottenPassword(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}
