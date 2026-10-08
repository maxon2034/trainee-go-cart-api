package entity

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEmptyProduct = errors.New("product cannot be empty")
	ErrInvalidPrice = errors.New("price must be greater than zero")
)

type CartItem struct {
	ID      uuid.UUID
	CartID  uuid.UUID
	Product string
	Price   float64
}

func NewCartItem(product string, price float64) (*CartItem, error) {
	if product == "" {
		return nil, ErrEmptyProduct
	}
	if price <= 0 {
		return nil, ErrInvalidPrice
	}

	return &CartItem{
		Product: product,
		Price:   price,
	}, nil
}
