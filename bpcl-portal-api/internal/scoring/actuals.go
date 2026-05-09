package scoring

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ActualsRepo struct {
	pool *pgxpool.Pool
}

func NewActualsRepo(pool *pgxpool.Pool) *ActualsRepo {
	return &ActualsRepo{pool: pool}
}

func (r *ActualsRepo) FetchAll(ctx context.Context, monthYear, asOf time.Time) ([]Actuals, error) {
	sql := `
WITH
ms_mtd AS (
    SELECT cc_code, COALESCE(SUM(kl), 0) AS kl
    FROM cr_daily_ms
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
speed_mtd AS (
    SELECT cc_code, COALESCE(SUM(kl), 0) AS kl
    FROM cr_daily_speed
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
hsd_mtd AS (
    SELECT cc_code, COALESCE(SUM(kl), 0) AS kl
    FROM cr_daily_hsd
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
qoc_mtd AS (
    SELECT cc_code, COALESCE(SUM(count), 0) AS cnt
    FROM cr_daily_qoc
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
ufill_mtd AS (
    SELECT cc_code, COALESCE(SUM(count), 0) AS cnt
    FROM cr_daily_ufill
    WHERE DATE_TRUNC('month', txn_date) = $1 AND txn_date <= $2
    GROUP BY cc_code
),
mak_ge_delta AS (
    SELECT
        cc_code,
        MAX(meter_reading) - MIN(meter_reading) AS sales_kl,
        COUNT(*) AS reading_count
    FROM cr_mak_ge_readings
    WHERE DATE_TRUNC('month', reading_date) = $1 AND reading_date <= $2
    GROUP BY cc_code
),
google_latest AS (
    SELECT DISTINCT ON (cc_code)
        cc_code,
        rating,
        review_count,
        rating * review_count AS composite
    FROM cr_google_ratings
    WHERE DATE_TRUNC('month', snapshot_date) = $1 AND snapshot_date <= $2
    ORDER BY cc_code, snapshot_date DESC
),
sangam_mtd AS (
    SELECT cc_code, COALESCE(cert_count, 0) AS cert_count
    FROM cr_sangam_data
    WHERE month_year = $1
),
manual_bonus AS (
    SELECT cc_code, COALESCE(actual_value, 0) AS bonus_marks
    FROM cr_dealer_scores
    WHERE month_year = $1 AND metric_key = 'bonus'
),
manual_clean AS (
    SELECT cc_code, COALESCE(actual_value::int, 0) AS cleanliness_val
    FROM cr_dealer_scores
    WHERE month_year = $1 AND metric_key = 'cleanliness_audit'
)
SELECT
    d.cc_code,
    COALESCE(ms.kl,  0)                        AS ms_kl,
    COALESCE(sp.kl,  0)                        AS speed_kl,
    COALESCE(hsd.kl, 0)                        AS hsd_kl,
    -- Prorate monthly LY to match days elapsed so growth % is apples-to-apples.
    -- days_elapsed = $2 - DATE_TRUNC('month',$2) + 1, days_in_month = EXTRACT(days FROM last_day)
    COALESCE(t.ms_ly,  0) * (($2::date - DATE_TRUNC('month',$2)::date + 1)::numeric
        / EXTRACT(DAY FROM (DATE_TRUNC('month',$2) + INTERVAL '1 month - 1 day'))::numeric) AS ms_ly,
    COALESCE(t.hsd_ly, 0) * (($2::date - DATE_TRUNC('month',$2)::date + 1)::numeric
        / EXTRACT(DAY FROM (DATE_TRUNC('month',$2) + INTERVAL '1 month - 1 day'))::numeric) AS hsd_ly,
    COALESCE(qoc.cnt, 0)                       AS qoc_count,
    COALESCE(uf.cnt,  0)                       AS ufill_count,
    COALESCE(ge.sales_kl, 0)                   AS mak_ge_sales,
    COALESCE(ge.reading_count, 0) >= 2         AS mak_ge_has_data,
    COALESCE(gr.rating, 0)                     AS google_rating,
    COALESCE(gr.composite, 0)                  AS google_composite,
    COALESCE(sg.cert_count, 0)                 AS sangam_certs,
    COALESCE(mb.bonus_marks, 0)                AS bonus_marks,
    CASE COALESCE(mc.cleanliness_val, 0)
        WHEN 1 THEN 'Excellent'
        WHEN 2 THEN 'Good'
        WHEN 3 THEN 'Average'
        WHEN 4 THEN 'Below Average'
        WHEN 5 THEN 'Poor'
        ELSE ''
    END                                        AS cleanliness_grade
FROM cr_dealers d
LEFT JOIN ms_mtd        ms  ON ms.cc_code  = d.cc_code
LEFT JOIN speed_mtd     sp  ON sp.cc_code  = d.cc_code
LEFT JOIN hsd_mtd       hsd ON hsd.cc_code = d.cc_code
LEFT JOIN qoc_mtd       qoc ON qoc.cc_code = d.cc_code
LEFT JOIN ufill_mtd     uf  ON uf.cc_code  = d.cc_code
LEFT JOIN mak_ge_delta  ge  ON ge.cc_code  = d.cc_code
LEFT JOIN google_latest gr  ON gr.cc_code  = d.cc_code
LEFT JOIN sangam_mtd    sg  ON sg.cc_code  = d.cc_code
LEFT JOIN cr_monthly_targets t ON t.cc_code = d.cc_code AND t.month_year = $1
LEFT JOIN manual_bonus  mb  ON mb.cc_code  = d.cc_code
LEFT JOIN manual_clean  mc  ON mc.cc_code  = d.cc_code
WHERE d.is_active = TRUE
	`

	rows, err := r.pool.Query(ctx, sql, monthYear, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Actuals
	for rows.Next() {
		var a Actuals
		if err := rows.Scan(
			&a.CCCode,
			&a.MSVolKL, &a.SpeedVolKL, &a.HSDVolKL,
			&a.MSLYKl, &a.HSDLYKl,
			&a.QOCCount, &a.UFillCount,
			&a.MAKGESalesKL, &a.MAKGEHasData,
			&a.GoogleRating, &a.GoogleComposite,
			&a.SangamCerts,
			&a.BonusMarks, &a.CleanlinessGrade,
		); err != nil {
			return nil, err
		}
		a.MSVolKL = a.MSVolKL + a.SpeedVolKL
		out = append(out, a)
	}
	return out, rows.Err()
}
