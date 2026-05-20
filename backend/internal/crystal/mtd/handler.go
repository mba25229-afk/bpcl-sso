package mtd

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type Dealer struct {
	CCCode  string `json:"cc_code"`
	ROName  string `json:"ro_name"`
	Area    string `json:"area"`
}

type MTDRow struct {
	CCCode    string   `json:"cc_code"`
	ROName    string   `json:"ro_name"`
	Ufill     *int     `json:"ufill"`
	QOC       *int     `json:"qoc"`
	SpeedKL   *float64 `json:"speed_kl"`
	MSKL      *float64 `json:"ms_kl"`
	HSDKL     *float64 `json:"hsd_kl"`
	// Targets for colour coding
	UfillTarget *int     `json:"ufill_target"`
	QOCTarget   *int     `json:"qoc_target"`
	SpeedTarget *float64 `json:"speed_target"`
	MSTarget    *float64 `json:"ms_target"`
	HSDTarget   *float64 `json:"hsd_target"`
}

// GET /api/v1/crystal/dealers — list all active dealers (auth required)
func (h *Handler) ListDealers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(),
		`SELECT cc_code, ro_name, area FROM cr_dealers WHERE is_active = TRUE ORDER BY ro_name`,
	)
	if err != nil {
		writeError(w, 500, "failed to list dealers")
		return
	}
	defer rows.Close()

	dealers := []Dealer{}
	for rows.Next() {
		var d Dealer
		if err := rows.Scan(&d.CCCode, &d.ROName, &d.Area); err != nil {
			writeError(w, 500, "scan error")
			return
		}
		dealers = append(dealers, d)
	}
	writeJSON(w, 200, dealers)
}

// GET /api/v1/crystal/dealers/{cc_code}/mtd?date=YYYY-MM-DD
// Returns MTD actuals (sum from month-start to date) plus targets for that month.
func (h *Handler) GetMTD(w http.ResponseWriter, r *http.Request) {
	ccCode := r.PathValue("cc_code")
	if ccCode == "" {
		writeError(w, 400, "cc_code is required")
		return
	}

	dateStr := r.URL.Query().Get("date")
	var asOf time.Time
	if dateStr == "" {
		asOf = time.Now().UTC()
	} else {
		var err error
		asOf, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, 400, "invalid date format, use YYYY-MM-DD")
			return
		}
	}

	monthStart := time.Date(asOf.Year(), asOf.Month(), 1, 0, 0, 0, 0, time.UTC)

	row, err := buildMTD(r.Context(), h.pool, ccCode, monthStart, asOf)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, row)
}

// GET /api/v1/crystal/dashboard?month=YYYY-MM — all dealers MTD for month (auth required)
func (h *Handler) GetMTDDashboard(w http.ResponseWriter, r *http.Request) {
	monthStr := r.URL.Query().Get("month")
	var monthStart time.Time
	if monthStr == "" {
		now := time.Now().UTC()
		monthStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		var err error
		monthStart, err = time.Parse("2006-01", monthStr)
		if err != nil {
			writeError(w, 400, "invalid month format, use YYYY-MM")
			return
		}
	}
	// asOf = last day of month or today, whichever is earlier
	now := time.Now().UTC()
	asOf := time.Date(monthStart.Year(), monthStart.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	if now.Before(asOf) {
		asOf = now
	}

	ccRows, err := h.pool.Query(r.Context(),
		`SELECT cc_code FROM cr_dealers WHERE is_active = TRUE ORDER BY cc_code`,
	)
	if err != nil {
		writeError(w, 500, "failed to list dealers")
		return
	}
	defer ccRows.Close()

	var codes []string
	for ccRows.Next() {
		var cc string
		if err := ccRows.Scan(&cc); err != nil {
			writeError(w, 500, "scan error")
			return
		}
		codes = append(codes, cc)
	}
	ccRows.Close()

	result := make([]MTDRow, 0, len(codes))
	for _, cc := range codes {
		row, err := buildMTD(r.Context(), h.pool, cc, monthStart, asOf)
		if err != nil {
			continue
		}
		result = append(result, *row)
	}
	writeJSON(w, 200, map[string]any{
		"month":   monthStart.Format("2006-01"),
		"as_of":   asOf.Format("2006-01-02"),
		"dealers": result,
	})
}

func buildMTD(ctx context.Context, pool *pgxpool.Pool, ccCode string, monthStart, asOf time.Time) (*MTDRow, error) {
	row := &MTDRow{CCCode: ccCode}

	// RO name
	_ = pool.QueryRow(ctx, `SELECT ro_name FROM cr_dealers WHERE cc_code = $1`, ccCode).Scan(&row.ROName)

	// MTD actuals — sum daily tables
	pool.QueryRow(ctx,
		`SELECT SUM(count) FROM cr_daily_ufill WHERE cc_code=$1 AND txn_date >= $2 AND txn_date <= $3`,
		ccCode, monthStart, asOf,
	).Scan(&row.Ufill)

	pool.QueryRow(ctx,
		`SELECT SUM(count) FROM cr_daily_qoc WHERE cc_code=$1 AND txn_date >= $2 AND txn_date <= $3`,
		ccCode, monthStart, asOf,
	).Scan(&row.QOC)

	pool.QueryRow(ctx,
		`SELECT SUM(kl) FROM cr_daily_speed WHERE cc_code=$1 AND txn_date >= $2 AND txn_date <= $3`,
		ccCode, monthStart, asOf,
	).Scan(&row.SpeedKL)

	pool.QueryRow(ctx,
		`SELECT SUM(kl) FROM cr_daily_ms WHERE cc_code=$1 AND txn_date >= $2 AND txn_date <= $3`,
		ccCode, monthStart, asOf,
	).Scan(&row.MSKL)

	pool.QueryRow(ctx,
		`SELECT SUM(kl) FROM cr_daily_hsd WHERE cc_code=$1 AND txn_date >= $2 AND txn_date <= $3`,
		ccCode, monthStart, asOf,
	).Scan(&row.HSDKL)

	// Targets for the month
	pool.QueryRow(ctx,
		`SELECT ufill_target, qoc_target, speed_kl, ms_kl, hsd_kl
         FROM cr_monthly_targets WHERE cc_code=$1 AND month_year=$2`,
		ccCode, monthStart,
	).Scan(&row.UfillTarget, &row.QOCTarget, &row.SpeedTarget, &row.MSTarget, &row.HSDTarget)

	return row, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
