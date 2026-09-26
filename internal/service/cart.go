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

func (s *CartService) AddItem(ctx context.Context, cartId uuid.UUID, product string, price float64) (CartItemDTO, error) {
	cartItem, err := s.repo.AddCartItem(ctx, cartId, product, price)
	if err != nil {
		if errors.Is(err, errs.ErrFullCart) {
			return CartItemDTO{}, errs.ErrFullCart
		}
		if errors.Is(err, errs.ErrEmptyProduct) {
			return CartItemDTO{}, errs.ErrEmptyProduct
		}
		if errors.Is(err, errs.ErrNegativePrice) {
			return CartItemDTO{}, errs.ErrNegativePrice
		}
		return CartItemDTO{}, fmt.Errorf("s.AddItem: %w", err)
	}

	cartItemDTO := ItemToDTO(cartItem)

	fmt.Println(cartItemDTO)

	return cartItemDTO, nil
}

func (s *CartService) UpdateCartItem(ctx context.Context, ID uuid.UUID, newProduct string, newPrice float64) (CartItemDTO, error) {
	cartItem, err := s.repo.UpdateCartItem(ctx, ID, newProduct, newPrice)
	if err != nil {
		if errors.Is(err, errs.ErrCartItemNotFound) {
			return CartItemDTO{}, errs.ErrCartItemNotFound
		}
		if errors.Is(err, errs.ErrEmptyProduct) {
			return CartItemDTO{}, errs.ErrEmptyProduct
		}
		if errors.Is(err, errs.ErrNegativePrice) {
			return CartItemDTO{}, errs.ErrNegativePrice
		}
		return CartItemDTO{}, fmt.Errorf("s.UpdateItem: %w", err)
	}

	cartItemDTO := ItemToDTO(cartItem)

	return cartItemDTO, nil
}

//func (s *CartService) RemoveItem() {
//	//TODO implement me
//	panic("implement me")
//}

//func (s *CartService) CalculatePrice() {
//	//TODO implement me
//	panic("implement me")
//}
