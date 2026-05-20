package model

import (
	"time"

	"github.com/google/uuid"
)

type Target struct {
	ID          uuid.UUID  `json:"id"`
	CCNumber    string     `json:"cc_number"`
	ProductID   int16      `json:"product_id"`
	Period      time.Time  `json:"period"`
	TargetValue float64    `json:"target_value"`
	SetBy       *uuid.UUID `json:"set_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TargetsResponse groups targets by category for the API response.
type TargetsResponse struct {
	Period  string             `json:"period"`
	Fuel    map[string]float64 `json:"fuel"`    // product_code → target_value
	NonFuel map[string]float64 `json:"non_fuel"` // product_code → target_value
}
