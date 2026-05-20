package etl

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Loader struct {
	pool *pgxpool.Pool
}

func NewLoader(pool *pgxpool.Pool) *Loader {
	return &Loader{pool: pool}
}

type LoadResult struct {
	Inserted int
	Skipped  int
	Errors   []string
}

func (l *Loader) LoadTargets(ctx context.Context, records []TargetRecord) (*LoadResult, error) {
	if len(records) == 0 {
		return &LoadResult{Inserted: 0, Skipped: 0, Errors: nil}, nil
	}

	result := &LoadResult{
		Inserted: 0,
		Skipped:  0,
		Errors:   []string{},
	}

	for i, record := range records {
		exists, err := l.targetExists(ctx, record.CCNumber, record.ProductID, record.Period)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: check failed: %v", i+1, err))
			continue
		}

		if exists {
			result.Skipped++
			continue
		}

		err = l.insertTarget(ctx, &record)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: insert failed: %v", i+1, err))
			continue
		}

		result.Inserted++
	}

	return result, nil
}

func (l *Loader) targetExists(ctx context.Context, ccNumber string, productID int16, period time.Time) (bool, error) {
	var count int
	err := l.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM targets 
		WHERE cc_number = $1 AND product_id = $2 AND period = $3
	`, ccNumber, productID, period).Scan(&count)
	return count > 0, err
}

func (l *Loader) insertTarget(ctx context.Context, record *TargetRecord) error {
	_, err := l.pool.Exec(ctx, `
		INSERT INTO targets (cc_number, product_id, period, target_value, set_by)
		VALUES ($1, $2, $3, $4, NULL)
	`, record.CCNumber, record.ProductID, record.Period, record.TargetValue)
	return err
}

func (l *Loader) LogETLRun(ctx context.Context, status string, detail string, durationMs int) error {
	_, err := l.pool.Exec(ctx, `
		INSERT INTO cr_etl_log (status, detail, duration_ms, source)
		VALUES ($1, $2, $3, 'microsoft_excel')
	`, status, detail, durationMs)
	return err
}