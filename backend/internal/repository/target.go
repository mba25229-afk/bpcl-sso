package repository

import (
	"context"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TargetRepo struct {
	pool *pgxpool.Pool
}

func NewTargetRepo(pool *pgxpool.Pool) *TargetRepo {
	return &TargetRepo{pool: pool}
}

const targetSelect = `SELECT id, cc_number, product_id, period, target_value, set_by, created_at, updated_at FROM targets`

func (r *TargetRepo) GetByPeriod(ctx context.Context, cc string, period time.Time) ([]*model.Target, error) {
	rows, err := r.pool.Query(ctx,
		targetSelect+` WHERE cc_number = $1 AND period = $2 ORDER BY product_id`,
		cc, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectTargets(rows)
}

func (r *TargetRepo) Upsert(ctx context.Context, t *model.Target) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO targets (cc_number, product_id, period, target_value, set_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cc_number, product_id, period) DO UPDATE SET
			target_value = EXCLUDED.target_value,
			set_by       = EXCLUDED.set_by,
			updated_at   = NOW()`,
		t.CCNumber, t.ProductID, t.Period, t.TargetValue, t.SetBy,
	)
	return err
}

func (r *TargetRepo) UpsertBatch(ctx context.Context, targets []*model.Target) error {
	if len(targets) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	const q = `
		INSERT INTO targets (cc_number, product_id, period, target_value, set_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cc_number, product_id, period) DO UPDATE SET
			target_value = EXCLUDED.target_value,
			set_by       = EXCLUDED.set_by,
			updated_at   = NOW()`
	for _, t := range targets {
		batch.Queue(q, t.CCNumber, t.ProductID, t.Period, t.TargetValue, t.SetBy)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range targets {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func scanTarget(row pgx.Row) (*model.Target, error) {
	var t model.Target
	var rawID [16]byte
	var targetVal pgtype.Numeric
	var setBy pgtype.UUID

	err := row.Scan(&rawID, &t.CCNumber, &t.ProductID, &t.Period, &targetVal, &setBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	t.ID = uuid.UUID(rawID)
	if v := numericToFloat64Ptr(targetVal); v != nil {
		t.TargetValue = *v
	}
	t.SetBy = pguuid(setBy)
	return &t, nil
}

func collectTargets(rows pgx.Rows) ([]*model.Target, error) {
	var result []*model.Target
	for rows.Next() {
		t, err := scanTarget(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}
