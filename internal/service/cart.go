package service

import (
	"context"
	"fmt"

	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

func (s *CartService) CreateCart(ctx context.Context) (entity.CreateCartResponse, error) {
	cart, err := s.repo.AddCart(ctx)
	if err != nil {
		return entity.CreateCartResponse{}, fmt.Errorf("error in adding cart: %w", err)
	}
	return entity.CreateCartResponse{ID: cart.ID, Items: make([]entity.CartItem, 0)}, nil
}

func (s *CartService) ViewCart() {
	//TODO implement me
	panic("implement me")
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
