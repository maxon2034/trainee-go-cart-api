package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
)

func (s *CartService) CreateCart(ctx context.Context) (CartDTO, error) {
	cart, err := s.repo.AddCart(ctx)
	if err != nil {
		return CartDTO{}, fmt.Errorf("s.CreateCart: %w", err)
	}

	cartDTO := ToDTO(cart)
	return cartDTO, nil
}

func (s *CartService) ViewCart(ctx context.Context, id uuid.UUID) (CartDTO, error) {
	cart, err := s.repo.GetCart(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			return CartDTO{}, errs.ErrCartNotFound
		}
		return CartDTO{}, fmt.Errorf("s.ViewCart: %w", err)
	}

	cartDTO := ToDTO(cart)

	return cartDTO, nil
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
