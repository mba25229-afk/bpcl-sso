package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bpcl/portal-api/internal/model"
	"github.com/google/uuid"
)

// product catalog — mirrors migration 000008_seed_products
var productCatalog = map[int16]model.Product{
	1: {ID: 1, Code: "MS", Name: "Motor Spirit (Petrol)", Category: model.CategoryFuel, Unit: "kL", DisplayOrder: 1},
	2: {ID: 2, Code: "HSD", Name: "High Speed Diesel", Category: model.CategoryFuel, Unit: "kL", DisplayOrder: 2},
	3: {ID: 3, Code: "SPEED", Name: "Speed (Premium Petrol)", Category: model.CategoryFuel, Unit: "kL", DisplayOrder: 3},
	4: {ID: 4, Code: "QOC", Name: "Quick Oil Change", Category: model.CategoryNonFuel, Unit: "Nos", DisplayOrder: 4},
	5: {ID: 5, Code: "Lubricants", Name: "MAK/GE Lubricants", Category: model.CategoryNonFuel, Unit: "₹", DisplayOrder: 5},
	6: {ID: 6, Code: "UFill", Name: "UFill Transactions", Category: model.CategoryNonFuel, Unit: "Nos", DisplayOrder: 6},
	7: {ID: 7, Code: "SBI", Name: "SBI Card Transactions", Category: model.CategoryNonFuel, Unit: "Nos", DisplayOrder: 7},
	8: {ID: 8, Code: "BeCafe", Name: "Be Café Revenue", Category: model.CategoryNonFuel, Unit: "₹", DisplayOrder: 8},
}

type PerformanceResponse struct {
	CCNumber              string                    `json:"cc_number"`
	OutletName            string                    `json:"outlet_name"`
	Period                string                    `json:"period"`
	Fuel                  []model.PerformanceSummary `json:"fuel"`
	NonFuel               []model.PerformanceSummary `json:"non_fuel"`
	TotalRevenueCr        float64                   `json:"total_revenue_cr"`
	TargetAchievementPct  *float64                  `json:"target_achievement_pct"`
	YoYGrowthPct          *float64                  `json:"yoy_growth_pct"`
	FuelAchieved          float64                   `json:"fuel_achieved"`
	NonFuelAchieved       float64                   `json:"non_fuel_achieved"`
}

type MonthlyDataPoint struct {
	Period string   `json:"period"`
	Value  float64  `json:"value"`
}

type ProductMixItem struct {
	ProductCode string  `json:"product_code"`
	ProductName string  `json:"product_name"`
	Value       float64 `json:"value"`
	Pct         float64 `json:"pct"`
}

type AnalysisResponse struct {
	CCNumber         string             `json:"cc_number"`
	OutletName       string             `json:"outlet_name"`
	From             string             `json:"from"`
	To               string             `json:"to"`
	FuelMix          []ProductMixItem   `json:"fuel_mix"`
	NonFuelMix       []ProductMixItem   `json:"non_fuel_mix"`
	TargetVsAchieved []struct {
		Period   string  `json:"period"`
		Target   float64 `json:"target"`
		Achieved float64 `json:"achieved"`
	} `json:"target_vs_achieved"`
	MonthlyGrowth  []MonthlyDataPoint `json:"monthly_growth"`
	CategoryGrowth []struct {
		Category string  `json:"category"`
		GrowthPct float64 `json:"growth_pct"`
	} `json:"category_growth"`
	WeeklyTrend []struct {
		Week    string  `json:"week"`
		Volume  float64 `json:"volume"`
	} `json:"weekly_trend"`
}

type PerformanceService struct {
	outlets     OutletRepository
	performance PerformanceRepository
	targets     TargetRepository
}

func NewPerformanceService(outlets OutletRepository, perf PerformanceRepository, targets TargetRepository) *PerformanceService {
	return &PerformanceService{outlets: outlets, performance: perf, targets: targets}
}

