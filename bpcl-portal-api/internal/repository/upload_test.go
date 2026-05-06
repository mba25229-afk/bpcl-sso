package repository_test

import (
	"context"
	"testing"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUploadRepo_CreateGetUpdateDelete(t *testing.T) {
	repo := repository.NewUploadRepo(testPool)
	ctx := context.Background()

	adminID, _ := uuid.Parse(seedAdminUUID)
	cc := seedOutletCC

	f := &model.UploadedFile{
		UserID:       adminID,
		OriginalName: "test_upload.xlsx",
		StoredPath:   "/tmp/test_upload.xlsx",
		UploadType:   "performance",
		Status:       "pending",
		CCNumber:     &cc,
	}

	// Create
	require.NoError(t, repo.Create(ctx, f))
	require.NotEqual(t, uuid.Nil, f.ID)

	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM uploaded_files WHERE id = $1`, f.ID)
	})

	// GetByID
	got, err := repo.GetByID(ctx, f.ID)
	require.NoError(t, err)
	require.Equal(t, f.ID, got.ID)
	require.Equal(t, "test_upload.xlsx", got.OriginalName)
	require.Equal(t, "pending", got.Status)

	// ListByUser
	files, total, err := repo.ListByUser(ctx, adminID, "", 10, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, 1)
	_ = files

	// ListByUser filtered by cc
	files2, total2, err := repo.ListByUser(ctx, adminID, cc, 10, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, total2, 1)
	_ = files2

	// UpdateStatus
	require.NoError(t, repo.UpdateStatus(ctx, f.ID, "done", "", 100))
	got2, err := repo.GetByID(ctx, f.ID)
	require.NoError(t, err)
	require.Equal(t, "done", got2.Status)
	require.NotNil(t, got2.RowCount)
	require.Equal(t, 100, *got2.RowCount)

	// SoftDelete
	require.NoError(t, repo.SoftDelete(ctx, f.ID))
	got3, err := repo.GetByID(ctx, f.ID)
	require.NoError(t, err)
	require.Equal(t, "deleted", got3.Status)
}

func TestUploadRepo_GetByID_NotFound(t *testing.T) {
	repo := repository.NewUploadRepo(testPool)
	_, err := repo.GetByID(context.Background(), uuid.New())
	require.ErrorIs(t, err, model.ErrNotFound)
}
