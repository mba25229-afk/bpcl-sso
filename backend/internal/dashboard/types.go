package dashboard

import "time"

type DashboardRow struct {
	CCCode          string                `json:"cc_code"`
	ROName          string                `json:"ro_name"`
	Rank            int                   `json:"rank"`
	TotalMarks      float64               `json:"total_marks"`
	MaxPossible     float64               `json:"max_possible"`
	AchievePct      float64               `json:"achieve_pct"`
	Status          string                `json:"status"`
	MetricBreakdown []MetricBreakdownItem `json:"metric_breakdown"`
}

type MetricBreakdownItem struct {
	MetricKey   string   `json:"metric_key"`
	DisplayName string   `json:"display_name"`
	MarksScored *float64 `json:"marks_scored"`
	MaxMarks    float64  `json:"max_marks"`
	Rank        *int     `json:"rank"`
}

type DashboardResponse struct {
	MonthYear  string           `json:"month_year"`
	ComputedAt time.Time        `json:"computed_at"`
	Dealers    []DashboardRow   `json:"dealers"`
	Summary    DashboardSummary `json:"summary"`
}

type DashboardSummary struct {
	TotalDealers int     `json:"total_dealers"`
	GreenCount   int     `json:"green_count"`
	AmberCount   int     `json:"amber_count"`
	RedCount     int     `json:"red_count"`
	AvgScore     float64 `json:"avg_score"`
	TopScore     float64 `json:"top_score"`
}

// TrafficLight is exported so it can be tested directly.
func TrafficLight(pct float64) string {
	switch {
	case pct >= 90:
		return "green"
	case pct >= 70:
		return "amber"
	default:
		return "red"
	}
}

// BuildSummary is exported so it can be tested directly.
func BuildSummary(dealers []DashboardRow) DashboardSummary {
	s := DashboardSummary{TotalDealers: len(dealers)}
	var total float64
	for _, d := range dealers {
		switch d.Status {
		case "green":
			s.GreenCount++
		case "amber":
			s.AmberCount++
		case "red":
			s.RedCount++
		}
		total += d.TotalMarks
		if d.TotalMarks > s.TopScore {
			s.TopScore = d.TotalMarks
		}
	}
	if s.TotalDealers > 0 {
		s.AvgScore = total / float64(s.TotalDealers)
	}
	return s
}
