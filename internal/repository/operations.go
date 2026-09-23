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
	err := r.db.QueryRowContext(ctx, q, id).Scan(&cartDBO.ID)
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
	var count int
	var cartItemDBO CartItemDBO
	var cartItem entity.CartItem
	q := `SELECT COUNT(*) FROM cart_items WHERE cart_id=$1`

	if err := r.db.GetContext(ctx, &count, q, cartID); err != nil {
		return nil, fmt.Errorf("r.AddCartItem1: %w", err)
	}
	if count == 5 {
		return nil, errs.ErrFullCart
	}

	q = `INSERT INTO cart_items (cart_id, product, price)
VALUES ($1, $2, $3)
RETURNING id,cart_id, product, price`
	if err := r.db.GetContext(ctx, &cartItemDBO, q, cartID, product, price); err != nil {
		return nil, fmt.Errorf("r.AddCartItem2: %w", err)
	}
	cartItem.CartID = cartItemDBO.CartID
	cartItem.ID = cartItemDBO.ID
	cartItem.Product = cartItemDBO.Product
	cartItem.Price = cartItemDBO.Price

	fmt.Println(cartItemDBO)
	return &cartItem, nil
}
func (r *CartRepository) UpdateCartItem(ctx context.Context, item *entity.CartItem) error {
	return errors.New("not implemented")
}
func (r *CartRepository) RemoveCartItem(ctx context.Context, cartID, itemID string) error {
	return errors.New("not implemented")
}
