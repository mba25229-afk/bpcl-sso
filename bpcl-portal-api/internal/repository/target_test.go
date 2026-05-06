package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var seedTargetPeriod = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

func TestTargetRepo_GetByPeriod(t *testing.T) {
	repo := repository.NewTargetRepo(testPool)
	ctx := context.Background()

	targets, err := repo.GetByPeriod(ctx, seedOutletCC, seedTargetPeriod)
	require.NoError(t, err)
	require.NotEmpty(t, targets)
	for _, tgt := range targets {
		require.Equal(t, seedOutletCC, tgt.CCNumber)
		require.Greater(t, tgt.TargetValue, 0.0)
	}
}

func TestTargetRepo_Upsert(t *testing.T) {
	repo := repository.NewTargetRepo(testPool)
	ctx := context.Background()

	testPeriod := time.Date(2099, 2, 1, 0, 0, 0, 0, time.UTC)
	adminID, _ := uuid.Parse(seedAdminUUID)

	t.Cleanup(func() {
		testPool.Exec(context.Background(),
			`DELETE FROM targets WHERE cc_number = $1 AND period = $2`,
			seedOutletCC, testPeriod)
	})

	tgt := &model.Target{
		CCNumber:    seedOutletCC,
		ProductID:   1,
		Period:      testPeriod,
		TargetValue: 500.0,
		SetBy:       &adminID,
	}
	require.NoError(t, repo.Upsert(ctx, tgt))

	got, err := repo.GetByPeriod(ctx, seedOutletCC, testPeriod)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.InDelta(t, 500.0, got[0].TargetValue, 0.01)

	// Upsert update
	tgt.TargetValue = 600.0
	require.NoError(t, repo.Upsert(ctx, tgt))

	got2, err := repo.GetByPeriod(ctx, seedOutletCC, testPeriod)
	require.NoError(t, err)
	require.InDelta(t, 600.0, got2[0].TargetValue, 0.01)
}

func TestTargetRepo_UpsertBatch(t *testing.T) {
	repo := repository.NewTargetRepo(testPool)
	ctx := context.Background()

	testPeriod := time.Date(2098, 2, 1, 0, 0, 0, 0, time.UTC)

	t.Cleanup(func() {
		testPool.Exec(context.Background(),
			`DELETE FROM targets WHERE period = $1`, testPeriod)
	})

	targets := []*model.Target{
		{CCNumber: seedOutletCC, ProductID: 1, Period: testPeriod, TargetValue: 100},
		{CCNumber: seedOutletCC, ProductID: 2, Period: testPeriod, TargetValue: 200},
	}

	require.NoError(t, repo.UpsertBatch(ctx, targets))

	got, err := repo.GetByPeriod(ctx, seedOutletCC, testPeriod)
	require.NoError(t, err)
	require.Len(t, got, 2)
}
