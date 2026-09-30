package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
)

func (r *CartRepository) AddCart(ctx context.Context) (*entity.Cart, error) {
	var cart entity.Cart
	q := `
		INSERT INTO carts DEFAULT VALUES
		RETURNING id;
	`
	if err := r.db.GetContext(ctx, &cart.ID, q); err != nil {
		return nil, fmt.Errorf("r.AddCart: %w", err)
	}
	cart.Items = make([]entity.CartItem, 0)
	return &cart, nil
}

func (r *CartRepository) GetCart(ctx context.Context, cartID uuid.UUID) (*entity.Cart, error) {
	var cart entity.Cart
	var items []entity.CartItem

	q := `SELECT c.id AS cart_id,
       i.id AS id,
       i.product AS product,
       i.price AS price
		FROM carts c
		LEFT JOIN cart_items i ON c.id = i.cart_id
		WHERE c.id = $1`
	err := r.db.SelectContext(ctx, &items, q, cartID)
	if err != nil {
		return nil, fmt.Errorf("r.GetCart: %w", err)
	}
	if len(items) == 0 {
		return nil, errs.ErrCartNotFound
	}

	cart.ID = items[0].CartID

	for _, item := range items {
		if item.ID == nil {
			continue
		}
		if item.Product == nil {
			continue
		}
		if item.Price == nil {
			continue
		}
		cart.Items = append(cart.Items, item)
	}

	return &cart, nil
}

func (r *CartRepository) AddCartItem(ctx context.Context, cartID uuid.UUID, product string, price float64, itemLimit int) (*entity.CartItem, error) {

	var pgErr *pgconn.PgError

	var count int
	var cartItem entity.CartItem
	q := `SELECT COUNT(*) FROM cart_items WHERE cart_id=$1;`
	if err := r.db.GetContext(ctx, &count, q, cartID); err != nil {
		return nil, fmt.Errorf("r.AddCartItem: %w", err)
	}
	if count >= itemLimit {
		return nil, errs.ErrFullCart
	}

	q = `INSERT INTO cart_items (cart_id, product, price)
VALUES ($1, $2, $3)
RETURNING id,cart_id, product, price`

	if err := r.db.GetContext(ctx, &cartItem, q, cartID, product, price); err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, errs.ErrCartNotFound
		}
		return nil, fmt.Errorf("r.AddCartItem: %w", err)
	}

	return &cartItem, nil
}

func (r *CartRepository) UpdateCartItem(ctx context.Context, cartID, itemID uuid.UUID, newProduct string, newPrice float64) (*entity.CartItem, error) {
	var cartItem entity.CartItem

	q := `SELECT
    c.id AS cart_id,
    i.id AS id,
    i.product,
    i.price
FROM carts c
         LEFT JOIN cart_items i ON i.cart_id = c.id AND i.id =$1
WHERE c.id = $2;`

	err := r.db.GetContext(ctx, &cartItem, q, itemID, cartID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrCartNotFound
		}
		return nil, fmt.Errorf("r.UpdateCartItem: %w", err)
	}
	if cartItem.ID == nil || cartItem.Product == nil || cartItem.Price == nil {
		return nil, errs.ErrCartItemNotFound
	}

	q = `UPDATE cart_items SET product=$1, price=$2 WHERE id=$3
RETURNING id,cart_id,product,price`

	err = r.db.GetContext(ctx, &cartItem, q, newProduct, newPrice, cartItem.ID)
	if err != nil {
		return nil, fmt.Errorf("r.UpdateCartItem: %w", err)
	}
	return &cartItem, nil
}

func (r *CartRepository) RemoveCartItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	var pgErr *pgconn.PgError

	q := `DELETE FROM cart_items WHERE id=$1 AND cart_id=$2`
	res, err := r.db.ExecContext(ctx, q, itemID, cartID)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return errs.ErrCartNotFound
		}
		return fmt.Errorf("r.RemoveCartItem: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("r.RemoveCartItem: %w", err)
	}
	if rowsAffected == 0 {
		return errs.ErrCartItemNotFound
	}

	return nil
}
