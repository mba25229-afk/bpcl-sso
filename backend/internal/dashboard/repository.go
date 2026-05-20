package dashboard

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

// RepoInterface allows handler tests to inject stubs.
type RepoInterface interface {
	GetDashboard(ctx context.Context, monthYear time.Time) (*DashboardResponse, error)
}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) GetDashboard(ctx context.Context, monthYear time.Time) (*DashboardResponse, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT
            dts.cc_code,
            d.ro_name,
            dts.rank_overall,
            dts.total_marks,
            SUM(sp.max_marks) FILTER (WHERE sp.is_active) AS max_possible,
            MAX(dts.computed_at)
        FROM cr_dealer_total_scores dts
        JOIN cr_dealers d ON d.cc_code = dts.cc_code
        JOIN cr_scoring_params sp ON sp.month_year = dts.month_year
        WHERE dts.month_year = $1
        GROUP BY dts.cc_code, d.ro_name, dts.rank_overall, dts.total_marks, dts.computed_at
        ORDER BY dts.rank_overall`,
		monthYear,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resp := &DashboardResponse{MonthYear: monthYear.Format("2006-01")}
	rowMap := map[string]*DashboardRow{}

	for rows.Next() {
		var dr DashboardRow
		var computedAt time.Time
		if err := rows.Scan(
			&dr.CCCode, &dr.ROName, &dr.Rank,
			&dr.TotalMarks, &dr.MaxPossible, &computedAt,
		); err != nil {
			return nil, err
		}
		if dr.MaxPossible > 0 {
			dr.AchievePct = (dr.TotalMarks / dr.MaxPossible) * 100
		}
		dr.Status = TrafficLight(dr.AchievePct)
		resp.Dealers = append(resp.Dealers, dr)
		rowMap[dr.CCCode] = &resp.Dealers[len(resp.Dealers)-1]
		if computedAt.After(resp.ComputedAt) {
			resp.ComputedAt = computedAt
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	mRows, err := r.pool.Query(ctx, `
        SELECT ds.cc_code, ds.metric_key, sp.display_name,
               ds.marks_scored, ds.max_marks, ds.rank_in_group
        FROM cr_dealer_scores ds
        JOIN cr_scoring_params sp
            ON sp.metric_key = ds.metric_key AND sp.month_year = ds.month_year
        WHERE ds.month_year = $1 AND sp.is_active = TRUE
        ORDER BY ds.cc_code, sp.sort_order`,
		monthYear,
	)
	if err != nil {
		return nil, err
	}
	defer mRows.Close()

	for mRows.Next() {
		var ccCode string
		var item MetricBreakdownItem
		if err := mRows.Scan(
			&ccCode, &item.MetricKey, &item.DisplayName,
			&item.MarksScored, &item.MaxMarks, &item.Rank,
		); err != nil {
			return nil, err
		}
		if dr, ok := rowMap[ccCode]; ok {
			dr.MetricBreakdown = append(dr.MetricBreakdown, item)
		}
	}

	resp.Summary = BuildSummary(resp.Dealers)
	return resp, nil
}

func (r *Repository) ExportXLSX(ctx context.Context, w http.ResponseWriter, monthYear time.Time) error {
	data, err := r.GetDashboard(ctx, monthYear)
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	sheet := "Dealer Rankings"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"003D66"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
	})
	greenStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"DAFBE1"}, Pattern: 1},
	})
	amberStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFF3B0"}, Pattern: 1},
	})
	redStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFEAE6"}, Pattern: 1},
	})

	headers := []string{"Rank", "CC Code", "RO Name", "Total Marks", "Max Possible", "Achievement %", "Status"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	statusLabel := map[string]string{"green": "On Track", "amber": "Watch", "red": "Behind"}
	rowStyleMap := map[string]int{"green": greenStyle, "amber": amberStyle, "red": redStyle}

	for i, d := range data.Dealers {
		row := i + 2
		vals := []interface{}{
			d.Rank, d.CCCode, d.ROName,
			d.TotalMarks, d.MaxPossible,
			d.AchievePct,
			statusLabel[d.Status],
		}
		rowStyle := rowStyleMap[d.Status]
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
			f.SetCellStyle(sheet, cell, cell, rowStyle)
		}
	}

	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "B", 12)
	f.SetColWidth(sheet, "C", "C", 30)
	f.SetColWidth(sheet, "D", "F", 14)

	filename := "BPCL_Rankings_" + monthYear.Format("2006_01") + ".xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	return f.Write(w)
}
