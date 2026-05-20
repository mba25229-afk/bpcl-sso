package model

type ProductCategory string

const (
	CategoryFuel    ProductCategory = "fuel"
	CategoryNonFuel ProductCategory = "non_fuel"
)

type Product struct {
	ID           int16           `json:"id"`
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	Category     ProductCategory `json:"category"`
	Unit         string          `json:"unit"`
	DisplayOrder int16           `json:"display_order"`
}
