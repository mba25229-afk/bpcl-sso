package etl

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TargetRecord struct {
	CCNumber    string
	ProductID   int16
	Period      time.Time
	TargetValue float64
}

type Transformer struct {
	pool   *pgxpool.Pool
	period time.Time
}

func NewTransformer(pool *pgxpool.Pool, period time.Time) *Transformer {
	return &Transformer{
		pool:   pool,
		period: period,
	}
}

var productCodeToID = map[string]int16{
	"MS":         1,
	"HSD":        2,
	"SPEED":      3,
	"SPEEDDIESEL": 3,
	"QOC":        4,
	"LUBRICANTS": 5,
	"OIL":        5,
	"OILCHANGES": 5,
	"UFILL":      6,
	"UFILLNOS":   6,
	"DSW":        9,
	"NITROGEN":   10,
	"MAKGE":      11,
}

func (t *Transformer) Transform(ctx context.Context, rawRows []RawTargetRow) ([]TargetRecord, error) {
	_, err := t.loadProductMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("transformer: failed to load product map: %w", err)
	}

	var records []TargetRecord
	for _, row := range rawRows {
		ccNumber := strings.ToUpper(strings.TrimSpace(row.CCCode))
		if ccNumber == "" {
			continue
		}

		if row.UFill != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   6,
				Period:      t.period,
				TargetValue: floatOrZero(row.UFill),
			})
		}

		if row.OilChg != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   5,
				Period:      t.period,
				TargetValue: floatOrZero(row.OilChg),
			})
		}

		if row.Speed != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   3,
				Period:      t.period,
				TargetValue: floatOrZero(row.Speed),
			})
		}

		if row.MS != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   1,
				Period:      t.period,
				TargetValue: floatOrZero(row.MS),
			})
		}

		if row.HSD != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   2,
				Period:      t.period,
				TargetValue: floatOrZero(row.HSD),
			})
		}

		if row.DSW != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   9,
				Period:      t.period,
				TargetValue: floatOrZero(row.DSW),
			})
		}

		if row.Nitrogen != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   10,
				Period:      t.period,
				TargetValue: floatOrZero(row.Nitrogen),
			})
		}

		if row.MAKGE != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   11,
				Period:      t.period,
				TargetValue: floatOrZero(row.MAKGE),
			})
		}

		if row.OtherLube != "" {
			records = append(records, TargetRecord{
				CCNumber:    ccNumber,
				ProductID:   5,
				Period:      t.period,
				TargetValue: floatOrZero(row.OtherLube),
			})
		}
	}

	return records, nil
}

func (t *Transformer) loadProductMap(ctx context.Context) (map[string]int16, error) {
	rows, err := t.pool.Query(ctx, "SELECT id, code FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int16)
	for rows.Next() {
		var id int16
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			continue
		}
		result[code] = id
		result[strings.ToUpper(code)] = id
	}

	return result, rows.Err()
}

func floatOrZero(s string) float64 {
	if s == "" {
		return 0
	}
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func init() {
	for code, id := range productCodeToID {
		productCodeToID[strings.ToUpper(code)] = id
	}
}