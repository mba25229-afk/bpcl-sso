package service_test

import (
	"context"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ctxWithClaims(role model.UserRole, territoryCode *string) context.Context {
	id := uuid.New()
	claims := &service.Claims{
		UserID:        id,
		Role:          string(role),
		TerritoryCode: territoryCode,
	}
	return context.WithValue(context.Background(), service.ClaimsKey, claims)
}

func strPtr(s string) *string { return &s }

func TestOutletService_GetOutlet_Success(t *testing.T) {
	outlets := &mockOutletRepo{}
	audit := &mockAuditRepo{}
	svc := service.NewOutletService(outlets, audit)

	tc := "DELHI-01"
	outlet := &model.RetailOutlet{CCNumber: "123456", Name: "Test Outlet", TerritoryCode: &tc, OutletType: "regular"}
	userID := uuid.New()
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)

	outlets.On("GetByCC", ctx, "123456").Return(outlet, nil)
	audit.On("Log", ctx, mock.Anything, "view_outlet", "123456", mock.Anything).Return(nil)

	got, err := svc.GetOutlet(ctx, "123456", userID)
	require.NoError(t, err)
	assert.Equal(t, "Test Outlet", got.Name)
}

func TestOutletService_GetOutlet_Forbidden(t *testing.T) {
	outlets := &mockOutletRepo{}
	audit := &mockAuditRepo{}
	svc := service.NewOutletService(outlets, audit)

	tc := "DELHI-02"
	outlet := &model.RetailOutlet{CCNumber: "123456", Name: "Other Outlet", TerritoryCode: strPtr("DELHI-01"), OutletType: "regular"}
	userID := uuid.New()
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)

	outlets.On("GetByCC", ctx, "123456").Return(outlet, nil)

	_, err := svc.GetOutlet(ctx, "123456", userID)
	assert.ErrorIs(t, err, model.ErrForbidden)
}

func TestOutletService_GetOutlet_AdminCanSeeAll(t *testing.T) {
	outlets := &mockOutletRepo{}
	audit := &mockAuditRepo{}
	svc := service.NewOutletService(outlets, audit)

	outlet := &model.RetailOutlet{CCNumber: "123456", Name: "Any Outlet", TerritoryCode: strPtr("DELHI-01"), OutletType: "regular"}
	userID := uuid.New()
	ctx := ctxWithClaims(model.RoleAdmin, nil)

	outlets.On("GetByCC", ctx, "123456").Return(outlet, nil)
	audit.On("Log", ctx, mock.Anything, "view_outlet", "123456", mock.Anything).Return(nil)

	got, err := svc.GetOutlet(ctx, "123456", userID)
	require.NoError(t, err)
	assert.Equal(t, "Any Outlet", got.Name)
}

func TestOutletService_GetOutlet_NotFound(t *testing.T) {
	outlets := &mockOutletRepo{}
	audit := &mockAuditRepo{}
	svc := service.NewOutletService(outlets, audit)

	ctx := ctxWithClaims(model.RoleAdmin, nil)
	outlets.On("GetByCC", ctx, "XXXXX").Return(nil, model.ErrNotFound)

	_, err := svc.GetOutlet(ctx, "XXXXX", uuid.New())
	assert.ErrorIs(t, err, model.ErrNotFound)
}
