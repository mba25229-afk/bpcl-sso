package repository_test

import (
	"context"
	"testing"

	"github.com/bpcl/portal-api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuditRepo_Log(t *testing.T) {
	repo := repository.NewAuditRepo(testPool)
	ctx := context.Background()

	adminID, _ := uuid.Parse(seedAdminUUID)

	// Log always returns nil (non-blocking)
	err := repo.Log(ctx, adminID, "test_action", seedOutletCC, map[string]string{"key": "value"})
	require.NoError(t, err)

	// Verify the entry was written
	var count int
	testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_log WHERE user_id = $1 AND action = 'test_action'`,
		adminID,
	).Scan(&count)
	require.GreaterOrEqual(t, count, 1)

	// Cleanup
	t.Cleanup(func() {
		testPool.Exec(context.Background(),
			`DELETE FROM audit_log WHERE user_id = $1 AND action = 'test_action'`,
			adminID)
	})
}

func TestAuditRepo_Log_EmptyCC(t *testing.T) {
	repo := repository.NewAuditRepo(testPool)
	ctx := context.Background()

	adminID, _ := uuid.Parse(seedAdminUUID)

	// Log with empty cc should also work (stores NULL)
	err := repo.Log(ctx, adminID, "test_no_cc", "", nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		testPool.Exec(context.Background(),
			`DELETE FROM audit_log WHERE user_id = $1 AND action = 'test_no_cc'`,
			adminID)
	})
}
