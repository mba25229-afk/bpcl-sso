package admin

import "time"

type Dealer struct {
	CCCode      string `json:"cc_code"`
	ROName      string `json:"ro_name"`
	Area        string `json:"area"`
	IsActive    bool   `json:"is_active"`
	DealerEmail string `json:"dealer_email,omitempty"`
}

type CompetitionPeriod struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	MonthYear  time.Time `json:"month_year"`
	TotalSlots int       `json:"total_slots"`
	IsActive   bool      `json:"is_active"`
}

type CreatePeriodRequest struct {
	Name      string `json:"name"`
	MonthYear string `json:"month_year"`
}

type ToggleDealerRequest struct {
	IsActive bool `json:"is_active"`
}

type TargetRow struct {
	CCCode        string   `json:"cc_code"`
	ROName        string   `json:"ro_name"`
	UfillTarget   *int     `json:"ufill_target"`
	QOCTarget     *int     `json:"qoc_target"`
	SpeedKL       *float64 `json:"speed_kl"`
	MSKL          *float64 `json:"ms_kl"`
	HSDKL         *float64 `json:"hsd_kl"`
	MSLY          *float64 `json:"ms_ly"`
	HSDLY         *float64 `json:"hsd_ly"`
	MAKGETarget   *float64 `json:"mak_ge_target"`
	DarpanTarget  *int     `json:"darpan_target"`
	CoolantLubeKL *float64 `json:"coolant_lube_kl"`
	DSWAvailable  *bool    `json:"dsw_available"`
	Nitrogen      *bool    `json:"nitrogen"`
}

type IngestSummaryRow struct {
	CCCode     string  `json:"cc_code"`
	ROName     string  `json:"ro_name"`
	MTDSum     float64 `json:"mtd_sum"`
	DaysFilled int     `json:"days_filled"`
}

type MAKGERow struct {
	CCCode       string  `json:"cc_code"`
	ROName       string  `json:"ro_name"`
	ReadingDate  string  `json:"reading_date"`
	MeterReading float64 `json:"meter_reading"`
}

type ManualScoreRow struct {
	CCCode           string   `json:"cc_code"`
	ROName           string   `json:"ro_name"`
	BonusMarks       *float64 `json:"bonus_marks"`
	CleanlinessGrade *string  `json:"cleanliness_grade"`
}

type BulkManualScoresRequest struct {
	MonthYear string           `json:"month_year"`
	Rows      []ManualScoreRow `json:"rows"`
}
