package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
)

const ItemLimit = 5

func (s *CartService) CreateCart(ctx context.Context) (*entity.Cart, error) {
	cart, err := s.repo.AddCart(ctx)
	if err != nil {
		s.logger.Error("create cart", slog.Any("error", err))
		return nil, fmt.Errorf("s.CreateCart: %w", err)
	}

	return cart, nil
}

func (s *CartService) ViewCart(ctx context.Context, cartID uuid.UUID) (*entity.Cart, error) {
	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		s.logger.ErrorContext(ctx, "view cart", slog.Any("cart id", cartID), slog.Any("error", err))
		return nil, fmt.Errorf("s.ViewCart: %w", err)
	}

	return cart, nil
}

func (s *CartService) AddItem(ctx context.Context, cartID uuid.UUID, product string, price float64) (*entity.CartItem, error) {
	cartItem, err := s.repo.AddCartItem(ctx, cartID, product, price, ItemLimit)
	if err != nil {
		s.logger.ErrorContext(ctx, "add cart item", slog.Any("cart id", cartID), slog.Any("error", err))
		return nil, fmt.Errorf("s.AddItem: %w", err)
	}

	return cartItem, nil
}

func (s *CartService) UpdateCartItem(ctx context.Context, cartID, itemID uuid.UUID, newProduct string, newPrice float64) (*entity.CartItem, error) {
	cartItem, err := s.repo.UpdateCartItem(ctx, cartID, itemID, newProduct, newPrice)
	if err != nil {
		if errors.Is(err, errs.ErrCartItemNotFound) {
			return nil, fmt.Errorf("s.UpdateItem: %w", err)
		}
		if errors.Is(err, errs.ErrCartNotFound) {
			return nil, fmt.Errorf("s.UpdateItem: %w", err)
		}
		s.logger.ErrorContext(ctx, "update cart item", slog.Any("cart id", cartID), slog.Any("item id", itemID), slog.Any("error", err))
		return nil, fmt.Errorf("s.UpdateItem: %w", err)
	}

	return cartItem, nil
}

func (s *CartService) RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	err := s.repo.RemoveCartItem(ctx, cartID, itemID)
	if err != nil {
		if errors.Is(err, errs.ErrCartItemNotFound) {
			return fmt.Errorf("s.RemoveItem: %w", err)
		}
		if errors.Is(err, errs.ErrCartNotFound) {
			return fmt.Errorf("s.RemoveItem: %w", err)
		}
		s.logger.ErrorContext(ctx, "remove item", slog.Any("cart id", cartID), slog.Any("item id", itemID), slog.Any("error", err))
		return fmt.Errorf("s.RemoveItem: %w", err)
	}
	return nil
}

func (s *CartService) CalculateDiscount(ctx context.Context, cartID uuid.UUID) (uuid.UUID, float64, float64, float64, error) {
	var totalPrice float64
	var discountPercent float64
	var finalPrice float64

	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			return uuid.Nil, 0, 0, 0, fmt.Errorf("s.CalculateDiscount: %w", err)
		}
		if errors.Is(err, errs.ErrCartItemNotFound) {
			return cart.ID, 0, 0, 0, nil
		}
		s.logger.ErrorContext(ctx, "calculate discount", slog.Any("cart id", cartID), slog.Any("error", err))
		return uuid.Nil, 0, 0, 0, fmt.Errorf("s.CalculateDiscount: %w", err)
	}

	for _, item := range cart.Items {
		totalPrice += *item.Price
	}

	if totalPrice > 5000 && len(cart.Items) > 3 {
		discountPercent = 0.1
		finalPrice = totalPrice - (totalPrice * discountPercent)
		return cart.ID, totalPrice, discountPercent, finalPrice, nil
	}

	if totalPrice > 5000 {
		discountPercent = 0.1
		finalPrice = totalPrice - (totalPrice * discountPercent)
	}

	if len(cart.Items) > 3 {
		discountPercent = 0.05
		finalPrice = totalPrice - (totalPrice * discountPercent)
	}

	return cartID, totalPrice, discountPercent, finalPrice, nil
}
