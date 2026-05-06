package service_test

import (
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformanceService_GetPerformance_Success(t *testing.T) {
	outlets := &mockOutletRepo{}
	perf := &mockPerfRepo{}
	tgts := &mockTargetRepo{}
	svc := service.NewPerformanceService(outlets, perf, tgts)

	tc := "DELHI-01"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	outlet := &model.RetailOutlet{CCNumber: "123456", Name: "Test Outlet", TerritoryCode: &tc, OutletType: "regular"}
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)
	userID := uuid.New()

	achieved := 500.0
	lastYear := 400.0
	recs := []*model.PerformanceRecord{
		{ID: uuid.New(), CCNumber: "123456", ProductID: 1, Period: period, Achieved: &achieved, LastYear: &lastYear},
	}
	targets := []*model.Target{
		{ID: uuid.New(), CCNumber: "123456", ProductID: 1, Period: period, TargetValue: 480},
	}

	outlets.On("GetByCC", ctx, "123456").Return(outlet, nil)
	perf.On("GetByPeriod", ctx, "123456", period).Return(recs, nil)
	tgts.On("GetByPeriod", ctx, "123456", period).Return(targets, nil)

	resp, err := svc.GetPerformance(ctx, "123456", period, userID)
	require.NoError(t, err)
	assert.Equal(t, "123456", resp.CCNumber)
	assert.Equal(t, "Test Outlet", resp.OutletName)
	assert.Equal(t, "2026-03", resp.Period)
	require.NotEmpty(t, resp.Fuel)
	assert.Equal(t, "MS", resp.Fuel[0].ProductCode)
	assert.InDelta(t, 500.0, *resp.Fuel[0].Achieved, 0.001)
	assert.InDelta(t, 480.0, *resp.Fuel[0].Target, 0.001)
	// target_achievement_pct = 500/480*100 ≈ 104.17
	require.NotNil(t, resp.TargetAchievementPct)
	assert.InDelta(t, 104.166, *resp.TargetAchievementPct, 0.01)
	// yoy_growth_pct = (500-400)/400*100 = 25
	require.NotNil(t, resp.YoYGrowthPct)
	assert.InDelta(t, 25.0, *resp.YoYGrowthPct, 0.001)
}

func TestPerformanceService_GetPerformance_Forbidden(t *testing.T) {
	outlets := &mockOutletRepo{}
	perf := &mockPerfRepo{}
	tgts := &mockTargetRepo{}
	svc := service.NewPerformanceService(outlets, perf, tgts)

	tc := "DELHI-02"
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	outlet := &model.RetailOutlet{CCNumber: "123456", TerritoryCode: strPtr("DELHI-01"), OutletType: "regular"}
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)

	outlets.On("GetByCC", ctx, "123456").Return(outlet, nil)

	_, err := svc.GetPerformance(ctx, "123456", period, uuid.New())
	assert.ErrorIs(t, err, model.ErrForbidden)
}
