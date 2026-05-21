package repository_test

import (
	"context"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Seeded in chunk 1 seed_dev.sql
const (
	seedAdminUUID     = "a1b2c3d4-0001-0001-0001-000000000001"
	seedAdminEmpID    = "EMP10001"
	seedAdminEmail    = "admin@bpcl.in"
)

func TestUserRepo_GetByEmployeeID(t *testing.T) {
	repo := repository.NewUserRepo(testPool)
	ctx := context.Background()

	u, err := repo.GetByEmployeeID(ctx, seedAdminEmpID)
	require.NoError(t, err)
	require.Equal(t, seedAdminEmpID, u.EmployeeID)
	require.Equal(t, seedAdminEmail, u.Email)
	require.Equal(t, model.RoleAdmin, u.Role)
	require.Nil(t, u.TerritoryCode) // admin has NULL territory
}

func TestUserRepo_GetByEmployeeID_NotFound(t *testing.T) {
	repo := repository.NewUserRepo(testPool)
	_, err := repository.NewUserRepo(testPool).GetByEmployeeID(context.Background(), "NOBODY")
	require.ErrorIs(t, err, model.ErrNotFound)
	_ = repo
}

func TestUserRepo_GetByID(t *testing.T) {
	repo := repository.NewUserRepo(testPool)
	ctx := context.Background()

	id, err := uuid.Parse(seedAdminUUID)
	require.NoError(t, err)

	u, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, seedAdminEmpID, u.EmployeeID)
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	repo := repository.NewUserRepo(testPool)
	_, err := repo.GetByID(context.Background(), uuid.New())
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestUserRepo_UpdateLastLogin(t *testing.T) {
	repo := repository.NewUserRepo(testPool)
	ctx := context.Background()

	id, err := uuid.Parse(seedAdminUUID)
	require.NoError(t, err)

	// Should not error
	require.NoError(t, repo.UpdateLastLogin(ctx, id))

	// Verify last_login_at is now set
	u, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, u.LastLoginAt)
}
