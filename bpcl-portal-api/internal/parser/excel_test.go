package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bpcl/portal-api/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func makePerformanceXLSX(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "perf.xlsx")

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"product_code", "achieved", "last_year", "volume_kl"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	row := []any{"MS", "500.0", "400.0", "480.5"}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheet, cell, v)
	}
	require.NoError(t, f.SaveAs(path))
	return path
}

func makeDelhiMasterXLSX(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "master.xlsx")

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []any{"Outlet Name", "CC Number", "OMC", "District", "Trading Area", "Apr-24", "May-24"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	data := []any{"TEST OUTLET", "112847", "BPC", "West Delhi", "Kirti Nagar", "150.5", "165.0"}
	for i, v := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheet, cell, v)
	}
	require.NoError(t, f.SaveAs(path))
	return path
}

func TestParsePerformance_Success(t *testing.T) {
	path := makePerformanceXLSX(t)
	rows, err := parser.ParsePerformance(path)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "MS", rows[0].ProductCode)
	assert.Equal(t, "500.0", rows[0].Achieved)
	assert.Equal(t, "400.0", rows[0].LastYear)
	assert.Equal(t, "480.5", rows[0].VolumeKL)
}

func TestParsePerformance_WrongFormat(t *testing.T) {
	// Use a Delhi Master format file where ParsePerformance expects performance format
	path := makeDelhiMasterXLSX(t)
	_, err := parser.ParsePerformance(path)
	assert.Error(t, err)
}

func TestParseDelhiMaster_Success(t *testing.T) {
	path := makeDelhiMasterXLSX(t)
	rows, err := parser.ParseDelhiMaster(path)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "TEST OUTLET", rows[0].OutletName)
	assert.Equal(t, "112847", rows[0].CCNumber)
	assert.Equal(t, "BPC", rows[0].OMC)
	assert.Contains(t, rows[0].MonthlyVolumes, "Apr-24")
	assert.Equal(t, "150.5", rows[0].MonthlyVolumes["Apr-24"])
}

func TestParseDelhiMaster_WrongFormat(t *testing.T) {
	path := makePerformanceXLSX(t)
	_, err := parser.ParseDelhiMaster(path)
	assert.Error(t, err)
}

func TestParsePerformance_UnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	require.NoError(t, os.WriteFile(path, []byte("product_code,achieved\nMS,500\n"), 0644))

	_, err := parser.ParsePerformance(path)
	assert.Error(t, err)
}

func TestParsePerformance_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.xlsx")
	f := excelize.NewFile()
	defer f.Close()
	require.NoError(t, f.SaveAs(path))

	_, err := parser.ParsePerformance(path)
	assert.Error(t, err)
}