func (s *PerformanceService) GetPerformance(ctx context.Context, cc string, period time.Time, userID uuid.UUID) (*PerformanceResponse, error) {
	outlet, err := s.outlets.GetByCC(ctx, cc)
	if err != nil {
		return nil, err
	}
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	if !canAccess(claims, outlet) {
		return nil, model.ErrForbidden
	}

	recs, err := s.performance.GetByPeriod(ctx, cc, period)
	if err != nil {
		return nil, err
	}
	tgts, err := s.targets.GetByPeriod(ctx, cc, period)
	if err != nil {
		return nil, err
	}

	targetByProduct := make(map[int16]float64, len(tgts))
	for _, t := range tgts {
		targetByProduct[t.ProductID] = t.TargetValue
	}

	recByProduct := make(map[int16]*model.PerformanceRecord, len(recs))
	for _, r := range recs {
		recByProduct[r.ProductID] = r
	}

	var fuel, nonFuel []model.PerformanceSummary
	var totalAchieved, totalTarget, totalLastYear float64
	var fuelAchieved, nonFuelAchieved float64
	hasTarget := false
	hasLastYear := false

	for id := int16(1); id <= 8; id++ {
		prod, ok := productCatalog[id]
		if !ok {
			continue
		}
		rec := recByProduct[id]
		tgt := targetByProduct[id]
		if tgt > 0 {
			hasTarget = true
		}

		row := model.PerformanceSummary{
			ProductCode: prod.Code,
			ProductName: prod.Name,
			Category:    string(prod.Category),
			Unit:        prod.Unit,
		}
		if rec != nil {
			row.Achieved = rec.Achieved
			row.LastYear = rec.LastYear
			row.VolumeKL = rec.VolumeKL
		}
		if tgt > 0 {
			row.Target = &tgt
		}

		if row.Achieved != nil && row.Target != nil && *row.Target > 0 {
			pct := (*row.Achieved / *row.Target) * 100
			row.TargetHitPct = &pct
		}
		if row.Achieved != nil && row.LastYear != nil && *row.LastYear != 0 {
			pct := ((*row.Achieved - *row.LastYear) / *row.LastYear) * 100
			row.GrowthPct = &pct
		}

		achieved := 0.0
		if rec != nil && rec.Achieved != nil {
			achieved = *rec.Achieved
		}
		ly := 0.0
		if rec != nil && rec.LastYear != nil {
			ly = *rec.LastYear
			hasLastYear = true
		}

		totalAchieved += achieved
		totalTarget += tgt
		totalLastYear += ly

		if prod.Category == model.CategoryFuel {
			fuelAchieved += achieved
			fuel = append(fuel, row)
		} else {
			nonFuelAchieved += achieved
			nonFuel = append(nonFuel, row)
		}
	}

	resp := &PerformanceResponse{
		CCNumber:        cc,
		OutletName:      outlet.Name,
		Period:          period.Format("2006-01"),
		Fuel:            fuel,
		NonFuel:         nonFuel,
		TotalRevenueCr:  totalAchieved / 10_000_000,
		FuelAchieved:    fuelAchieved,
		NonFuelAchieved: nonFuelAchieved,
	}
	if hasTarget && totalTarget > 0 {
		pct := (totalAchieved / totalTarget) * 100
		resp.TargetAchievementPct = &pct
	}
	if hasLastYear && totalLastYear != 0 {
		pct := ((totalAchieved - totalLastYear) / totalLastYear) * 100
		resp.YoYGrowthPct = &pct
	}
	return resp, nil
}

