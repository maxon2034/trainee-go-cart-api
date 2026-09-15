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

func (r *CartRepository) GetCart(ctx context.Context, id int) (*entity.CartResponse, error) {
	var cart entity.CreateCartResponse
	cart.ID = id

	query := `SELECT * FROM cart_items WHERE id = ?;`

	rows, err := r.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error in getting cart: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cartItem entity.CartItemResponse
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
