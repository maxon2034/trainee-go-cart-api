package handlers

import (
	"context"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

type Service interface {
	CreateCart(ctx context.Context) (entity.CartDTO, error)
	ViewCart(ctx context.Context, id int) (entity.CartDTO, error)
	AddItem()
	UpdateItem()
	RemoveItem()
	CalculatePrice()
}
