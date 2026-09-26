package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=../../mocks/mock_repository.go -package=mocks Repository

type Repository interface {
	AddCart(ctx context.Context) (*entity.Cart, error)
	GetCart(ctx context.Context, cartID uuid.UUID) (*entity.Cart, error)
	AddCartItem(ctx context.Context, cartID uuid.UUID, product string, price float64) (*entity.CartItem, error)
	UpdateCartItem(ctx context.Context, cartID, itemID uuid.UUID, newProduct string, newPrice float64) (*entity.CartItem, error)
	RemoveCartItem(ctx context.Context, cartID, itemID uuid.UUID) error
	CalculateDiscount(ctx context.Context, cartID uuid.UUID) (*entity.CartDiscount, error)
}
