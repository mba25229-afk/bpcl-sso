package repository_test

import (
	"context"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/repository"
	"github.com/stretchr/testify/require"
)

const seedOutletCC = "112847"
const seedOutletName = "M.L. SETHI SERVICE STATION"

func TestOutletRepo_GetByCC(t *testing.T) {
	repo := repository.NewOutletRepo(testPool)
	ctx := context.Background()

	o, err := repo.GetByCC(ctx, seedOutletCC)
	require.NoError(t, err)
	require.Equal(t, seedOutletCC, o.CCNumber)
	require.Equal(t, seedOutletName, o.Name)
	require.NotNil(t, o.TerritoryCode)
	require.Equal(t, "DELHI-W", *o.TerritoryCode)
}

func TestOutletRepo_GetByCC_NotFound(t *testing.T) {
	repo := repository.NewOutletRepo(testPool)
	_, err := repo.GetByCC(context.Background(), "999999")
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestOutletRepo_ListByTerritory(t *testing.T) {
	repo := repository.NewOutletRepo(testPool)
	ctx := context.Background()

	outlets, err := repo.ListByTerritory(ctx, "DELHI-W")
	require.NoError(t, err)
	require.NotEmpty(t, outlets)

	for _, o := range outlets {
		require.NotNil(t, o.TerritoryCode)
		require.Equal(t, "DELHI-W", *o.TerritoryCode)
	}
}

func TestOutletRepo_ListAll(t *testing.T) {
	repo := repository.NewOutletRepo(testPool)
	ctx := context.Background()

	outlets, err := repo.ListAll(ctx)
	require.NoError(t, err)
	// Seed has 40 outlets (39 dealers + MAHADEV + adhoc)
	require.Equal(t, 40, len(outlets))
}
