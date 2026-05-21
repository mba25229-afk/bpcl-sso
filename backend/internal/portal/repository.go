package portal

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RepositoryInterface allows handler tests to inject stubs.
type RepositoryInterface interface {
	GetScoreCard(ctx context.Context, ccCode string, monthYear time.Time) (*ScoreCard, error)
}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) GetScoreCard(ctx context.Context, ccCode string, monthYear time.Time) (*ScoreCard, error) {
	card := &ScoreCard{}

	err := r.pool.QueryRow(ctx,
		`SELECT cc_code, ro_name, area FROM cr_dealers WHERE cc_code = $1`, ccCode,
	).Scan(&card.Dealer.CCCode, &card.Dealer.ROName, &card.Dealer.Area)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
        SELECT
            ds.metric_key,
            sp.display_name,
            ds.actual_value,
            ds.rank_in_group,
            ds.marks_scored,
            ds.max_marks,
            CASE ds.metric_key
                WHEN 'ms_absolute_vol'  THEN t.ms_kl
                WHEN 'hsd_absolute_vol' THEN t.hsd_kl
                WHEN 'speed_vol'        THEN t.speed_kl
                WHEN 'ufill_txns'       THEN t.ufill_target::numeric
                WHEN 'oil_change_count' THEN t.qoc_target::numeric
                WHEN 'mak_ge_sales'     THEN t.mak_ge_target
                ELSE NULL
            END AS target_value
        FROM cr_dealer_scores ds
        JOIN cr_scoring_params sp
            ON sp.metric_key = ds.metric_key AND sp.month_year = ds.month_year
        LEFT JOIN cr_monthly_targets t
            ON t.cc_code = ds.cc_code AND t.month_year = ds.month_year
        WHERE ds.cc_code = $1 AND ds.month_year = $2
        ORDER BY sp.sort_order`,
		ccCode, monthYear,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	unitMap := map[string]string{
		"ms_absolute_vol": "KL", "ms_growth_pct": "%",
		"hsd_absolute_vol": "KL", "hsd_growth_pct": "%",
		"oil_change_count": "count", "mak_ge_sales": "KL",
		"ufill_txns": "count", "speed_vol": "KL",
		"cleanliness_audit": "grade", "sangam_certs": "count",
		"google_rating": "stars", "ips_pct": "%", "bonus": "pts",
	}

	var totalMarks, maxPossible float64
	for rows.Next() {
		var m MetricCard
		var target *float64
		if err := rows.Scan(
			&m.MetricKey, &m.DisplayName,
			&m.ActualValue, &m.RankInGroup, &m.MarksScored, &m.MaxMarks,
			&target,
		); err != nil {
			return nil, err
		}
		m.Target = target
		if target != nil && *target > 0 && m.ActualValue != nil {
			pct := (*m.ActualValue / *target) * 100
			m.AchievePct = &pct
		}
		m.Unit = unitMap[m.MetricKey]
		maxPossible += m.MaxMarks // always count max, regardless of whether scored
		if m.MarksScored != nil {
			totalMarks += *m.MarksScored
		}
		card.Metrics = append(card.Metrics, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	card.TotalMarks = totalMarks
	card.MaxPossible = maxPossible

	r.pool.QueryRow(ctx,
		`SELECT rank_overall,
                (SELECT COUNT(*) FROM cr_dealer_total_scores WHERE month_year = $2) AS total
         FROM cr_dealer_total_scores
         WHERE cc_code = $1 AND month_year = $2`,
		ccCode, monthYear,
	).Scan(&card.RankOverall, &card.TotalDealers)

	dRows, err := r.pool.Query(ctx, `
        WITH dates AS (
            SELECT generate_series(
                DATE_TRUNC('month', $2::date),
                $2::date,
                '1 day'::interval
            )::date AS d
        )
        SELECT
            dates.d::text,
            COALESCE(ms.kl,0) + COALESCE(sp.kl,0) AS ms_kl,
            COALESCE(hsd.kl,0) AS hsd_kl,
            COALESCE(qoc.count,0) AS qoc,
            COALESCE(uf.count,0)  AS ufill
        FROM dates
        LEFT JOIN cr_daily_ms    ms  ON ms.cc_code = $1 AND ms.txn_date = dates.d
        LEFT JOIN cr_daily_speed sp  ON sp.cc_code = $1 AND sp.txn_date = dates.d
        LEFT JOIN cr_daily_hsd   hsd ON hsd.cc_code = $1 AND hsd.txn_date = dates.d
        LEFT JOIN cr_daily_qoc   qoc ON qoc.cc_code = $1 AND qoc.txn_date = dates.d
        LEFT JOIN cr_daily_ufill uf  ON uf.cc_code  = $1 AND uf.txn_date = dates.d
        ORDER BY dates.d`,
		ccCode, monthYear,
	)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()
	for dRows.Next() {
		var p DailyPoint
		if err := dRows.Scan(&p.Date, &p.MSKL, &p.HSDKL, &p.QOC, &p.UFill); err != nil {
			return nil, err
		}
		card.DailyTrend = append(card.DailyTrend, p)
	}

	gRows, err := r.pool.Query(ctx, `
        SELECT reading_date::text, meter_reading,
               COALESCE(meter_reading - LAG(meter_reading) OVER (ORDER BY reading_date), 0) AS delta
        FROM cr_mak_ge_readings
        WHERE cc_code = $1 AND DATE_TRUNC('month', reading_date) = $2
        ORDER BY reading_date`,
		ccCode, monthYear,
	)
	if err != nil {
		return nil, err
	}
	defer gRows.Close()
	for gRows.Next() {
		var e MAKGEEntry
		if err := gRows.Scan(&e.ReadingDate, &e.MeterReading, &e.Delta); err != nil {
			return nil, err
		}
		card.MAKGELog = append(card.MAKGELog, e)
	}

	grRows, err := r.pool.Query(ctx, `
        SELECT snapshot_date::text, rating, review_count
        FROM cr_google_ratings
        WHERE cc_code = $1 AND DATE_TRUNC('month', snapshot_date) = $2
        ORDER BY snapshot_date`,
		ccCode, monthYear,
	)
	if err != nil {
		return nil, err
	}
	defer grRows.Close()
	for grRows.Next() {
		var e GoogleEntry
		if err := grRows.Scan(&e.SnapshotDate, &e.Rating, &e.ReviewCount); err != nil {
			return nil, err
		}
		card.GoogleLog = append(card.GoogleLog, e)
	}

	card.MonthYear = monthYear.Format("2006-01")
	return card, nil
}
