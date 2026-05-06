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

type PerformanceRepo struct {
	pool *pgxpool.Pool
}

func NewPerformanceRepo(pool *pgxpool.Pool) *PerformanceRepo {
	return &PerformanceRepo{pool: pool}
}

const perfSelect = `SELECT id, cc_number, product_id, period, achieved, last_year,
	volume_kl, source, uploaded_file_id, created_at, updated_at
	FROM performance_records`

func (r *PerformanceRepo) GetByPeriod(ctx context.Context, cc string, period time.Time) ([]*model.PerformanceRecord, error) {
	rows, err := r.pool.Query(ctx,
		perfSelect+` WHERE cc_number = $1 AND period = $2 ORDER BY product_id`,
		cc, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPerf(rows)
}

func (r *PerformanceRepo) GetDateRange(ctx context.Context, cc string, from, to time.Time) ([]*model.PerformanceRecord, error) {
	rows, err := r.pool.Query(ctx,
		perfSelect+` WHERE cc_number = $1 AND period >= $2 AND period <= $3 ORDER BY period, product_id`,
		cc, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPerf(rows)
}

func (r *PerformanceRepo) Upsert(ctx context.Context, rec *model.PerformanceRecord) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO performance_records
			(cc_number, product_id, period, achieved, last_year, volume_kl, source, uploaded_file_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (cc_number, product_id, period) DO UPDATE SET
			achieved         = EXCLUDED.achieved,
			last_year        = EXCLUDED.last_year,
			volume_kl        = EXCLUDED.volume_kl,
			source           = EXCLUDED.source,
			uploaded_file_id = EXCLUDED.uploaded_file_id,
			updated_at       = NOW()`,
		rec.CCNumber, rec.ProductID, rec.Period,
		rec.Achieved, rec.LastYear, rec.VolumeKL,
		rec.Source, rec.UploadedFileID,
	)
	return err
}

func (r *PerformanceRepo) UpsertBatch(ctx context.Context, pool *pgxpool.Pool, recs []*model.PerformanceRecord) error {
	if len(recs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	const q = `
		INSERT INTO performance_records
			(cc_number, product_id, period, achieved, last_year, volume_kl, source, uploaded_file_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (cc_number, product_id, period) DO UPDATE SET
			achieved         = EXCLUDED.achieved,
			last_year        = EXCLUDED.last_year,
			volume_kl        = EXCLUDED.volume_kl,
			source           = EXCLUDED.source,
			uploaded_file_id = EXCLUDED.uploaded_file_id,
			updated_at       = NOW()`
	for _, rec := range recs {
		batch.Queue(q,
			rec.CCNumber, rec.ProductID, rec.Period,
			rec.Achieved, rec.LastYear, rec.VolumeKL,
			rec.Source, rec.UploadedFileID,
		)
	}
	br := pool.SendBatch(ctx, batch)
	defer br.Close()
	for range recs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func scanPerf(row pgx.Row) (*model.PerformanceRecord, error) {
	var rec model.PerformanceRecord
	var rawID [16]byte
	var achieved, lastYear, volumeKL pgtype.Numeric
	var uploadedFileID pgtype.UUID

	err := row.Scan(
		&rawID, &rec.CCNumber, &rec.ProductID, &rec.Period,
		&achieved, &lastYear, &volumeKL,
		&rec.Source, &uploadedFileID,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	rec.ID = uuid.UUID(rawID)
	rec.Achieved = numericToFloat64Ptr(achieved)
	rec.LastYear = numericToFloat64Ptr(lastYear)
	rec.VolumeKL = numericToFloat64Ptr(volumeKL)
	rec.UploadedFileID = pguuid(uploadedFileID)
	return &rec, nil
}

func collectPerf(rows pgx.Rows) ([]*model.PerformanceRecord, error) {
	var result []*model.PerformanceRecord
	for rows.Next() {
		rec, err := scanPerf(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, rec)
	}
	return result, rows.Err()
}

func numericToFloat64Ptr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	return &f.Float64
}

func (r *PerformanceRepo) GetTrend(ctx context.Context, cc string, months int) ([]*model.TrendRow, error) {
	now := time.Now()
	to := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	from := to.AddDate(0, -months+1, 0)

	rows, err := r.pool.Query(ctx, `
		SELECT 
			p.period,
			SUM(CASE WHEN p.product_id IN (1,2,3) THEN p.achieved ELSE 0 END) as fuel_achieved,
			SUM(CASE WHEN p.product_id NOT IN (1,2,3) THEN p.achieved ELSE 0 END) as non_fuel_achieved,
			COALESCE(SUM(p.achieved), 0) as total_achieved
		FROM performance_records p
		WHERE p.cc_number = $1 AND p.period >= $2 AND p.period <= $3
		GROUP BY p.period
		ORDER BY p.period DESC`,
		cc, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.TrendRow
	for rows.Next() {
		var period time.Time
		var fuelAchieved, nonFuelAchieved, totalAchieved pgtype.Numeric
		if err := rows.Scan(&period, &fuelAchieved, &nonFuelAchieved, &totalAchieved); err != nil {
			return nil, err
		}
		rec := &model.TrendRow{
			Period:            period.Format("2006-01"),
			FuelAchieved:     numericToFloat64Ptr(fuelAchieved),
			NonFuelAchieved:  numericToFloat64Ptr(nonFuelAchieved),
			TotalAchieved:   numericToFloat64Ptr(totalAchieved),
		}
		result = append(result, rec)
	}
	return result, rows.Err()
}
