package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

var ErrCartNotFound = errors.New("cart not found")

func (r *CartRepository) AddCart(ctx context.Context) (*entity.Cart, error) {
	var cart entity.Cart
	var cartDBO CartDBO
	query := `
		INSERT INTO carts DEFAULT VALUES
		RETURNING id;
	`
	if err := r.db.QueryRowContext(ctx, query).Scan(&cartDBO.ID); err != nil {
		return nil, fmt.Errorf("r.AddCart: %w", err)
	}
	cart.ID = cartDBO.ID
	cart.Items = make([]entity.CartItem, 0)
	return &cart, nil
}

func (r *CartRepository) GetCart(ctx context.Context, id uuid.UUID) (*entity.Cart, error) {
	var cart entity.Cart
	var cartDBO CartDBO
	q := `SELECT id FROM carts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(&cartDBO.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("r.GetCart: %w", err)
	}
	cart.ID = cartDBO.ID

	q = `SELECT id,product,price FROM cart_items WHERE cart_id=$1`

	if err := r.db.SelectContext(ctx, &cart.Items, q, id); err != nil {
		return nil, fmt.Errorf("r.GetCart: %w", err)
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
