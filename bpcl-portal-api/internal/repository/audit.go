package repository

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

// Log writes an audit entry. Non-blocking: DB failures are logged but not returned.
func (r *AuditRepo) Log(ctx context.Context, userID uuid.UUID, action, cc string, payload any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Printf("audit: marshal payload: %v", err)
		payloadJSON = []byte("{}")
	}

	var ccPtr *string
	if cc != "" {
		ccPtr = &cc
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO audit_log (user_id, action, cc_number, payload) VALUES ($1, $2, $3, $4)`,
		userID, action, ccPtr, payloadJSON,
	)
	if err != nil {
		log.Printf("audit: write failed (action=%s cc=%v): %v", action, ccPtr, err)
	}
	return nil
}
