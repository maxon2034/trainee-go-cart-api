package entity

import "errors"

var (
	ErrEmptyProduct = errors.New("product cannot be empty")
	ErrInvalidPrice = errors.New("price must be greater than zero")
)

type CartItem struct {
	ID      int     `db:"id"`
	CartID  int     `db:"cart_id"`
	Product string  `db:"product"`
	Price   float64 `db:"price"`
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
