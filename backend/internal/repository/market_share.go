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

type MarketShareRepo struct {
	pool *pgxpool.Pool
}

func NewMarketShareRepo(pool *pgxpool.Pool) *MarketShareRepo {
	return &MarketShareRepo{pool: pool}
}

func (r *MarketShareRepo) UpsertMarketShareData(ctx context.Context, rows []model.MarketShareData) error {
	if len(rows) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, row := range rows {
		var uploadedFileID *uuid.UUID
		if row.UploadedFileID != nil {
			uploadedFileID = row.UploadedFileID
		}
		batch.Queue(`
			INSERT INTO market_share_data (outlet_name, cc_number, omc, trading_area_id, period, ms_vol_kl, hsd_vol_kl, source, uploaded_file_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (outlet_name, omc, trading_area_id, period) DO UPDATE SET
				cc_number = EXCLUDED.cc_number,
				ms_vol_kl = EXCLUDED.ms_vol_kl,
				hsd_vol_kl = EXCLUDED.hsd_vol_kl,
				source = EXCLUDED.source,
				uploaded_file_id = EXCLUDED.uploaded_file_id`,
			row.OutletName, row.CCNumber, row.OMC, row.TradingAreaID, row.Period,
			row.MSVolKL, row.HSDVolKL, row.Source, uploadedFileID,
		)
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

func (r *MarketShareRepo) UpsertTradingAreaTotals(ctx context.Context, totals []model.TradingAreaTotal) error {
	if len(totals) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, t := range totals {
		batch.Queue(`
			INSERT INTO trading_area_totals (trading_area_id, period, total_ms_kl, total_hsd_kl, bpcl_ms_kl, hpcl_ms_kl, iocl_ms_kl, bpcl_hsd_kl, hpcl_hsd_kl, iocl_hsd_kl)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (trading_area_id, period) DO UPDATE SET
				total_ms_kl = EXCLUDED.total_ms_kl,
				total_hsd_kl = EXCLUDED.total_hsd_kl,
				bpcl_ms_kl = EXCLUDED.bpcl_ms_kl,
				hpcl_ms_kl = EXCLUDED.hpcl_ms_kl,
				iocl_ms_kl = EXCLUDED.iocl_ms_kl,
				bpcl_hsd_kl = EXCLUDED.bpcl_hsd_kl,
				hpcl_hsd_kl = EXCLUDED.hpcl_hsd_kl,
				iocl_hsd_kl = EXCLUDED.iocl_hsd_kl,
				computed_at = NOW()`,
			t.TradingAreaID, t.Period, t.TotalMSKL, t.TotalHSDKL,
			t.BPCLMSKL, t.HPCLMSKL, t.IOCLMSKL,
			t.BPCLHSDKL, t.HPCLHSDKL, t.IOCLHSDKL,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range totals {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *MarketShareRepo) GetDealerVolumeForPeriod(ctx context.Context, cc string, period time.Time) (msVol, hsdVol float64, err error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(ms_vol_kl, 0), COALESCE(hsd_vol_kl, 0)
		FROM market_share_data
		WHERE cc_number = $1 AND period = $2`,
		cc, period,
	)
	err = row.Scan(&msVol, &hsdVol)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return msVol, hsdVol, nil
}

func (r *MarketShareRepo) GetTradingAreaTotal(ctx context.Context, tradingAreaID int, period time.Time) (*model.TradingAreaTotal, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, trading_area_id, period, total_ms_kl, total_hsd_kl, 
		       bpcl_ms_kl, hpcl_ms_kl, iocl_ms_kl, bpcl_hsd_kl, hpcl_hsd_kl, iocl_hsd_kl, computed_at
		FROM trading_area_totals
		WHERE trading_area_id = $1 AND period = $2`,
		tradingAreaID, period,
	)

	var t model.TradingAreaTotal
	var rawID [16]byte
	var computedAt pgtype.Timestamp
	var totalMS, totalHSD, bpclMS, hpclMS, ioclMS, bpclHSD, hpclHSD, ioclHSD pgtype.Numeric

	err := row.Scan(&rawID, &t.TradingAreaID, &t.Period, &totalMS, &totalHSD,
		&bpclMS, &hpclMS, &ioclMS, &bpclHSD, &hpclHSD, &ioclHSD, &computedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	t.ID = uuid.UUID(rawID)
	t.TotalMSKL = numericToFloat64(totalMS)
	t.TotalHSDKL = numericToFloat64(totalHSD)
	t.BPCLMSKL = numericToFloat64(bpclMS)
	t.HPCLMSKL = numericToFloat64(hpclMS)
	t.IOCLMSKL = numericToFloat64(ioclMS)
	t.BPCLHSDKL = numericToFloat64(bpclHSD)
	t.HPCLHSDKL = numericToFloat64(hpclHSD)
	t.IOCLHSDKL = numericToFloat64(ioclHSD)
	if computedAt.Valid {
		t.ComputedAt = computedAt.Time
	}
	return &t, nil
}

func (r *MarketShareRepo) GetTradingAreaByName(ctx context.Context, name string) (*model.TradingArea, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, district, state
		FROM trading_areas
		WHERE LOWER(name) = LOWER($1)`,
		name,
	)

	var t model.TradingArea
	err := row.Scan(&t.ID, &t.Name, &t.District, &t.State)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *MarketShareRepo) GetOutletsWithTradingArea(ctx context.Context, competitionID uuid.UUID) ([]*model.RetailOutlet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ro.cc_number, ro.name, ro.trading_area_id
		FROM retail_outlets ro
		JOIN competition_scores cs ON ro.cc_number = cs.cc_number
		WHERE cs.competition_id = $1 AND ro.outlet_type = 'regular'
		ORDER BY cs.total_score DESC NULLS LAST`,
		competitionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.RetailOutlet
	for rows.Next() {
		var o model.RetailOutlet
		var taID pgtype.Int4
		if err := rows.Scan(&o.CCNumber, &o.Name, &taID); err != nil {
			return nil, err
		}
		if taID.Valid {
			v := int(taID.Int32)
			o.TradingAreaID = &v
		}
		result = append(result, &o)
	}
	return result, rows.Err()
}

func (r *MarketShareRepo) UpdateCompetitionScoreMSGain(ctx context.Context, competitionID uuid.UUID, cc string, msGainPP, hsdGainPP float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE competition_scores
		SET ms_ta_gain_pp = $1, hsd_ta_gain_pp = $2, computed_at = NOW()
		WHERE competition_id = $3 AND cc_number = $4`,
		msGainPP, hsdGainPP, competitionID, cc,
	)
	return err
}

func (r *MarketShareRepo) RecomputeTotalScores(ctx context.Context, competitionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY total_score DESC NULLS LAST) as new_rank
			FROM competition_scores
			WHERE competition_id = $1
		)
		UPDATE competition_scores cs
		SET rank = r.new_rank
		FROM ranked r
		WHERE cs.id = r.id AND cs.competition_id = $1`,
		competitionID,
	)
	return err
}

func (r *MarketShareRepo) GetCountScoresWithNullMSGain(ctx context.Context, competitionID uuid.UUID) (int, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM competition_scores cs
		JOIN retail_outlets ro ON cs.cc_number = ro.cc_number
		WHERE cs.competition_id = $1 AND ro.outlet_type = 'regular' 
		  AND (ms_ta_gain_pp IS NULL OR hsd_ta_gain_pp IS NULL)`,
		competitionID,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MarketShareRepo) GetTotalRegularDealers(ctx context.Context, competitionID uuid.UUID) (int, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM competition_scores cs
		JOIN retail_outlets ro ON cs.cc_number = ro.cc_number
		WHERE cs.competition_id = $1 AND ro.outlet_type = 'regular'`,
		competitionID,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MarketShareRepo) GetActiveCompetitionPeriod(ctx context.Context) (*model.CompetitionPeriod, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, period, territory_code, status, published_at, created_by, created_at
		FROM competition_periods
		WHERE status IN ('active','published')
		ORDER BY period DESC LIMIT 1`)

	var p model.CompetitionPeriod
	var rawID [16]byte
	var publishedAt pgtype.Timestamp
	var createdBy pgtype.UUID

	err := row.Scan(&rawID, &p.Name, &p.Period, &p.TerritoryCode, &p.Status, &publishedAt, &createdBy, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	p.ID = uuid.UUID(rawID)
	if publishedAt.Valid {
		p.PublishedAt = &publishedAt.Time
	}
	if createdBy.Valid {
		id := uuid.UUID(createdBy.Bytes)
		p.CreatedBy = &id
	}
	return &p, nil
}

func (r *MarketShareRepo) GetOutletNameByCC(ctx context.Context, cc string) (string, error) {
	row := r.pool.QueryRow(ctx, `SELECT name FROM retail_outlets WHERE cc_number = $1`, cc)
	var name string
	if err := row.Scan(&name); err != nil {
		if err == pgx.ErrNoRows {
			return cc, nil
		}
		return "", err
	}
	return name, nil
}

func (r *MarketShareRepo) UpdateScoresWithMSGain(ctx context.Context, competitionID uuid.UUID, updates map[string]struct{ MS, HSD float64 }) error {
	batch := &pgx.Batch{}
	for cc, gains := range updates {
		batch.Queue(`
			UPDATE competition_scores
			SET ms_ta_gain_pp = $1, hsd_ta_gain_pp = $2, computed_at = NOW()
			WHERE competition_id = $3 AND cc_number = $4`,
			gains.MS, gains.HSD, competitionID, cc,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range updates {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

type TradingAreaAgg struct {
	TradingAreaID int     `db:"trading_area_id"`
	TotalMS       float64 `db:"total_ms"`
	TotalHSD      float64 `db:"total_hsd"`
	BPCLMS        float64 `db:"bpcl_ms"`
	HPCLMS        float64 `db:"hpcl_ms"`
	IOCLMS        float64 `db:"iocl_ms"`
	BPCLHSD       float64 `db:"bpcl_hsd"`
	HPCLHSD       float64 `db:"hpcl_hsd"`
	IOCLHSD       float64 `db:"iocl_hsd"`
}

func (r *MarketShareRepo) AggregateByTradingArea(ctx context.Context, period time.Time) ([]TradingAreaAgg, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			trading_area_id,
			SUM(ms_vol_kl) as total_ms,
			SUM(hsd_vol_kl) as total_hsd,
			COALESCE(SUM(CASE WHEN omc = 'BPCL' THEN ms_vol_kl ELSE 0 END), 0) as bpcl_ms,
			COALESCE(SUM(CASE WHEN omc = 'HPCL' THEN ms_vol_kl ELSE 0 END), 0) as hpcl_ms,
			COALESCE(SUM(CASE WHEN omc = 'IOCL' THEN ms_vol_kl ELSE 0 END), 0) as iocl_ms,
			COALESCE(SUM(CASE WHEN omc = 'BPCL' THEN hsd_vol_kl ELSE 0 END), 0) as bpcl_hsd,
			COALESCE(SUM(CASE WHEN omc = 'HPCL' THEN hsd_vol_kl ELSE 0 END), 0) as hpcl_hsd,
			COALESCE(SUM(CASE WHEN omc = 'IOCL' THEN hsd_vol_kl ELSE 0 END), 0) as iocl_hsd
		FROM market_share_data
		WHERE period = $1
		GROUP BY trading_area_id`,
		period,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TradingAreaAgg
	for rows.Next() {
		var a TradingAreaAgg
		if err := rows.Scan(&a.TradingAreaID, &a.TotalMS, &a.TotalHSD,
			&a.BPCLMS, &a.HPCLMS, &a.IOCLMS, &a.BPCLHSD, &a.HPCLHSD, &a.IOCLHSD); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (r *MarketShareRepo) GetDealerMSGain(ctx context.Context, cc string, competitionID uuid.UUID) (msGain, hsdGain float64, err error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(ms_ta_gain_pp, 0), COALESCE(hsd_ta_gain_pp, 0)
		FROM competition_scores
		WHERE cc_number = $1 AND competition_id = $2`,
		cc, competitionID,
	)
	err = row.Scan(&msGain, &hsdGain)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return msGain, hsdGain, nil
}

func (r *MarketShareRepo) UpdateParameterScores(ctx context.Context, competitionID uuid.UUID, cc string, msScore, hsdScore float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE competition_scores
		SET score_ms_ta_gain = $1, score_hsd_ta_gain = $2, computed_at = NOW()
		WHERE competition_id = $3 AND cc_number = $4`,
		msScore, hsdScore, competitionID, cc,
	)
	return err
}

func (r *MarketShareRepo) RecalculateTotalScore(ctx context.Context, competitionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE competition_scores
		SET total_score = 
			COALESCE(score_ms_vol, 0) + COALESCE(score_ms_growth, 0) + COALESCE(score_ms_ta_gain, 0) +
			COALESCE(score_hsd_vol, 0) + COALESCE(score_hsd_growth, 0) + COALESCE(score_hsd_ta_gain, 0) +
			COALESCE(score_qoc, 0) + COALESCE(score_lubricants, 0) + COALESCE(score_ufill, 0) +
			COALESCE(score_speed, 0) + COALESCE(score_cleanliness, 0) + COALESCE(score_sangam, 0) +
			COALESCE(score_google, 0) + COALESCE(score_ips, 0) + COALESCE(score_bonus, 0),
			computed_at = NOW()
		WHERE competition_id = $1`,
		competitionID,
	)
	return err
}

func (r *MarketShareRepo) GetCompetitionByID(ctx context.Context, competitionID uuid.UUID) (*model.CompetitionPeriod, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, period, territory_code, status, published_at, created_by, created_at
		FROM competition_periods
		WHERE id = $1`,
		competitionID,
	)

	var p model.CompetitionPeriod
	var rawID [16]byte
	var publishedAt pgtype.Timestamp
	var createdBy pgtype.UUID

	err := row.Scan(&rawID, &p.Name, &p.Period, &p.TerritoryCode, &p.Status, &publishedAt, &createdBy, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	p.ID = uuid.UUID(rawID)
	if publishedAt.Valid {
		p.PublishedAt = &publishedAt.Time
	}
	if createdBy.Valid {
		id := uuid.UUID(createdBy.Bytes)
		p.CreatedBy = &id
	}
	return &p, nil
}

func (r *MarketShareRepo) GetTopScores(ctx context.Context, competitionID uuid.UUID, limit int) ([]model.TopScore, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cs.rank, cs.cc_number, ro.name, cs.total_score
		FROM competition_scores cs
		JOIN retail_outlets ro ON cs.cc_number = ro.cc_number
		WHERE cs.competition_id = $1 AND ro.outlet_type = 'regular'
		ORDER BY cs.rank ASC
		LIMIT $2`,
		competitionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.TopScore
	for rows.Next() {
		var ts model.TopScore
		if err := rows.Scan(&ts.Rank, &ts.CCNumber, &ts.OutletName, &ts.TotalScore); err != nil {
			return nil, err
		}
		result = append(result, ts)
	}
	return result, rows.Err()
}

func (r *MarketShareRepo) CategorizeMissingDealers(ctx context.Context, competitionID uuid.UUID, period time.Time) (missingFromSource, withPeriodGaps int, err error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			CASE WHEN COUNT(m.id) = 0 THEN 'not_in_source'
				 ELSE 'period_gap'
			END as reason
		FROM competition_scores cs
		JOIN retail_outlets ro ON ro.cc_number = cs.cc_number
		LEFT JOIN market_share_data m
			ON m.cc_number = cs.cc_number
			AND m.period = $2
		WHERE cs.competition_id = $1 AND ro.outlet_type = 'regular'
		  AND (cs.ms_ta_gain_pp IS NULL OR cs.hsd_ta_gain_pp IS NULL)
		GROUP BY cs.cc_number
		ORDER BY reason`,
		competitionID, period,
	)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var reason string
		if err := rows.Scan(&reason); err != nil {
			return 0, 0, err
		}
		if reason == "not_in_source" {
			missingFromSource++
		} else {
			withPeriodGaps++
		}
	}
	return missingFromSource, withPeriodGaps, rows.Err()
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}