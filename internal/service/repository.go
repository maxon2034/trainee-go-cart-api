package service

import (
	"context"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

type Repository interface {
	AddCart(ctx context.Context) (*entity.Cart, error)
	// TODO: Rename Cart DTO
	GetCart(ctx context.Context, id int) (*entity.CreateCartResponse, error)
	AddCartItem(ctx context.Context, item *entity.CartItem) error
	UpdateCartItem(ctx context.Context, item *entity.CartItem) error
	RemoveCartItem(ctx context.Context, cartID, itemID string) error
}
