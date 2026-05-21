package portal

type ScoreCard struct {
	Dealer       DealerInfo    `json:"dealer"`
	MonthYear    string        `json:"month_year"`
	AsOf         string        `json:"as_of"`
	TotalMarks   float64       `json:"total_marks"`
	MaxPossible  float64       `json:"max_possible"`
	RankOverall  int           `json:"rank_overall"`
	TotalDealers int           `json:"total_dealers"`
	Metrics      []MetricCard  `json:"metrics"`
	DailyTrend   []DailyPoint  `json:"daily_trend"`
	MAKGELog     []MAKGEEntry  `json:"mak_ge_log"`
	GoogleLog    []GoogleEntry `json:"google_log"`
}

type DealerInfo struct {
	CCCode string `json:"cc_code"`
	ROName string `json:"ro_name"`
	Area   string `json:"area"`
}

type MetricCard struct {
	MetricKey   string   `json:"metric_key"`
	DisplayName string   `json:"display_name"`
	ActualValue *float64 `json:"actual_value"`
	Target      *float64 `json:"target"`
	AchievePct  *float64 `json:"achieve_pct"`
	RankInGroup *int     `json:"rank_in_group"`
	MarksScored *float64 `json:"marks_scored"`
	MaxMarks    float64  `json:"max_marks"`
	Unit        string   `json:"unit"`
	Trend       string   `json:"trend"`
}

type DailyPoint struct {
	Date  string  `json:"date"`
	MSKL  float64 `json:"ms_kl"`
	HSDKL float64 `json:"hsd_kl"`
	QOC   int     `json:"qoc"`
	UFill int     `json:"ufill"`
}

type MAKGEEntry struct {
	ReadingDate  string  `json:"reading_date"`
	MeterReading float64 `json:"meter_reading"`
	Delta        float64 `json:"delta"`
}

type GoogleEntry struct {
	SnapshotDate string  `json:"snapshot_date"`
	Rating       float64 `json:"rating"`
	ReviewCount  int     `json:"review_count"`
}
