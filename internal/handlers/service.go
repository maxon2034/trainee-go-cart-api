package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/service"
)

type Service interface {
	CreateCart(ctx context.Context) (service.CartDTO, error)
	ViewCart(ctx context.Context, id uuid.UUID) (service.CartDTO, error)
	AddItem(ctx context.Context, cartId uuid.UUID, product string, price float64) (service.CartItemDTO, error)
	UpdateItem()
	RemoveItem()
	CalculatePrice()
}
