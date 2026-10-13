package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
	"github.com/shopspring/decimal"
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

func (s *CartService) AddItem(ctx context.Context, cartID uuid.UUID, product string, price decimal.Decimal) (*entity.CartItem, error) {
	cartItem, err := s.repo.AddCartItem(ctx, cartID, product, price, ItemLimit)
	if err != nil {
		s.logger.ErrorContext(ctx, "add cart item", slog.Any("cart id", cartID), slog.Any("error", err))
		return nil, fmt.Errorf("s.AddItem: %w", err)
	}

	return cartItem, nil
}

func (s *CartService) UpdateCartItem(ctx context.Context, cartID, itemID uuid.UUID, newProduct string, newPrice decimal.Decimal) (*entity.CartItem, error) {
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

func (s *CartService) CalculateDiscount(ctx context.Context, cartID uuid.UUID) (*entity.CartDiscount, error) {
	var cartDiscount entity.CartDiscount

	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			return nil, fmt.Errorf("s.CalculateDiscount: %w", err)
		}
		s.logger.ErrorContext(ctx, "calculate discount", slog.Any("cart id", cartID), slog.Any("error", err))
		return nil, fmt.Errorf("s.CalculateDiscount: %w", err)
	}
	cartDiscount.TotalPrice = decimal.Zero
	if len(cart.Items) != 0 {
		for _, item := range cart.Items {
			cartDiscount.TotalPrice = cartDiscount.TotalPrice.Add(item.Price)
		}
	}

	cartDiscount.CartID = cartID
	cartDiscount.DiscountPercent = 0
	cartDiscount.FinalPrice = cartDiscount.TotalPrice

	switch {
	case cartDiscount.TotalPrice.GreaterThan(s.cfg.DiscountTotalPrice):
		cartDiscount.DiscountPercent = s.cfg.DiscountPercentBig
		cartDiscount.FinalPrice = cartDiscount.TotalPrice.Sub(cartDiscount.TotalPrice.Mul(decimal.NewFromFloat(s.cfg.DiscountPercentBig)))
	case len(cart.Items) > s.cfg.DiscountItemAmount:
		cartDiscount.DiscountPercent = s.cfg.DiscountPercentSmall
		cartDiscount.FinalPrice = cartDiscount.TotalPrice.Sub(cartDiscount.TotalPrice.Mul(decimal.NewFromFloat(s.cfg.DiscountPercentSmall)))
	}

	return &cartDiscount, nil
}
