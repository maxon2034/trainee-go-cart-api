package handlers

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateCartResponse struct {
	ID    uuid.UUID                `json:"id"`
	Items []CreateCartItemResponse `json:"items"`
}

type CreateCartItemResponse struct {
	ID      uuid.UUID       `json:"id"`
	CartID  uuid.UUID       `json:"cart_id"`
	Product string          `json:"product"`
	Price   decimal.Decimal `json:"price"`
}

type ViewCartResponse struct {
	ID    uuid.UUID              `json:"id"`
	Items []ViewCartItemResponse `json:"items"`
}

type ViewCartItemResponse struct {
	ID      uuid.UUID       `json:"id"`
	CartID  uuid.UUID       `json:"cart_id"`
	Product string          `json:"product"`
	Price   decimal.Decimal `json:"price"`
}

type AddItemRequest struct {
	Product string          `json:"product"`
	Price   decimal.Decimal `json:"price"`
}

type AddItemResponse struct {
	ID      uuid.UUID       `json:"id"`
	CartID  uuid.UUID       `json:"cart_id"`
	Product string          `json:"product"`
	Price   decimal.Decimal `json:"price"`
}

type UpdateItemRequest struct {
	Product string          `json:"product"`
	Price   decimal.Decimal `json:"price"`
}

type UpdateItemResponse struct {
	ID      uuid.UUID       `json:"id"`
	CartID  uuid.UUID       `json:"cart"`
	Product string          `json:"product"`
	Price   decimal.Decimal `json:"price"`
}

type CalculateDiscountResponse struct {
	CartID          uuid.UUID       `json:"cart_id"`
	TotalPrice      decimal.Decimal `json:"total_price"`
	DiscountPercent float64         `json:"discount_percent"`
	FinalPrice      decimal.Decimal `json:"final_price"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
