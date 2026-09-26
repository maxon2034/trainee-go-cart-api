package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=../../mocks/mock_repository.go -package=mocks Repository

type Repository interface {
	AddCart(ctx context.Context) (*entity.Cart, error)
	// TODO: Rename Cart DTO
	GetCart(ctx context.Context, id uuid.UUID) (*entity.Cart, error)
	AddCartItem(ctx context.Context, cartId uuid.UUID, product string, price float64) (*entity.CartItem, error)
	UpdateCartItem(ctx context.Context, ID uuid.UUID, newProduct string, newPrice float64) (*entity.CartItem, error)
	//RemoveCartItem(ctx context.Context, cartID, itemID string) error
}
