package handlers

import (
	"context"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

type Service interface {
	CreateCart(ctx context.Context) (entity.CreateCartResponse, error)
	ViewCart()
	AddItem()
	UpdateItem()
	RemoveItem()
	CalculatePrice()
}
