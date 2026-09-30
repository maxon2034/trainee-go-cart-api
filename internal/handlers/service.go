package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

//go:generate mockgen -source=service.go -destination=../../mocks/mock_service.go -package=mocks Service

type Service interface {
	CreateCart(ctx context.Context) (*entity.Cart, error)
	ViewCart(ctx context.Context, cartID uuid.UUID) (*entity.Cart, error)
	AddItem(ctx context.Context, cartID uuid.UUID, product string, price float64) (*entity.CartItem, error)
	UpdateCartItem(ctx context.Context, cartID, itemID uuid.UUID, newProduct string, newPrice float64) (*entity.CartItem, error)
	RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error
	CalculateDiscount(ctx context.Context, cartID uuid.UUID) (uuid.UUID, float64, float64, float64, error)
}
