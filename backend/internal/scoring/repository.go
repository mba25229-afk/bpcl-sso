package scoring

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScoreRepository struct {
	pool *pgxpool.Pool
}

func NewScoreRepository(pool *pgxpool.Pool) *ScoreRepository {
	return &ScoreRepository{pool: pool}
}

func (r *ScoreRepository) BulkUpsert(ctx context.Context, rows []ScoreRow) error {
	if len(rows) == 0 {
		return nil
	}

	sql := `
		INSERT INTO cr_dealer_scores
			(cc_code, month_year, metric_key, actual_value, rank_in_group, marks_scored, max_marks)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (cc_code, month_year, metric_key) DO UPDATE SET
			actual_value  = EXCLUDED.actual_value,
			rank_in_group = EXCLUDED.rank_in_group,
			marks_scored  = EXCLUDED.marks_scored,
			max_marks     = EXCLUDED.max_marks,
			computed_at   = NOW()
	`

	batch := &pgx.Batch{}
	for _, r := range rows {
		batch.Queue(sql, r.CCCode, r.MonthYear, r.MetricKey,
			r.ActualValue, r.RankInGroup, r.MarksScored, r.MaxMarks)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range rows {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

type ParamsRepo struct {
	pool *pgxpool.Pool
}

func NewParamsRepo(pool *pgxpool.Pool) *ParamsRepo {
	return &ParamsRepo{pool: pool}
}

func (r *ParamsRepo) FetchActive(ctx context.Context, monthYear time.Time) ([]ScoringParam, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT metric_key, max_marks, negative_scale_enabled, is_active
		 FROM cr_scoring_params
		 WHERE month_year = $1
		 ORDER BY sort_order`,
		monthYear,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ScoringParam
	for rows.Next() {
		var p ScoringParam
		if err := rows.Scan(&p.MetricKey, &p.MaxMarks, &p.NegativeScaleEnabled, &p.IsActive); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}