package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RepositoryInterface allows handler tests to inject stubs.
type RepositoryInterface interface {
	ListDealers(ctx context.Context) ([]Dealer, error)
	ToggleDealer(ctx context.Context, ccCode string, active bool) error
	ListPeriods(ctx context.Context) ([]CompetitionPeriod, error)
	CreatePeriod(ctx context.Context, name string, monthYear time.Time) (*CompetitionPeriod, error)
	ActivatePeriod(ctx context.Context, monthYear time.Time) error
	GetTargets(ctx context.Context, monthYear time.Time) ([]TargetRow, error)
	GetIngestSummary(ctx context.Context, metric, table, col string, monthYear time.Time) ([]IngestSummaryRow, error)
	GetMAKGEReadings(ctx context.Context, monthYear time.Time) ([]MAKGERow, error)
	GetManualScores(ctx context.Context, monthYear time.Time) ([]ManualScoreRow, error)
	UpsertManualScores(ctx context.Context, monthYear time.Time, rows []ManualScoreRow) error
	GetLastETLRun(ctx context.Context) (map[string]any, error)
}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) ListDealers(ctx context.Context) ([]Dealer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT cc_code, ro_name, area, is_active, COALESCE(dealer_email,'')
         FROM cr_dealers ORDER BY ro_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Dealer
	for rows.Next() {
		var d Dealer
		if err := rows.Scan(&d.CCCode, &d.ROName, &d.Area, &d.IsActive, &d.DealerEmail); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) ToggleDealer(ctx context.Context, ccCode string, active bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cr_dealers SET is_active = $1, updated_at = NOW() WHERE cc_code = $2`,
		active, ccCode)
	return err
}

func (r *Repository) ListPeriods(ctx context.Context) ([]CompetitionPeriod, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, month_year, total_slots, is_active
         FROM cr_competition_periods ORDER BY month_year DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CompetitionPeriod
	for rows.Next() {
		var p CompetitionPeriod
		if err := rows.Scan(&p.ID, &p.Name, &p.MonthYear, &p.TotalSlots, &p.IsActive); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) CreatePeriod(ctx context.Context, name string, monthYear time.Time) (*CompetitionPeriod, error) {
	var p CompetitionPeriod
	err := r.pool.QueryRow(ctx,
		`INSERT INTO cr_competition_periods (name, month_year, total_slots, is_active)
         VALUES ($1, $2, 40, FALSE)
         RETURNING id, name, month_year, total_slots, is_active`,
		name, monthYear,
	).Scan(&p.ID, &p.Name, &p.MonthYear, &p.TotalSlots, &p.IsActive)
	return &p, err
}

func (r *Repository) ActivatePeriod(ctx context.Context, monthYear time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE cr_competition_periods SET is_active = FALSE`); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx,
		`UPDATE cr_competition_periods SET is_active = TRUE WHERE month_year = $1`, monthYear)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("period %s not found", monthYear.Format("2006-01"))
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetTargets(ctx context.Context, monthYear time.Time) ([]TargetRow, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT d.cc_code, d.ro_name,
               t.ufill_target, t.qoc_target, t.speed_kl, t.ms_kl, t.hsd_kl,
               t.ms_ly, t.hsd_ly, t.mak_ge_target, t.darpan_target,
               t.coolant_lube_kl, t.dsw_available, t.nitrogen
        FROM cr_dealers d
        LEFT JOIN cr_monthly_targets t
            ON t.cc_code = d.cc_code AND t.month_year = $1
        WHERE d.is_active = TRUE
        ORDER BY d.ro_name`, monthYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TargetRow
	for rows.Next() {
		var t TargetRow
		if err := rows.Scan(
			&t.CCCode, &t.ROName,
			&t.UfillTarget, &t.QOCTarget, &t.SpeedKL, &t.MSKL, &t.HSDKL,
			&t.MSLY, &t.HSDLY, &t.MAKGETarget, &t.DarpanTarget,
			&t.CoolantLubeKL, &t.DSWAvailable, &t.Nitrogen,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// metricTableMap maps metric keys to their cr_-prefixed tables and value columns.
var metricTableMap = map[string]struct{ table, col string }{
	"ms":    {"cr_daily_ms", "kl"},
	"hsd":   {"cr_daily_hsd", "kl"},
	"speed": {"cr_daily_speed", "kl"},
	"ufill": {"cr_daily_ufill", "count"},
	"qoc":   {"cr_daily_qoc", "count"},
}

func MetricTable(metric string) (table, col string, ok bool) {
	m, ok := metricTableMap[metric]
	return m.table, m.col, ok
}

func (r *Repository) GetIngestSummary(ctx context.Context, metric, table, col string, monthYear time.Time) ([]IngestSummaryRow, error) {
	sql := fmt.Sprintf(`
        SELECT d.cc_code, d.ro_name,
               COALESCE(SUM(s.%s), 0)   AS mtd_sum,
               COALESCE(COUNT(s.%s), 0) AS days_filled
        FROM cr_dealers d
        LEFT JOIN %s s ON s.cc_code = d.cc_code
            AND DATE_TRUNC('month', s.txn_date) = $1
        WHERE d.is_active = TRUE
        GROUP BY d.cc_code, d.ro_name
        ORDER BY d.ro_name`, col, col, table)
	rows, err := r.pool.Query(ctx, sql, monthYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []IngestSummaryRow
	for rows.Next() {
		var s IngestSummaryRow
		if err := rows.Scan(&s.CCCode, &s.ROName, &s.MTDSum, &s.DaysFilled); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) GetMAKGEReadings(ctx context.Context, monthYear time.Time) ([]MAKGERow, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT d.cc_code, d.ro_name,
               COALESCE(g.reading_date::text, ''), COALESCE(g.meter_reading, 0)
        FROM cr_dealers d
        LEFT JOIN cr_mak_ge_readings g ON g.cc_code = d.cc_code
            AND DATE_TRUNC('month', g.reading_date) = $1
        WHERE d.is_active = TRUE
        ORDER BY d.ro_name, g.reading_date`, monthYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MAKGERow
	for rows.Next() {
		var m MAKGERow
		if err := rows.Scan(&m.CCCode, &m.ROName, &m.ReadingDate, &m.MeterReading); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetManualScores fetches bonus and cleanliness scores from cr_dealer_scores.
// bonus_marks → metric_key = 'bonus', cleanliness_grade → metric_key = 'cleanliness_audit'
func (r *Repository) GetManualScores(ctx context.Context, monthYear time.Time) ([]ManualScoreRow, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT d.cc_code, d.ro_name,
               b.actual_value  AS bonus_marks,
               c.actual_value  AS cleanliness_actual
        FROM cr_dealers d
        LEFT JOIN cr_dealer_scores b
            ON b.cc_code = d.cc_code AND b.month_year = $1 AND b.metric_key = 'bonus'
        LEFT JOIN cr_dealer_scores c
            ON c.cc_code = d.cc_code AND c.month_year = $1 AND c.metric_key = 'cleanliness_audit'
        WHERE d.is_active = TRUE
        ORDER BY d.ro_name`, monthYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ManualScoreRow
	for rows.Next() {
		var s ManualScoreRow
		var cleanlinessVal *float64
		if err := rows.Scan(&s.CCCode, &s.ROName, &s.BonusMarks, &cleanlinessVal); err != nil {
			return nil, err
		}
		// Cleanliness is stored as numeric grade (1=Excellent..5=Poor); return as string label
		if cleanlinessVal != nil {
			grade := gradeLabel(*cleanlinessVal)
			s.CleanlinessGrade = &grade
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func gradeLabel(v float64) string {
	switch int(v) {
	case 1:
		return "Excellent"
	case 2:
		return "Good"
	case 3:
		return "Average"
	case 4:
		return "Below Average"
	default:
		return "Poor"
	}
}

// UpsertManualScores upserts bonus and cleanliness metric rows into cr_dealer_scores.
func (r *Repository) UpsertManualScores(ctx context.Context, monthYear time.Time, rows []ManualScoreRow) error {
	batch := &pgx.Batch{}
	for _, row := range rows {
		if row.BonusMarks != nil {
			batch.Queue(`
                INSERT INTO cr_dealer_scores (cc_code, month_year, metric_key, actual_value, max_marks)
                VALUES ($1, $2, 'bonus', $3, 10)
                ON CONFLICT (cc_code, month_year, metric_key) DO UPDATE SET
                    actual_value = EXCLUDED.actual_value,
                    computed_at  = NOW()`,
				row.CCCode, monthYear, row.BonusMarks)
		}
		if row.CleanlinessGrade != nil {
			numeric := gradeNumeric(*row.CleanlinessGrade)
			batch.Queue(`
                INSERT INTO cr_dealer_scores (cc_code, month_year, metric_key, actual_value, max_marks)
                VALUES ($1, $2, 'cleanliness_audit', $3, 10)
                ON CONFLICT (cc_code, month_year, metric_key) DO UPDATE SET
                    actual_value = EXCLUDED.actual_value,
                    computed_at  = NOW()`,
				row.CCCode, monthYear, numeric)
		}
	}
	if batch.Len() == 0 {
		return nil
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func gradeNumeric(grade string) float64 {
	switch grade {
	case "Excellent":
		return 1
	case "Good":
		return 2
	case "Average":
		return 3
	case "Below Average":
		return 4
	default:
		return 5
	}
}

type ETLRLogRow struct {
	ID        string    `json:"id"`
	RunAt     time.Time `json:"run_at"`
	Status    string    `json:"status"`
	Detail    string    `json:"detail"`
	Duration  int       `json:"duration_ms"`
	Source    string    `json:"source"`
}

func (r *Repository) GetLastETLRun(ctx context.Context) (map[string]any, error) {
	var row ETLRLogRow
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, run_at, status, COALESCE(detail,''), COALESCE(duration_ms,0), COALESCE(source,'')
		FROM cr_etl_log
		ORDER BY run_at DESC
		LIMIT 1
	`).Scan(&row.ID, &row.RunAt, &row.Status, &row.Detail, &row.Duration, &row.Source)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return map[string]any{"status": "never_run", "message": "No ETL runs recorded"}, nil
		}
		return nil, err
	}
	return map[string]any{
		"id":         row.ID,
		"run_at":     row.RunAt.Format(time.RFC3339),
		"status":     row.Status,
		"detail":     row.Detail,
		"duration":   row.Duration,
		"source":     row.Source,
	}, nil
}
