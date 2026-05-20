package ingest

import "time"

type MetricKey string

const (
	MetricUfill MetricKey = "ufill"
	MetricQOC   MetricKey = "qoc"
	MetricMS    MetricKey = "ms"
	MetricHSD   MetricKey = "hsd"
	MetricSpeed MetricKey = "speed"
)

type DailyRow struct {
	CCCode string  `json:"cc_code"`
	Value  float64 `json:"value"`
}

type BulkDailyRequest struct {
	Date   string     `json:"date"`
	Metric MetricKey  `json:"metric"`
	Rows   []DailyRow `json:"rows"`
}

type MAKGERequest struct {
	CCCode       string  `json:"cc_code"`
	ReadingDate  string  `json:"reading_date"`
	MeterReading float64 `json:"meter_reading"`
}

type GoogleRatingRequest struct {
	CCCode       string  `json:"cc_code"`
	SnapshotDate string  `json:"snapshot_date"`
	Rating       float64 `json:"rating"`
	ReviewCount  int     `json:"review_count"`
}

type TargetRow struct {
	CCCode        string   `json:"cc_code"`
	UfillTarget   *int     `json:"ufill_target"`
	QOCTarget     *int     `json:"qoc_target"`
	SpeedKL       *float64 `json:"speed_kl"`
	MSKL          *float64 `json:"ms_kl"`
	HSDKL         *float64 `json:"hsd_kl"`
	MSLY          *float64 `json:"ms_ly"`
	HSDLY         *float64 `json:"hsd_ly"`
	DSWAvailable  *bool    `json:"dsw_available"`
	Nitrogen      *bool    `json:"nitrogen"`
	MAKGETarget   *float64 `json:"mak_ge_target"`
	DarpanTarget  *int     `json:"darpan_target"`
	CoolantLubeKL *float64 `json:"coolant_lube_kl"`
	Remarks       *string  `json:"remarks"`
}

type BulkTargetsRequest struct {
	MonthYear string      `json:"month_year"`
	Rows      []TargetRow `json:"rows"`
}

type IngestResult struct {
	Inserted int      `json:"inserted"`
	Updated  int      `json:"updated"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

type Period struct {
	MonthYear  time.Time
	TotalSlots int
	IsActive   bool
}
