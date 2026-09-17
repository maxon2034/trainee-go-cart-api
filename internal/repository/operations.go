package repository

import (
	"context"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

func (r *CartRepository) AddCart(ctx context.Context) (*entity.Cart, error) {
	panic("implement me")
	return nil, nil
}
func (r *CartRepository) GetCart(ctx context.Context, id string) (*entity.Cart, error) {
	panic("implement me")
	return nil, nil
}
func (r *CartRepository) AddCartItem(ctx context.Context, item *entity.CartItem) error {
	panic("implement me")
	return nil
}
func (r *CartRepository) UpdateCartItem(ctx context.Context, item *entity.CartItem) error {
	panic("implement me")
	return nil
}
func (r *CartRepository) RemoveCartItem(ctx context.Context, cartID, itemID string) error {
	panic("implement me")
	return nil
}
