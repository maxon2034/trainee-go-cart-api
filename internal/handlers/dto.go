package handlers

import (
	"github.com/google/uuid"
)

type CreateCartResponse struct {
	ID    uuid.UUID                `json:"id"`
	Items []CreateCartItemResponse `json:"items"`
}

type CreateCartItemResponse struct {
	ID      uuid.UUID `json:"id"`
	CartID  uuid.UUID `json:"cart_id"`
	Product string    `json:"product"`
	Price   float64   `json:"price"`
}

type ViewCartResponse struct {
	ID    uuid.UUID              `json:"id"`
	Items []ViewCartItemResponse `json:"items"`
}

type ViewCartItemResponse struct {
	ID      uuid.UUID `json:"id"`
	CartID  uuid.UUID `json:"cart_id"`
	Product string    `json:"product"`
	Price   float64   `json:"price"`
}

type AddItemRequest struct {
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}

type AddItemResponse struct {
	ID      uuid.UUID `json:"id"`
	CartID  uuid.UUID `json:"cart_id"`
	Product string    `json:"product"`
	Price   float64   `json:"price"`
}

type UpdateItemRequest struct {
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}

type UpdateItemResponse struct {
	ID      uuid.UUID `json:"id"`
	CartID  uuid.UUID `json:"cart"`
	Product string    `json:"product"`
	Price   float64   `json:"price"`
}

type CalculateDiscountResponse struct {
	CartID          uuid.UUID `json:"cart_id"`
	TotalPrice      float64   `json:"total_price"`
	DiscountPercent float64   `json:"discount_percent"`
	FinalPrice      float64   `json:"final_price"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
