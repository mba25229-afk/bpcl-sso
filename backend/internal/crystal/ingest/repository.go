package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertDaily writes daily rows into the appropriate cr_ table.
// Duplicate (cc_code, txn_date) updates the value.
func (r *Repository) UpsertDaily(ctx context.Context, metric MetricKey, date time.Time, rows []DailyRow) (inserted, updated int, err error) {
	table, col := dailyTable(metric)

	sql := fmt.Sprintf(`
		INSERT INTO %s (cc_code, txn_date, %s)
		VALUES ($1, $2, $3)
		ON CONFLICT (cc_code, txn_date) DO UPDATE SET %s = EXCLUDED.%s
	`, table, col, col, col)

	batch := &pgx.Batch{}
	for _, row := range rows {
		batch.Queue(sql, row.CCCode, date, row.Value)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range rows {
		tag, e := br.Exec()
		if e != nil {
			return inserted, updated, fmt.Errorf("upsert %s: %w", table, e)
		}
		if tag.String() == "INSERT 0 1" {
			inserted++
		} else {
			updated++
		}
	}
	return inserted, updated, nil
}

// UpsertMAKGE inserts or updates a single MAK GE meter reading.
func (r *Repository) UpsertMAKGE(ctx context.Context, req MAKGERequest, date time.Time) (inserted bool, err error) {
	sql := `
		INSERT INTO cr_mak_ge_readings (cc_code, reading_date, meter_reading)
		VALUES ($1, $2, $3)
		ON CONFLICT (cc_code, reading_date) DO UPDATE SET meter_reading = EXCLUDED.meter_reading
		RETURNING (xmax = 0) AS was_inserted
	`
	err = r.pool.QueryRow(ctx, sql, req.CCCode, date, req.MeterReading).Scan(&inserted)
	return inserted, err
}

// UpsertGoogleRating inserts or updates a Google rating snapshot.
func (r *Repository) UpsertGoogleRating(ctx context.Context, req GoogleRatingRequest, date time.Time) (inserted bool, err error) {
	sql := `
		INSERT INTO cr_google_ratings (cc_code, snapshot_date, rating, review_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (cc_code, snapshot_date) DO UPDATE
			SET rating = EXCLUDED.rating, review_count = EXCLUDED.review_count
		RETURNING (xmax = 0) AS was_inserted
	`
	err = r.pool.QueryRow(ctx, sql, req.CCCode, date, req.Rating, req.ReviewCount).Scan(&inserted)
	return inserted, err
}

// UpsertTargets bulk-upserts monthly targets for all dealers.
func (r *Repository) UpsertTargets(ctx context.Context, monthYear time.Time, rows []TargetRow) (inserted, updated int, err error) {
	sql := `
		INSERT INTO cr_monthly_targets (
			cc_code, month_year,
			ufill_target, qoc_target, speed_kl, ms_kl, hsd_kl,
			ms_ly, hsd_ly, dsw_available, nitrogen,
			mak_ge_target, darpan_target, coolant_lube_kl, remarks
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (cc_code, month_year) DO UPDATE SET
			ufill_target    = EXCLUDED.ufill_target,
			qoc_target      = EXCLUDED.qoc_target,
			speed_kl        = EXCLUDED.speed_kl,
			ms_kl           = EXCLUDED.ms_kl,
			hsd_kl          = EXCLUDED.hsd_kl,
			ms_ly           = EXCLUDED.ms_ly,
			hsd_ly          = EXCLUDED.hsd_ly,
			dsw_available   = EXCLUDED.dsw_available,
			nitrogen        = EXCLUDED.nitrogen,
			mak_ge_target   = EXCLUDED.mak_ge_target,
			darpan_target   = EXCLUDED.darpan_target,
			coolant_lube_kl = EXCLUDED.coolant_lube_kl,
			remarks         = EXCLUDED.remarks
	`

	batch := &pgx.Batch{}
	for _, t := range rows {
		batch.Queue(sql,
			t.CCCode, monthYear,
			t.UfillTarget, t.QOCTarget, t.SpeedKL, t.MSKL, t.HSDKL,
			t.MSLY, t.HSDLY, t.DSWAvailable, t.Nitrogen,
			t.MAKGETarget, t.DarpanTarget, t.CoolantLubeKL, t.Remarks,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range rows {
		tag, e := br.Exec()
		if e != nil {
			return inserted, updated, fmt.Errorf("upsert target: %w", e)
		}
		if tag.String() == "INSERT 0 1" {
			inserted++
		} else {
			updated++
		}
	}
	return inserted, updated, nil
}

func dailyTable(m MetricKey) (table, col string) {
	switch m {
	case MetricUfill:
		return "cr_daily_ufill", "count"
	case MetricQOC:
		return "cr_daily_qoc", "count"
	case MetricMS:
		return "cr_daily_ms", "kl"
	case MetricHSD:
		return "cr_daily_hsd", "kl"
	case MetricSpeed:
		return "cr_daily_speed", "kl"
	default:
		panic("unknown metric: " + string(m))
	}
}
