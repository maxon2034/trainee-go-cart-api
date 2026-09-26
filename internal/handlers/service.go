package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/service"
)

//go:generate mockgen -source=service.go -destination=../../mocks/mock_service.go -package=mocks Service

type Service interface {
	CreateCart(ctx context.Context) (service.CartDTO, error)
	ViewCart(ctx context.Context, id uuid.UUID) (service.CartDTO, error)
	AddItem(ctx context.Context, cartID uuid.UUID, product string, price float64) (service.CartItemDTO, error)
	UpdateCartItem(ctx context.Context, ID uuid.UUID, newProduct string, newPrice float64) (service.CartItemDTO, error)
	RemoveItem(ctx context.Context, itemID uuid.UUID) error
	//CalculatePrice()
}
