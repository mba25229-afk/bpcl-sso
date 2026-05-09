package scoring

import "time"

type Actuals struct {
	CCCode           string
	MSVolKL          float64
	SpeedVolKL       float64
	HSDVolKL         float64
	MSLYKl           float64
	HSDLYKl          float64
	QOCCount         float64
	UFillCount       float64
	MAKGESalesKL     float64
	MAKGEHasData     bool
	GoogleComposite  float64
	GoogleRating     float64
	SangamCerts      float64
	BonusMarks       float64
	CleanlinessGrade string
}

type RankedActuals struct {
	Actuals
	Ranks map[string]int
}

type ScoringParam struct {
	MetricKey            string
	MaxMarks             float64
	NegativeScaleEnabled bool
	IsActive             bool
}

type ScoreRow struct {
	CCCode      string
	MonthYear   time.Time
	MetricKey   string
	ActualValue *float64
	RankInGroup *int
	MarksScored *float64
	MaxMarks    float64
}