package etl

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bpcl/etl/internal/microsoft"
)

type RawTargetRow struct {
	ROName   string
	CCCode   string
	UFill    string
	OilChg   string
	Speed    string
	MS       string
	HSD      string
	DSW      string
	Nitrogen string
	MAKGE    string
	OtherLube string
}

type Extractor struct {
	sheetsClient *microsoft.Client
	sheetName    string
	startRow     int
	endRow       int
}

func NewExtractor(client *microsoft.Client, sheetName string, startRow, endRow int) *Extractor {
	return &Extractor{
		sheetsClient: client,
		sheetName:    sheetName,
		startRow:     startRow,
		endRow:       endRow,
	}
}

func (e *Extractor) Extract(ctx context.Context) ([]RawTargetRow, error) {
	rows, err := e.sheetsClient.ReadSheet(ctx, e.sheetName, e.startRow, e.endRow)
	if err != nil {
		return nil, fmt.Errorf("extractor: failed to read sheet: %w", err)
	}

	var result []RawTargetRow
	for _, row := range rows {
		if len(row.Values) < 2 {
			continue
		}

		ccCode := strings.TrimSpace(getValue(row.Values, 1))
		if ccCode == "" {
			continue
		}

		targetRow := RawTargetRow{
			ROName:    strings.TrimSpace(getValue(row.Values, 0)),
			CCCode:    ccCode,
			UFill:     strings.TrimSpace(getValue(row.Values, 2)),
			OilChg:    strings.TrimSpace(getValue(row.Values, 3)),
			Speed:     strings.TrimSpace(getValue(row.Values, 4)),
			MS:        strings.TrimSpace(getValue(row.Values, 5)),
			HSD:       strings.TrimSpace(getValue(row.Values, 6)),
			DSW:       strings.TrimSpace(getValue(row.Values, 9)),
			Nitrogen:  strings.TrimSpace(getValue(row.Values, 10)),
			MAKGE:     strings.TrimSpace(getValue(row.Values, 12)),
			OtherLube: strings.TrimSpace(getValue(row.Values, 14)),
		}

		result = append(result, targetRow)
	}

	return result, nil
}

func getValue(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return values[index]
}

func ParseFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}