package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/bpcl/portal-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadService_GetHistory(t *testing.T) {
	uploads := &mockUploadRepo{}
	perf := &mockPerfRepo{}
	svc := service.NewUploadService(uploads, perf, t.TempDir(), 10)

	userID := uuid.New()
	ctx := context.Background()
	period := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	items := []*model.UploadedFile{
		{ID: uuid.New(), UserID: userID, Status: "done", Period: &period},
	}
	uploads.On("ListByUser", ctx, userID, "", 20, 0).Return(items, 1, nil)

	resp, err := svc.GetHistory(ctx, userID, "", 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Items, 1)
}

func TestUploadService_GetHistory_WithCC(t *testing.T) {
	uploads := &mockUploadRepo{}
	perf := &mockPerfRepo{}
	svc := service.NewUploadService(uploads, perf, t.TempDir(), 10)

	userID := uuid.New()
	ctx := context.Background()
	uploads.On("ListByUser", ctx, userID, "123456", 10, 5).Return([]*model.UploadedFile{}, 0, nil)

	resp, err := svc.GetHistory(ctx, userID, "123456", 10, 5)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
}

func TestUploadService_ProcessAsync_BadFile(t *testing.T) {
	uploads := &mockUploadRepo{}
	perf := &mockPerfRepo{}
	svc := service.NewUploadService(uploads, perf, t.TempDir(), 10)

	uploadID := uuid.New()
	uploads.On("UpdateStatus", mock.Anything, uploadID, "failed", mock.Anything, 0).Return(nil)

	// Call synchronously to test — file doesn't exist, so it will fail
	svc.ProcessAsync(uploadID, "/nonexistent/path.xlsx", "123456", time.Now(), uuid.New())

	uploads.AssertCalled(t, "UpdateStatus", mock.Anything, uploadID, "failed", mock.AnythingOfType("string"), 0)
}
