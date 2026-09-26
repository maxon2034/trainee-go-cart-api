package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
)

func (s *CartService) CreateCart(ctx context.Context) (*CartDTO, error) {
	cart, err := s.repo.AddCart(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.CreateCart: %w", err)
	}

	cartDTO := ToDTO(cart)
	return &cartDTO, nil
}

func (s *CartService) ViewCart(ctx context.Context, cartID uuid.UUID) (*CartDTO, error) {
	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			return nil, errs.ErrCartNotFound
		}
		return nil, fmt.Errorf("s.ViewCart: %w", err)
	}

	cartDTO := ToDTO(cart)

	return &cartDTO, nil
}

func (s *CartService) AddItem(ctx context.Context, cartID uuid.UUID, product string, price float64) (*CartItemDTO, error) {
	cartItem, err := s.repo.AddCartItem(ctx, cartID, product, price)
	if err != nil {
		if errors.Is(err, errs.ErrFullCart) {
			return nil, errs.ErrFullCart
		}
		if errors.Is(err, errs.ErrEmptyProduct) {
			return nil, errs.ErrEmptyProduct
		}
		if errors.Is(err, errs.ErrNegativePrice) {
			return nil, errs.ErrNegativePrice
		}
		return nil, fmt.Errorf("s.AddItem: %w", err)
	}

	cartItemDTO := ItemToDTO(cartItem)

	return &cartItemDTO, nil
}

func (s *CartService) UpdateCartItem(ctx context.Context, cartID, itemID uuid.UUID, newProduct string, newPrice float64) (*CartItemDTO, error) {
	cartItem, err := s.repo.UpdateCartItem(ctx, cartID, itemID, newProduct, newPrice)
	if err != nil {
		if errors.Is(err, errs.ErrCartItemNotFound) {
			return nil, errs.ErrCartItemNotFound
		}
		if errors.Is(err, errs.ErrCartNotFound) {
			return nil, errs.ErrCartNotFound
		}
		if errors.Is(err, errs.ErrEmptyProduct) {
			return nil, errs.ErrEmptyProduct
		}
		if errors.Is(err, errs.ErrNegativePrice) {
			return nil, errs.ErrNegativePrice
		}
		return nil, fmt.Errorf("s.UpdateItem: %w", err)
	}

	cartItemDTO := ItemToDTO(cartItem)

	return &cartItemDTO, nil
}

func (s *CartService) RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	err := s.repo.RemoveCartItem(ctx, cartID, itemID)
	if err != nil {
		if errors.Is(err, errs.ErrCartItemNotFound) {
			return errs.ErrCartItemNotFound
		}
		if errors.Is(err, errs.ErrCartNotFound) {
			return errs.ErrCartNotFound
		}
		return fmt.Errorf("s.RemoveItem: %w", err)
	}
	return nil
}

func (s *CartService) CalculatePrice(ctx context.Context, cartID uuid.UUID) (*entity.CartDiscount, error) {
	cartDiscount, err := s.repo.CalculateDiscount(ctx, cartID)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			return nil, errs.ErrCartNotFound
		}
		if errors.Is(err, errs.ErrEmptyCart) {
			return nil, errs.ErrEmptyCart
		}
		return nil, fmt.Errorf("s.CalculatePrice: %w", err)
	}
	return cartDiscount, nil
}
