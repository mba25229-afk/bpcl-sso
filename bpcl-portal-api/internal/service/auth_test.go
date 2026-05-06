package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newAuthSvc(t *testing.T, users *mockUserRepo) *service.AuthService {
	t.Helper()
	return service.NewAuthService(users, "test-secret-32-chars-long-enough!", 24*time.Hour)
}

func makeUser(t *testing.T, password string) *model.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	tc := "DELHI-01"
	return &model.User{
		ID:           uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		EmployeeID:   "EMP001",
		Email:        "emp@example.com",
		Name:         "Test User",
		PasswordHash: string(hash),
		Role:         model.RoleTerritoryManager,
		TerritoryCode: &tc,
		IsActive:     true,
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	users := &mockUserRepo{}
	svc := newAuthSvc(t, users)
	user := makeUser(t, "secret123")

	users.On("GetByEmployeeID", context.Background(), "EMP001").Return(user, nil)
	users.On("UpdateLastLogin", context.Background(), user.ID).Return(nil)

	resp, err := svc.Login(context.Background(), "EMP001", "secret123")
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, user.ID, resp.User.ID)

	users.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	users := &mockUserRepo{}
	svc := newAuthSvc(t, users)
	user := makeUser(t, "secret123")

	users.On("GetByEmployeeID", context.Background(), "EMP001").Return(user, nil)

	_, err := svc.Login(context.Background(), "EMP001", "wrongpassword")
	assert.ErrorIs(t, err, model.ErrUnauthorized)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	users := &mockUserRepo{}
	svc := newAuthSvc(t, users)

	users.On("GetByEmployeeID", context.Background(), "NOTEXIST").Return(nil, model.ErrNotFound)

	_, err := svc.Login(context.Background(), "NOTEXIST", "any")
	assert.ErrorIs(t, err, model.ErrUnauthorized)
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	users := &mockUserRepo{}
	svc := newAuthSvc(t, users)
	user := makeUser(t, "secret123")
	user.IsActive = false

	users.On("GetByEmployeeID", context.Background(), "EMP001").Return(user, nil)

	_, err := svc.Login(context.Background(), "EMP001", "secret123")
	assert.ErrorIs(t, err, model.ErrUnauthorized)
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	users := &mockUserRepo{}
	svc := newAuthSvc(t, users)
	user := makeUser(t, "secret123")

	users.On("GetByEmployeeID", context.Background(), "EMP001").Return(user, nil)
	users.On("UpdateLastLogin", context.Background(), user.ID).Return(nil)

	resp, err := svc.Login(context.Background(), "EMP001", "secret123")
	require.NoError(t, err)

	claims, err := svc.ValidateToken(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, string(user.Role), claims.Role)
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	users := &mockUserRepo{}
	svc := newAuthSvc(t, users)

	_, err := svc.ValidateToken("not.a.token")
	assert.ErrorIs(t, err, model.ErrUnauthorized)
}
