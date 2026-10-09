package entity

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CartDiscount struct {
	CartID          uuid.UUID
	TotalPrice      decimal.Decimal
	DiscountPercent float64
	FinalPrice      decimal.Decimal
}
