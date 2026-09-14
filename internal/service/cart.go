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

func ViewCart() {}

func AddItem() {}

func UpdateItem() {}

func RemoveItem() {}

func CalculatePrice() {}
