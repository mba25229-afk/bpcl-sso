package model

import (
	"time"

	"github.com/google/uuid"
)

type RetailOutlet struct {
	CCNumber      string     `json:"cc_number"`
	Name          string     `json:"name"`
	Location      *string    `json:"location"`
	District      *string    `json:"district"`
	State         *string    `json:"state"`
	Rank          *string    `json:"rank"`
	TerritoryCode *string    `json:"territory_code"`
	TradingAreaID *int       `json:"trading_area_id"`
	OutletType    string     `json:"outlet_type"`
	RoManagerID   *uuid.UUID `json:"ro_manager_id"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type TerritoryOutlet struct {
	CCNumber             string  `json:"cc_number"`
	Name                 string  `json:"name"`
	Location             *string `json:"location"`
	District             *string `json:"district"`
	State                *string `json:"state"`
	Rank                 *string `json:"rank"`
	TerritoryCode        *string `json:"territory_code"`
	CompetitionRank     *int    `json:"competition_rank"`
	TotalScore         *float64 `json:"total_score"`
	MSAchieved           float64 `json:"ms_achieved"`
	MSTarget             *float64 `json:"ms_target"`
	HSDAchieved           float64 `json:"hsd_achieved"`
	HSDTarget            *float64 `json:"hsd_target"`
	SpeedAchieved        float64 `json:"speed_achieved"`
	MSAchievementPct      float64 `json:"ms_achievement_pct"`
	HSDAchievementPct    float64 `json:"hsd_achievement_pct"`
	TotalAchievementPct  float64 `json:"total_achievement_pct"`
	YoYGrowthPct       *float64 `json:"yoy_growth_pct"`
	Trend              *string `json:"trend"`
}

// DistrictSummary aggregates outlets by district for territory/summary endpoint.
type DistrictSummary struct {
	District         string  `json:"district"`
	TotalOutlets     int     `json:"total_outlets"`
	AboveTarget     int     `json:"above_target"`
	BelowTarget     int     `json:"below_target"`
	AvgAchievementPct float64 `json:"avg_achievement_pct"`
}

type TerritorySummary struct {
	Period          string               `json:"period"`
	TerritoryCode   string               `json:"territory_code"`
	Summary        TerritoryStats        `json:"summary"`
	Districts      []DistrictSummary    `json:"districts,omitempty"`
	Outlets        []*TerritoryOutlet    `json:"outlets"`
}

type TerritoryStats struct {
	TotalOutlets     int     `json:"total_outlets"`
	AboveTarget     int     `json:"above_target"`
	BelowTarget     int     `json:"below_target"`
	AvgAchievementPct float64 `json:"avg_achievement_pct"`
}
