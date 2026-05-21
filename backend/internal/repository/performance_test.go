package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/repository"
	"github.com/stretchr/testify/require"
)

var (
	seedPerfPeriod = time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	seedPerfFrom   = time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	seedPerfTo     = time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
)

func TestPerformanceRepo_GetByPeriod(t *testing.T) {
	repo := repository.NewPerformanceRepo(testPool)
	ctx := context.Background()

	recs, err := repo.GetByPeriod(ctx, seedOutletCC, seedPerfPeriod)
	require.NoError(t, err)
	// 8 products for this outlet
	require.Len(t, recs, 8)
	for _, r := range recs {
		require.Equal(t, seedOutletCC, r.CCNumber)
	}
}

func TestPerformanceRepo_GetDateRange(t *testing.T) {
	repo := repository.NewPerformanceRepo(testPool)
	ctx := context.Background()

	recs, err := repo.GetDateRange(ctx, seedOutletCC, seedPerfFrom, seedPerfTo)
	require.NoError(t, err)
	// April + May = 2 months × 8 products = 16
	require.Len(t, recs, 16)
}

func TestPerformanceRepo_Upsert(t *testing.T) {
	repo := repository.NewPerformanceRepo(testPool)
	ctx := context.Background()

	testPeriod := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	achieved := 99.99
	rec := &model.PerformanceRecord{
		CCNumber:  seedOutletCC,
		ProductID: 1,
		Period:    testPeriod,
		Achieved:  &achieved,
		Source:    "manual",
	}

	t.Cleanup(func() {
		testPool.Exec(context.Background(),
			`DELETE FROM performance_records WHERE cc_number = $1 AND period = $2`,
			seedOutletCC, testPeriod)
	})

	require.NoError(t, repo.Upsert(ctx, rec))

	// Verify insert
	recs, err := repo.GetByPeriod(ctx, seedOutletCC, testPeriod)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.InDelta(t, achieved, *recs[0].Achieved, 0.01)

	// Upsert with new value
	updated := 123.45
	rec.Achieved = &updated
	require.NoError(t, repo.Upsert(ctx, rec))

	recs2, err := repo.GetByPeriod(ctx, seedOutletCC, testPeriod)
	require.NoError(t, err)
	require.InDelta(t, updated, *recs2[0].Achieved, 0.01)
}

func TestPerformanceRepo_UpsertBatch(t *testing.T) {
	repo := repository.NewPerformanceRepo(testPool)
	ctx := context.Background()

	testPeriod := time.Date(2098, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Cleanup(func() {
		testPool.Exec(context.Background(),
			`DELETE FROM performance_records WHERE period = $1`, testPeriod)
	})

	val1, val2 := 10.0, 20.0
	recs := []*model.PerformanceRecord{
		{CCNumber: seedOutletCC, ProductID: 1, Period: testPeriod, Achieved: &val1, Source: "manual"},
		{CCNumber: seedOutletCC, ProductID: 2, Period: testPeriod, Achieved: &val2, Source: "manual"},
	}

	require.NoError(t, repo.UpsertBatch(ctx, testPool, recs))

	got, err := repo.GetByPeriod(ctx, seedOutletCC, testPeriod)
	require.NoError(t, err)
	require.Len(t, got, 2)
}
