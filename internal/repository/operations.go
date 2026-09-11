package repository

import (
	"context"
	"fmt"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

func (r *CartRepository) AddCart(ctx context.Context, cart *entity.Cart) (*entity.Cart, error) {
	query := `
		INSERT INTO carts DEFAULT VALUES
		RETURNING id, created_at;
	`
	if err := r.DB.QueryRowContext(ctx, query).Scan(&cart.ID, cart.CreatedAt); err != nil {
		return nil, fmt.Errorf("error in creating cart: %w", err)
	}
	return cart, nil
}

func (r *CartRepository) GetCart(ctx context.Context, id string) (*entity.Cart, error)    {}
func (r *CartRepository) AddCartItem(ctx context.Context, item *entity.CartItem) error    {}
func (r *CartRepository) UpdateCartItem(ctx context.Context, item *entity.CartItem) error {}
func (r *CartRepository) RemoveCartItem(ctx context.Context, cartID, itemID string) error {}
