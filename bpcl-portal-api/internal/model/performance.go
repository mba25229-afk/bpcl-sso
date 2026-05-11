package model

import (
	"time"

	"github.com/google/uuid"
)

type PerformanceRecord struct {
	ID             uuid.UUID  `json:"id"`
	CCNumber       string     `json:"cc_number"`
	ProductID      int16      `json:"product_id"`
	Period         time.Time  `json:"period"`
	Achieved       *float64   `json:"achieved"`
	LastYear       *float64   `json:"last_year"`
	VolumeKL       *float64   `json:"volume_kl"`
	Source         string     `json:"source"`
	UploadedFileID *uuid.UUID `json:"uploaded_file_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PerformanceSummary is the per-product KPI row returned to the dashboard.
type PerformanceSummary struct {
	ProductCode  string   `json:"product_code"`
	ProductName  string   `json:"product_name"`
	Category     string   `json:"category"`
	Unit         string   `json:"unit"`
	Achieved     *float64 `json:"achieved"`
	Target       *float64 `json:"target"`
	LastYear     *float64 `json:"last_year"`
	VolumeKL     *float64 `json:"volume_kl"`
	GrowthPct    *float64 `json:"growth_pct"`
	TargetHitPct *float64 `json:"target_hit_pct"`
}

// KPIData is the full dashboard payload for a single outlet in a given period.
type KPIData struct {
	CCNumber  string               `json:"cc_number"`
	OutletName string              `json:"outlet_name"`
	Period    string               `json:"period"`
	Fuel      []PerformanceSummary `json:"fuel"`
	NonFuel   []PerformanceSummary `json:"non_fuel"`
}

// TrendRow is a single period's performance for trend chart.
type TrendRow struct {
	Period              string   `json:"period"`
	TotalAchieved      *float64 `json:"total_achieved"`
	TotalTarget        *float64 `json:"total_target"`
	FuelAchieved      *float64 `json:"fuel_achieved"`
	NonFuelAchieved   *float64 `json:"non_fuel_achieved"`
	AchievementPct    *float64 `json:"achievement_pct"`
}

// TrendResponse is the trend chart data for an outlet.
type TrendResponse struct {
	CCNumber string      `json:"cc_number"`
	Name    string      `json:"name"`
	Periods []TrendRow  `json:"periods"`
}

// DailySums holds monthly totals from crystal daily tables (cr_daily_ms, cr_daily_speed, cr_daily_ufill).
type DailySums struct {
	MSKL     float64
	SpeedKL  float64
	UfillCnt int64
}
