package service_test

import (
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTargetService_GetTargets_Success(t *testing.T) {
	outlets := &mockOutletRepo{}
	tgts := &mockTargetRepo{}
	audit := &mockAuditRepo{}
	svc := service.NewTargetService(outlets, tgts, audit)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	outlet := &model.RetailOutlet{CCNumber: "123456", TerritoryCode: &tc, OutletType: "regular"}
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)

	targets := []*model.Target{
		{ProductID: 1, TargetValue: 500},
		{ProductID: 2, TargetValue: 300},
	}
	outlets.On("GetByCC", ctx, "123456").Return(outlet, nil)
	tgts.On("GetByPeriod", ctx, "123456", period).Return(targets, nil)

	resp, err := svc.GetTargets(ctx, "123456", period, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "2026-03", resp.Period)
	assert.InDelta(t, 500.0, resp.Fuel["MS"], 0.001)
	assert.InDelta(t, 300.0, resp.Fuel["HSD"], 0.001)
}

func TestTargetService_SetTargets_ManagerCanSet(t *testing.T) {
	outlets := &mockOutletRepo{}
	tgts := &mockTargetRepo{}
	audit := &mockAuditRepo{}
	svc := service.NewTargetService(outlets, tgts, audit)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	outlet := &model.RetailOutlet{CCNumber: "123456", TerritoryCode: &tc, OutletType: "regular"}
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)
	userID := uuid.New()

	input := service.TargetInput{
		Fuel:    map[string]float64{"MS": 600, "HSD": 400},
		NonFuel: map[string]float64{"QOC": 50},
	}
	afterTargets := []*model.Target{
		{ProductID: 1, TargetValue: 600},
		{ProductID: 2, TargetValue: 400},
		{ProductID: 4, TargetValue: 50},
	}

	outlets.On("GetByCC", ctx, "123456").Return(outlet, nil)
	tgts.On("UpsertBatch", ctx, mock.Anything).Return(nil)
	tgts.On("GetByPeriod", ctx, "123456", period).Return(afterTargets, nil)
	audit.On("Log", ctx, userID, "set_target", "123456", mock.Anything).Return(nil)

	resp, err := svc.SetTargets(ctx, "123456", period, input, userID)
	require.NoError(t, err)
	assert.Equal(t, "2026-03", resp.Period)
	assert.InDelta(t, 600.0, resp.Fuel["MS"], 0.001)
}

func TestTargetService_SetTargets_ROManagerForbidden(t *testing.T) {
	outlets := &mockOutletRepo{}
	tgts := &mockTargetRepo{}
	audit := &mockAuditRepo{}
	svc := service.NewTargetService(outlets, tgts, audit)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	ctx := ctxWithClaims(model.RoleROManager, &tc)

	_, err := svc.SetTargets(ctx, "123456", period, service.TargetInput{}, uuid.New())
	assert.ErrorIs(t, err, model.ErrForbidden)
}
