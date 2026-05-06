package model

import (
	"time"

	"github.com/google/uuid"
)

type MarketShareData struct {
	ID            uuid.UUID `json:"id"`
	OutletName    string    `json:"outlet_name"`
	CCNumber      *string   `json:"cc_number"`
	OMC           string    `json:"omc"`
	TradingAreaID int       `json:"trading_area_id"`
	Period        time.Time `json:"period"`
	MSVolKL       float64   `json:"ms_vol_kl"`
	HSDVolKL      float64   `json:"hsd_vol_kl"`
	Source        string    `json:"source"`
	UploadedFileID *uuid.UUID `json:"uploaded_file_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type TradingAreaTotal struct {
	ID            uuid.UUID `json:"id"`
	TradingAreaID int       `json:"trading_area_id"`
	Period        time.Time `json:"period"`
	TotalMSKL     float64   `json:"total_ms_kl"`
	TotalHSDKL    float64   `json:"total_hsd_kl"`
	BPCLMSKL      float64   `json:"bpcl_ms_kl"`
	HPCLMSKL      float64   `json:"hpcl_ms_kl"`
	IOCLMSKL      float64   `json:"iocl_ms_kl"`
	BPCLHSDKL     float64   `json:"bpcl_hsd_kl"`
	HPCLHSDKL     float64   `json:"hpcl_hsd_kl"`
	IOCLHSDKL     float64   `json:"iocl_hsd_kl"`
	ComputedAt    time.Time `json:"computed_at"`
}

type TradingArea struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	District  string `json:"district"`
	State     string `json:"state"`
}

type MarketShareStatusResponse struct {
	CompetitionID              uuid.UUID `json:"competition_id"`
	TotalDealers               int       `json:"total_dealers"`
	DealersWithMSData          int       `json:"dealers_with_ms_data"`
	DealersMissingFromSource   int       `json:"dealers_missing_from_source"`
	DealersWithPeriodGaps      int       `json:"dealers_with_period_gaps"`
	Message                    string    `json:"message"`
	Status                     string    `json:"status"`
}

type RecomputeResponse struct {
	CompetitionID    uuid.UUID  `json:"competition_id"`
	ScoresUpdated    int        `json:"scores_updated"`
	MSGainComputed   int        `json:"ms_gain_computed"`
	Top10Before      []TopScore `json:"top10_before"`
	Top10After       []TopScore `json:"top10_after"`
}

type TopScore struct {
	Rank       int     `json:"rank"`
	CCNumber   string  `json:"cc_number"`
	OutletName string  `json:"outlet_name"`
	TotalScore float64 `json:"total_score"`
}