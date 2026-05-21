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

func TestCompetitionService_GetLeaderboard_FilterAdhoc(t *testing.T) {
	comp := &mockCompRepo{}
	outlets := &mockOutletRepo{}
	svc := service.NewCompetitionService(comp, outlets)

	tc := "DELHI-01"
	compID := uuid.New()
	period := &model.CompetitionPeriod{
		ID:            compID,
		Name:          "Boost and Win March 2026",
		Period:        time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		TerritoryCode: tc,
		Status:        "active",
	}

	score1 := 56.37
	score2 := 37.02
	scores := []*model.CompetitionScore{
		{CCNumber: "112847", TotalScore: &score1},
		{CCNumber: "259713", TotalScore: &score2}, // adhoc dealer
	}

	outletList := []*model.RetailOutlet{
		{CCNumber: "112847", Name: "M.L. SETHI SERVICE STATION", TerritoryCode: &tc, OutletType: "regular"},
		{CCNumber: "259713", Name: "SAKSHAM MOTORS ADHOC", TerritoryCode: &tc, OutletType: "adhoc"},
	}

	ctx := ctxWithClaims(model.RoleAdmin, nil)

	comp.On("GetActivePeriod", ctx, tc).Return(period, nil)
	comp.On("GetScores", ctx, compID).Return(scores, nil)
	outlets.On("ListAll", ctx).Return(outletList, nil)

	resp, err := svc.GetLeaderboard(ctx, tc, uuid.New())
	require.NoError(t, err)
	assert.Len(t, resp.Entries, 1, "adhoc outlet should be filtered out")
	assert.Equal(t, "112847", resp.Entries[0].CCNumber)
	assert.Equal(t, 1, resp.Entries[0].Rank)
	assert.True(t, resp.Entries[0].IsHighlighted)
}

func TestCompetitionService_GetLeaderboard_NoActivePeriod(t *testing.T) {
	comp := &mockCompRepo{}
	outlets := &mockOutletRepo{}
	svc := service.NewCompetitionService(comp, outlets)

	tc := "DELHI-01"
	ctx := ctxWithClaims(model.RoleAdmin, nil)
	comp.On("GetActivePeriod", ctx, tc).Return(nil, model.ErrNotFound)

	_, err := svc.GetLeaderboard(ctx, tc, uuid.New())
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestCompetitionService_GetDealerScorecard_Success(t *testing.T) {
	comp := &mockCompRepo{}
	outlets := &mockOutletRepo{}
	svc := service.NewCompetitionService(comp, outlets)

	tc := "DELHI-01"
	compID := uuid.New()
	total := 56.37
	rank := 1
	score := &model.CompetitionScore{
		CCNumber:   "112847",
		TotalScore: &total,
		Rank:       &rank,
	}
	outlet := &model.RetailOutlet{CCNumber: "112847", Name: "M.L. SETHI", TerritoryCode: &tc, OutletType: "regular"}
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)

	outlets.On("GetByCC", ctx, "112847").Return(outlet, nil)
	comp.On("GetDealerScore", ctx, compID, "112847").Return(score, nil)

	resp, err := svc.GetDealerScorecard(ctx, "112847", compID, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "112847", resp.CCNumber)
	assert.InDelta(t, 56.37, *resp.TotalScore, 0.001)
	assert.Equal(t, 1, *resp.Rank)
	assert.Len(t, resp.Parameters, 15)
}

func TestCompetitionService_GetDealerScorecard_Forbidden(t *testing.T) {
	comp := &mockCompRepo{}
	outlets := &mockOutletRepo{}
	svc := service.NewCompetitionService(comp, outlets)

	tc := "DELHI-02"
	compID := uuid.New()
	outlet := &model.RetailOutlet{CCNumber: "112847", TerritoryCode: strPtr("DELHI-01"), OutletType: "regular"}
	ctx := ctxWithClaims(model.RoleTerritoryManager, &tc)

	outlets.On("GetByCC", ctx, "112847").Return(outlet, nil)

	_, err := svc.GetDealerScorecard(ctx, "112847", compID, uuid.New())
	assert.ErrorIs(t, err, model.ErrForbidden)
}