func (s *PerformanceService) GetAnalysis(ctx context.Context, cc string, from, to time.Time, userID uuid.UUID) (*AnalysisResponse, error) {
	outlet, err := s.outlets.GetByCC(ctx, cc)
	if err != nil {
		return nil, err
	}
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	if !canAccess(claims, outlet) {
		return nil, model.ErrForbidden
	}

	recs, err := s.performance.GetDateRange(ctx, cc, from, to)
	if err != nil {
		return nil, err
	}

	// Group by period
	type periodKey = string
	byPeriod := make(map[periodKey]map[int16]*model.PerformanceRecord)
	for _, r := range recs {
		pk := r.Period.Format("2006-01")
		if byPeriod[pk] == nil {
			byPeriod[pk] = make(map[int16]*model.PerformanceRecord)
		}
		byPeriod[pk][r.ProductID] = r
	}

	// Fuel/non-fuel mix across whole range
	fuelTotals := make(map[int16]float64)
	nonFuelTotals := make(map[int16]float64)
	var totalFuel, totalNonFuel float64
	for _, pr := range byPeriod {
		for id, r := range pr {
			if r.Achieved == nil {
				continue
			}
			prod := productCatalog[id]
			if prod.Category == model.CategoryFuel {
				fuelTotals[id] += *r.Achieved
				totalFuel += *r.Achieved
			} else {
				nonFuelTotals[id] += *r.Achieved
				totalNonFuel += *r.Achieved
			}
		}
	}

	buildMix := func(totals map[int16]float64, grand float64) []ProductMixItem {
		var items []ProductMixItem
		for id := int16(1); id <= 8; id++ {
			v, ok := totals[id]
			if !ok {
				continue
			}
			prod := productCatalog[id]
			pct := 0.0
			if grand > 0 {
				pct = (v / grand) * 100
			}
			items = append(items, ProductMixItem{
				ProductCode: prod.Code,
				ProductName: prod.Name,
				Value:       v,
				Pct:         pct,
			})
		}
		return items
	}

	// Targets for range
	tgts, err := s.targets.GetByPeriod(ctx, cc, from)
	if err != nil {
		tgts = nil
	}
	targetByProduct := make(map[int16]float64)
	for _, t := range tgts {
		targetByProduct[t.ProductID] = t.TargetValue
	}

	// Monthly growth (total achieved per period)
	var monthlyGrowth []MonthlyDataPoint
	var targetVsAchieved []struct {
		Period   string  `json:"period"`
		Target   float64 `json:"target"`
		Achieved float64 `json:"achieved"`
	}

	// Build sorted period list
	periods := make([]string, 0, len(byPeriod))
	for pk := range byPeriod {
		periods = append(periods, pk)
	}
	// Sort periods
	for i := 0; i < len(periods)-1; i++ {
		for j := i + 1; j < len(periods); j++ {
			if periods[i] > periods[j] {
				periods[i], periods[j] = periods[j], periods[i]
			}
		}
	}

	for _, pk := range periods {
		pr := byPeriod[pk]
		var achieved, tgt float64
		for id, r := range pr {
			if r.Achieved != nil {
				achieved += *r.Achieved
			}
			tgt += targetByProduct[id]
		}
		monthlyGrowth = append(monthlyGrowth, MonthlyDataPoint{Period: pk, Value: achieved})
		targetVsAchieved = append(targetVsAchieved, struct {
			Period   string  `json:"period"`
			Target   float64 `json:"target"`
			Achieved float64 `json:"achieved"`
		}{Period: pk, Target: tgt, Achieved: achieved})
	}

	// Weekly trend — divide each month by 4
	var weeklyTrend []struct {
		Week   string  `json:"week"`
		Volume float64 `json:"volume"`
	}
	for _, dp := range monthlyGrowth {
		weekVol := dp.Value / 4
		for w := 1; w <= 4; w++ {
			weeklyTrend = append(weeklyTrend, struct {
				Week   string  `json:"week"`
				Volume float64 `json:"volume"`
			}{Week: fmt.Sprintf("%s-W%d", dp.Period, w), Volume: weekVol})
		}
	}

	// Category growth
	var fuelLY, nonFuelLY float64
	for _, r := range recs {
		if r.LastYear == nil {
			continue
		}
		prod := productCatalog[r.ProductID]
		if prod.Category == model.CategoryFuel {
			fuelLY += *r.LastYear
		} else {
			nonFuelLY += *r.LastYear
		}
	}
	var categoryGrowth []struct {
		Category  string  `json:"category"`
		GrowthPct float64 `json:"growth_pct"`
	}
	if fuelLY != 0 {
		categoryGrowth = append(categoryGrowth, struct {
			Category  string  `json:"category"`
			GrowthPct float64 `json:"growth_pct"`
		}{"fuel", ((totalFuel - fuelLY) / fuelLY) * 100})
	}
	if nonFuelLY != 0 {
		categoryGrowth = append(categoryGrowth, struct {
			Category  string  `json:"category"`
			GrowthPct float64 `json:"growth_pct"`
		}{"non_fuel", ((totalNonFuel - nonFuelLY) / nonFuelLY) * 100})
	}

	return &AnalysisResponse{
		CCNumber:         cc,
		OutletName:       outlet.Name,
		From:             from.Format("2006-01"),
		To:               to.Format("2006-01"),
		FuelMix:          buildMix(fuelTotals, totalFuel),
		NonFuelMix:       buildMix(nonFuelTotals, totalNonFuel),
		TargetVsAchieved: targetVsAchieved,
		MonthlyGrowth:    monthlyGrowth,
		CategoryGrowth:   categoryGrowth,
		WeeklyTrend:      weeklyTrend,
	}, nil
}

func (s *PerformanceService) GetTrend(ctx context.Context, cc string, months int) (*model.TrendResponse, error) {
	outlet, err := s.outlets.GetByCC(ctx, cc)
	if err != nil {
		return nil, err
	}
	claims, err := ExtractClaims(ctx)
	if err != nil {
		return nil, err
	}
	if !canAccess(claims, outlet) {
		return nil, model.ErrForbidden
	}

	rows, err := s.performance.GetTrend(ctx, cc, months)
	if err != nil {
		return nil, err
	}

	periods := make([]model.TrendRow, len(rows))
	for i, r := range rows {
		periods[i] = *r
	}

	return &model.TrendResponse{
		CCNumber: cc,
		Name:    outlet.Name,
		Periods: periods,
	}, nil
}
