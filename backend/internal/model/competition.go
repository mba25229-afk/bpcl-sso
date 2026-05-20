package model

import (
	"time"

	"github.com/google/uuid"
)

type CompetitionPeriod struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Period        time.Time  `json:"period"`
	TerritoryCode string     `json:"territory_code"`
	Status        string     `json:"status"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CompetitionScore struct {
	ID               uuid.UUID  `json:"id"`
	CompetitionID    uuid.UUID  `json:"competition_id"`
	CCNumber         string     `json:"cc_number"`
	MSVolKL          *float64   `json:"ms_vol_kl"`
	MSGrowthPct      *float64   `json:"ms_growth_pct"`
	MSTAGainPP       *float64   `json:"ms_ta_gain_pp"`
	HSDVolKL         *float64   `json:"hsd_vol_kl"`
	HSDGrowthPct     *float64   `json:"hsd_growth_pct"`
	HSDTAGainPP      *float64   `json:"hsd_ta_gain_pp"`
	QOCCount         *float64   `json:"qoc_count"`
	LubricantsValue  *float64   `json:"lubricants_value"`
	UFillCount       *float64   `json:"ufill_count"`
	SpeedVolKL       *float64   `json:"speed_vol_kl"`
	ScoreMSVol       *float64   `json:"score_ms_vol"`
	ScoreMSGrowth    *float64   `json:"score_ms_growth"`
	ScoreMSTAGain    *float64   `json:"score_ms_ta_gain"`
	ScoreHSDVol      *float64   `json:"score_hsd_vol"`
	ScoreHSDGrowth   *float64   `json:"score_hsd_growth"`
	ScoreHSDTAGain   *float64   `json:"score_hsd_ta_gain"`
	ScoreQOC         *float64   `json:"score_qoc"`
	ScoreLubricants  *float64   `json:"score_lubricants"`
	ScoreUFill       *float64   `json:"score_ufill"`
	ScoreSpeed       *float64   `json:"score_speed"`
	ScoreCleanliness *float64   `json:"score_cleanliness"`
	ScoreSangam      *float64   `json:"score_sangam"`
	ScoreGoogle      *float64   `json:"score_google"`
	ScoreIPS         *float64   `json:"score_ips"`
	ScoreBonus       *float64   `json:"score_bonus"`
	TotalScore       *float64   `json:"total_score"`
	Rank             *int       `json:"rank"`
	ComputedAt       *time.Time `json:"computed_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

type DealerAuditScore struct {
	ID                 uuid.UUID  `json:"id"`
	CCNumber           string     `json:"cc_number"`
	Period             time.Time  `json:"period"`
	CleanlinessGrade   *string    `json:"cleanliness_grade"`
	CleanlinessStatus  string     `json:"cleanliness_status"`
	SangamCount        *int       `json:"sangam_count"`
	GoogleRating       *float64   `json:"google_rating"`
	IPSPct             *float64   `json:"ips_pct"`
	AuditStatus        string     `json:"audit_status"`
	AuditedBy          *uuid.UUID `json:"audited_by"`
	AuditedAt          *time.Time `json:"audited_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

type CompetitionBonus struct {
	ID            uuid.UUID `json:"id"`
	CompetitionID uuid.UUID `json:"competition_id"`
	CCNumber      string    `json:"cc_number"`
	BonusMarks    float64   `json:"bonus_marks"`
	Remarks       string    `json:"remarks"`
	AwardedBy     uuid.UUID `json:"awarded_by"`
	CreatedAt     time.Time `json:"created_at"`
}
