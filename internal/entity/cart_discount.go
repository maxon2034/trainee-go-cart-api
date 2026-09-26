package entity

import "github.com/google/uuid"

type CartDiscount struct {
	CartID          uuid.UUID `json:"cart_id"`
	TotalPrice      float64   `json:"total_price"`
	DiscountPercent float64   `json:"discount_percent"`
	FinalPrice      float64   `json:"final_price"`
}
