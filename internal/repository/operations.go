package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

func (r *CartRepository) AddCart(ctx context.Context) (*entity.Cart, error) {
	var cart entity.Cart
	query := `
		INSERT INTO carts DEFAULT VALUES
		RETURNING id, created_at;
	`
	if err := r.DB.QueryRowContext(ctx, query).Scan(&cart.ID, &cart.CreatedAt); err != nil {
		return nil, fmt.Errorf("error in creating cart: %w", err)
	}
	return &cart, nil
}

func (r *CartRepository) GetCart(ctx context.Context, id int) (*entity.CartDTO, error) {
	var cart entity.CartDTO
	queryCart := `SELECT id FROM carts WHERE id = $1`
	err := r.DB.QueryRowContext(ctx, queryCart, id).Scan(&cart.ID)
	if err != nil {
		return nil, fmt.Errorf("error in getting cart: %w", err)
	}

	query := `SELECT * FROM cart_items WHERE cart_id=$1`

	rows, err := r.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error in getting cart: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cartItem entity.CartItemDTO
		if err := rows.Scan(&cartItem); err != nil {
			return nil, fmt.Errorf("error in getting cartItem: %w", err)
		}
		cart.Items = append(cart.Items, cartItem)
	}
	return &cart, nil
}
func (r *CartRepository) AddCartItem(ctx context.Context, item *entity.CartItem) error {
	return errors.New("not implemented")
}
func (r *CartRepository) UpdateCartItem(ctx context.Context, item *entity.CartItem) error {
	return errors.New("not implemented")
}
func (r *CartRepository) RemoveCartItem(ctx context.Context, cartID, itemID string) error {
	return errors.New("not implemented")
}
