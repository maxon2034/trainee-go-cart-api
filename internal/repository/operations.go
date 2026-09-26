package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
)

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
	var cartItemsDBO []CartItemDBO

	q := `SELECT id FROM carts WHERE id = $1`
	err := r.db.GetContext(ctx, &cartDBO.ID, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrCartNotFound
		}
		return nil, fmt.Errorf("r.GetCart: %w", err)
	}
	cart.ID = cartDBO.ID

	q = `SELECT id,product,price FROM cart_items WHERE cart_id=$1`
	if err := r.db.SelectContext(ctx, &cartItemsDBO, q, id); err != nil {
		return nil, fmt.Errorf("r.GetCart: %w", err)
	}
	for _, item := range cartItemsDBO {
		cart.Items = append(cart.Items, entity.CartItem{
			ID:      item.ID,
			CartID:  item.ID,
			Product: item.Product,
			Price:   item.Price,
		})
	}
	return &cart, nil
}

func (r *CartRepository) AddCartItem(ctx context.Context, cartID uuid.UUID, product string, price float64) (*entity.CartItem, error) {
	if product == "" {
		return nil, errs.ErrEmptyProduct
	}
	if price < 0 {
		return nil, errs.ErrNegativePrice
	}

	var exists bool
	q := `SELECT EXISTS(SELECT 1 FROM carts WHERE id = $1)`
	err := r.db.GetContext(ctx, &exists, q, cartID)
	if err != nil {
		return nil, fmt.Errorf("r.RemoveCartItem: %w", err)
	}
	if !exists {
		return nil, errs.ErrCartNotFound
	}

	var count int
	var cartItemDBO CartItemDBO
	var cartItem entity.CartItem
	q = `SELECT COUNT(*) FROM cart_items WHERE cart_id=$1`
	if err := r.db.GetContext(ctx, &count, q, cartID); err != nil {
		return nil, fmt.Errorf("r.AddCartItem: %w", err)
	}
	if count == 5 {
		return nil, errs.ErrFullCart
	}

	q = `INSERT INTO cart_items (cart_id, product, price)
VALUES ($1, $2, $3)
RETURNING id,cart_id, product, price`
	if err := r.db.GetContext(ctx, &cartItemDBO, q, cartID, product, price); err != nil {
		return nil, fmt.Errorf("r.AddCartItem: %w", err)
	}

	cartItem.CartID = cartItemDBO.CartID
	cartItem.ID = cartItemDBO.ID
	cartItem.Product = cartItemDBO.Product
	cartItem.Price = cartItemDBO.Price

	return &cartItem, nil
}

func (r *CartRepository) UpdateCartItem(ctx context.Context, cartID, itemID uuid.UUID, newProduct string, newPrice float64) (*entity.CartItem, error) {
	var cartItemDBO CartItemDBO
	var cartItem entity.CartItem

	var exists bool
	q := `SELECT EXISTS(SELECT 1 FROM carts WHERE id = $1)`
	err := r.db.GetContext(ctx, &exists, q, cartID)
	if err != nil {
		return nil, fmt.Errorf("r.RemoveCartItem: %w", err)
	}
	if !exists {
		return nil, errs.ErrCartNotFound
	}

	q = `SELECT cart_id,id,product,price FROM cart_items WHERE id=$1 AND cart_id=$2`
	err = r.db.GetContext(ctx, &cartItemDBO, q, itemID, cartID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrCartItemNotFound
		}
		return nil, fmt.Errorf("r.GetCart: %w", err)
	}

	if cartItemDBO.Price < 0 {
		return nil, errs.ErrNegativePrice
	}
	if cartItemDBO.Product == "" {
		return nil, errs.ErrEmptyProduct
	}

	q = `UPDATE cart_items SET product=$1, price=$2 WHERE id=$3
RETURNING id,cart_id,product,price`

	err = r.db.GetContext(ctx, &cartItemDBO, q, newProduct, newPrice, cartItemDBO.ID)
	if err != nil {
		return nil, fmt.Errorf("r.UpdateCartItem2: %w", err)
	}

	cartItem.CartID = cartItemDBO.CartID
	cartItem.ID = cartItemDBO.ID
	cartItem.Product = cartItemDBO.Product
	cartItem.Price = cartItemDBO.Price

	return &cartItem, nil
}

func (r *CartRepository) RemoveCartItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	var exists bool
	q := `SELECT EXISTS(SELECT 1 FROM carts WHERE id = $1)`
	err := r.db.GetContext(ctx, &exists, q, cartID)
	if err != nil {
		return fmt.Errorf("r.RemoveCartItem: %w", err)
	}
	if !exists {
		return errs.ErrCartNotFound
	}

	q = `DELETE FROM cart_items WHERE id=$1 AND cart_id=$2`
	res, err := r.db.ExecContext(ctx, q, itemID, cartID)
	if err != nil {
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
