package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutletRepo struct {
	pool *pgxpool.Pool
}

func NewOutletRepo(pool *pgxpool.Pool) *OutletRepo {
	return &OutletRepo{pool: pool}
}

const outletSelect = `SELECT cc_number, name, location, district, state, rank,
	territory_code, trading_area_id, outlet_type, ro_manager_id, is_active,
	created_at, updated_at FROM retail_outlets`

func (r *OutletRepo) GetByCC(ctx context.Context, cc string) (*model.RetailOutlet, error) {
	row := r.pool.QueryRow(ctx, outletSelect+` WHERE cc_number = $1`, cc)
	o, err := scanOutlet(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return o, nil
}

func (r *OutletRepo) ListByTerritory(ctx context.Context, territoryCode string) ([]*model.RetailOutlet, error) {
	rows, err := r.pool.Query(ctx, outletSelect+` WHERE territory_code = $1 ORDER BY name`, territoryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectOutlets(rows)
}

func (r *OutletRepo) ListAll(ctx context.Context) ([]*model.RetailOutlet, error) {
	rows, err := r.pool.Query(ctx, outletSelect+` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectOutlets(rows)
}

func scanOutlet(row pgx.Row) (*model.RetailOutlet, error) {
	var o model.RetailOutlet
	var location, district, state, rank, territory pgtype.Text
	var tradingAreaID pgtype.Int4
	var roManagerID pgtype.UUID

	err := row.Scan(
		&o.CCNumber, &o.Name, &location, &district, &state, &rank,
		&territory, &tradingAreaID, &o.OutletType, &roManagerID,
		&o.IsActive, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if location.Valid {
		o.Location = &location.String
	}
	if district.Valid {
		o.District = &district.String
	}
	if state.Valid {
		o.State = &state.String
	}
	if rank.Valid {
		o.Rank = &rank.String
	}
	if territory.Valid {
		o.TerritoryCode = &territory.String
	}
	if tradingAreaID.Valid {
		v := int(tradingAreaID.Int32)
		o.TradingAreaID = &v
	}
	o.RoManagerID = pguuid(roManagerID)
	return &o, nil
}

func collectOutlets(rows pgx.Rows) ([]*model.RetailOutlet, error) {
	var result []*model.RetailOutlet
	for rows.Next() {
		o, err := scanOutlet(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func (r *OutletRepo) GetByTerritoryWithPerformance(ctx context.Context, territoryCode string, period time.Time) ([]*model.TerritoryOutlet, error) {
	query := `
		SELECT 
			o.cc_number,
			o.name,
			o.location,
			o.district,
			o.state,
			o.rank,
			o.territory_code,
			COALESCE(pr_ms.achieved, 0) as ms_achieved,
			COALESCE(pr_hsd.achieved, 0) as hsd_achieved,
			COALESCE(pr_speed.achieved, 0) as speed_achieved,
			CASE WHEN t_ms.target_value > 0 THEN (pr_ms.achieved / t_ms.target_value * 100)::numeric(5,1) ELSE 0 END as ms_achievement_pct,
			CASE WHEN t_hsd.target_value > 0 THEN (pr_hsd.achieved / t_hsd.target_value * 100)::numeric(5,1) ELSE 0 END as hsd_achievement_pct
		FROM retail_outlets o
		LEFT JOIN LATERAL (
			SELECT achieved FROM performance_records 
			WHERE cc_number = o.cc_number AND product_id = 1 AND period = $2
		) pr_ms ON true
		LEFT JOIN LATERAL (
			SELECT achieved FROM performance_records 
			WHERE cc_number = o.cc_number AND product_id = 2 AND period = $2
		) pr_hsd ON true
		LEFT JOIN LATERAL (
			SELECT achieved FROM performance_records 
			WHERE cc_number = o.cc_number AND product_id = 3 AND period = $2
		) pr_speed ON true
		LEFT JOIN LATERAL (
			SELECT target_value FROM targets 
			WHERE cc_number = o.cc_number AND product_id = 1 AND period = $2
		) t_ms ON true
		LEFT JOIN LATERAL (
			SELECT target_value FROM targets 
			WHERE cc_number = o.cc_number AND product_id = 2 AND period = $2
		) t_hsd ON true
		WHERE o.territory_code = $1 AND o.outlet_type = 'regular'
		ORDER BY o.name`

	rows, err := r.pool.Query(ctx, query, territoryCode, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.TerritoryOutlet
	for rows.Next() {
		var o model.TerritoryOutlet
		var location, district, state, rank pgtype.Text
		var territory pgtype.Text
		var msAchieved, hsdAchieved, speedAchieved, msPct, hsdPct pgtype.Numeric

		err := rows.Scan(
			&o.CCNumber, &o.Name, &location, &district, &state, &rank, &territory,
			&msAchieved, &hsdAchieved, &speedAchieved, &msPct, &hsdPct,
		)
		if err != nil {
			return nil, err
		}
		if location.Valid {
			o.Location = &location.String
		}
		if district.Valid {
			o.District = &district.String
		}
		if state.Valid {
			o.State = &state.String
		}
		if rank.Valid {
			o.Rank = &rank.String
		}
		if territory.Valid {
			o.TerritoryCode = &territory.String
		}
		o.MSAchieved = numericToFloat64(msAchieved)
		o.HSDAchieved = numericToFloat64(hsdAchieved)
		o.SpeedAchieved = numericToFloat64(speedAchieved)
		o.MSAchievementPct = numericToFloat64(msPct)
		o.HSDAchievementPct = numericToFloat64(hsdPct)

		o.TotalAchievementPct = (o.MSAchievementPct + o.HSDAchievementPct) / 2
		result = append(result, &o)
	}
	return result, rows.Err()
}
