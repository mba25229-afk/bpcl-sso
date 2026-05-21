// Package parser handles Excel/CSV file parsing for BPCL portal uploads.
// Two formats are supported (auto-detected from row 1 headers):
//   - Performance upload: col A header = "product_code"
//   - Delhi Master:       col A header = "Outlet Name"
package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// RawPerformanceRow holds a single row from a standard performance upload file.
type RawPerformanceRow struct {
	ProductCode string
	Achieved    string
	LastYear    string
	VolumeKL    string
	Skipped     bool
}

// RawMarketShareRow holds a single row from a Delhi Master file.
type RawMarketShareRow struct {
	OutletName  string
	CCNumber    string
	OMC         string
	District    string
	TradingArea string
	// MonthlyVolumes: map from period string (e.g. "Apr-24") to raw volume string
	MonthlyVolumes map[string]string
}

// ParsePerformance parses a standard performance upload file.
// Invalid rows are skipped and counted; the whole file is never aborted for row errors.
func ParsePerformance(filePath string) ([]RawPerformanceRow, error) {
	f, err := openFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("parser: read rows: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("parser: file has no data rows")
	}

	// Detect format
	if len(rows[0]) == 0 {
		return nil, fmt.Errorf("parser: empty header row")
	}
	if strings.TrimSpace(rows[0][0]) != "product_code" {
		return nil, fmt.Errorf("parser: expected performance format (col A = 'product_code'), got %q", rows[0][0])
	}

	// Map header column names to indices
	colIndex := make(map[string]int, len(rows[0]))
	for i, h := range rows[0] {
		colIndex[strings.TrimSpace(strings.ToLower(h))] = i
	}

	var result []RawPerformanceRow
	for _, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}
		code := cellValue(row, colIndex["product_code"])
		if code == "" {
			continue
		}
		result = append(result, RawPerformanceRow{
			ProductCode: code,
			Achieved:    cellValue(row, colIndex["achieved"]),
			LastYear:    cellValue(row, colIndex["last_year"]),
			VolumeKL:    cellValue(row, colIndex["volume_kl"]),
		})
	}
	return result, nil
}

// ParseDelhiMaster parses a Delhi Master format file.
// The header row must have "Outlet Name" in column A.
// Monthly volume columns are detected by their "Mon-YY" header format.
func ParseDelhiMaster(filePath string) ([]RawMarketShareRow, error) {
	f, err := openFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("parser: read rows: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("parser: file has no data rows")
	}

	if len(rows[0]) == 0 || strings.TrimSpace(rows[0][0]) != "Outlet Name" {
		return nil, fmt.Errorf("parser: expected Delhi Master format (col A = 'Outlet Name'), got %q", rows[0][0])
	}

	// Identify monthly volume columns by "Mon-YY" pattern (e.g. "Apr-24")
	monthCols := make(map[int]string) // colIndex → period string
	for i, h := range rows[0] {
		h = strings.TrimSpace(h)
		if isMonthHeader(h) {
			monthCols[i] = h
		}
	}

	var result []RawMarketShareRow
	for _, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}
		outletName := cellValue(row, 0)
		if outletName == "" {
			continue
		}
		monthly := make(map[string]string, len(monthCols))
		for idx, period := range monthCols {
			v := cellValue(row, idx)
			if v != "" {
				monthly[period] = v
			}
		}
		result = append(result, RawMarketShareRow{
			OutletName:     outletName,
			CCNumber:       cellValue(row, 1),
			OMC:            cellValue(row, 2),
			District:       cellValue(row, 3),
			TradingArea:    cellValue(row, 4),
			MonthlyVolumes: monthly,
		})
	}
	return result, nil
}

func openFile(filePath string) (*excelize.File, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".xlsx" && ext != ".xls" {
		return nil, fmt.Errorf("parser: unsupported extension %q (use .xlsx or .xls)", ext)
	}
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parser: open %q: %w", filePath, err)
	}
	return f, nil
}

func cellValue(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// isMonthHeader returns true for strings like "Apr-24", "May-24", "Jan-25".
var months = map[string]bool{
	"jan": true, "feb": true, "mar": true, "apr": true,
	"may": true, "jun": true, "jul": true, "aug": true,
	"sep": true, "oct": true, "nov": true, "dec": true,
}

func isMonthHeader(h string) bool {
	parts := strings.SplitN(h, "-", 2)
	if len(parts) != 2 {
		return false
	}
	if !months[strings.ToLower(parts[0])] {
		return false
	}
	if len(parts[1]) != 2 {
		return false
	}
	for _, c := range parts[1] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
