package service

import (
	"context"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

type Repository interface {
	AddCart(ctx context.Context) (*entity.Cart, error)
	GetCart(ctx context.Context, id string) (*entity.Cart, error)
	AddCartItem(ctx context.Context, item *entity.CartItem) error
	UpdateCartItem(ctx context.Context, item *entity.CartItem) error
	RemoveCartItem(ctx context.Context, cartID, itemID string) error
}
