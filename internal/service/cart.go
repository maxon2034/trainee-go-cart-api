package service

import (
	"context"
	"fmt"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

func (s *CartService) CreateCart(ctx context.Context) (entity.CartDTO, error) {
	cart, err := s.repo.AddCart(ctx)
	if err != nil {
		return entity.CartDTO{}, fmt.Errorf("error in adding cart: %w", err)
	}
	return entity.CartDTO{ID: cart.ID, Items: make([]entity.CartItemDTO, 0)}, nil
}

func (s *CartService) ViewCart(ctx context.Context, id int) (entity.CartDTO, error) {
	cart, err := s.repo.GetCart(ctx, id)
	if err != nil {
		return entity.CartDTO{}, fmt.Errorf("error in adding cart: %w", err)
	}
	return entity.CartDTO{ID: cart.ID, Items: cart.Items}, nil
}

func (s *CartService) AddItem() {
	//TODO implement me
	panic("implement me")
}

func (s *CartService) UpdateItem() {
	//TODO implement me
	panic("implement me")
}

func (s *CartService) RemoveItem() {
	//TODO implement me
	panic("implement me")
}

func (s *CartService) CalculatePrice() {
	//TODO implement me
	panic("implement me")
}
