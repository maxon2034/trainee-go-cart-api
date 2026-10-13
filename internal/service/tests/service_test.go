package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/config"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
	"github.com/maxon2034/trainee-go-cart-api/internal/service"
	"github.com/maxon2034/trainee-go-cart-api/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestService(repo service.Repository) *service.CartService {
	cfg := config.CartConfig{
		DiscountPercentBig:   0.1,
		DiscountPercentSmall: 0.05,
		DiscountTotalPrice:   decimal.NewFromInt(5000),
		DiscountItemAmount:   3,
	}

	return service.New(repo, slog.New(slog.DiscardHandler), cfg)
}

func assertDecimalEqual(t *testing.T, expected, actual decimal.Decimal) {
	t.Helper()
	assert.Truef(t, expected.Equal(actual), "expected %s, got %s", expected, actual)
}

func TestCartService_CreateCart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		expectedID := uuid.New()
		expectedCart := &entity.Cart{
			ID:    expectedID,
			Items: []entity.CartItem{},
		}

		mockRepo.EXPECT().
			AddCart(gomock.Any()).
			Return(expectedCart, nil).
			Times(1)

		cartDTO, err := cartService.CreateCart(context.Background())

		require.NoError(t, err)
		require.NotNil(t, cartDTO)
		assert.Equal(t, expectedID, cartDTO.ID)
		assert.Empty(t, cartDTO.Items)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		expectedErr := errors.New("db insert error")

		mockRepo.EXPECT().
			AddCart(gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		cartDTO, err := cartService.CreateCart(context.Background())

		require.Error(t, err)
		assert.Nil(t, cartDTO)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.CreateCart")
	})
}

func TestCartService_ViewCart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Headphones"
		price := decimal.RequireFromString("120.50")

		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{
					ID:      itemID,
					CartID:  cartID,
					Product: product,
					Price:   price,
				},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		cartDTO, err := cartService.ViewCart(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, cartDTO)
		assert.Equal(t, cartID, cartDTO.ID)
		require.Len(t, cartDTO.Items, 1)

		assert.Equal(t, itemID, cartDTO.Items[0].ID)
		assert.Equal(t, cartID, cartDTO.Items[0].CartID)
		assert.Equal(t, product, cartDTO.Items[0].Product)
		assertDecimalEqual(t, price, cartDTO.Items[0].Price)
	})

	t.Run("cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(nil, errs.ErrCartNotFound).
			Times(1)

		cartDTO, err := cartService.ViewCart(context.Background(), cartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.ErrorContains(t, err, "s.ViewCart")
		assert.Nil(t, cartDTO)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		expectedErr := errors.New("db connection timeout")

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(nil, expectedErr).
			Times(1)

		cartDTO, err := cartService.ViewCart(context.Background(), cartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.ViewCart")
		assert.Nil(t, cartDTO)
	})
}

func TestCartService_AddItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Mechanical Keyboard"
		price := decimal.RequireFromString("150.00")

		expectedEntityItem := &entity.CartItem{
			ID:      itemID,
			CartID:  cartID,
			Product: product,
			Price:   price,
		}

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, product, price, service.ItemLimit).
			Return(expectedEntityItem, nil).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, product, price)

		require.NoError(t, err)
		require.NotNil(t, itemDTO)
		assert.Equal(t, itemID, itemDTO.ID)
		assert.Equal(t, cartID, itemDTO.CartID)
		assert.Equal(t, product, itemDTO.Product)
		assertDecimalEqual(t, price, itemDTO.Price)
	})

	t.Run("error - full cart", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		product := "Mouse"
		price := decimal.RequireFromString("50.00")

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, product, price, service.ItemLimit).
			Return(nil, errs.ErrFullCart).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, product, price)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrFullCart)
		assert.ErrorContains(t, err, "s.AddItem")
		assert.Nil(t, itemDTO)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		price := decimal.RequireFromString("300.00")
		expectedErr := errors.New("db connection lost")

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, "Monitor", price, service.ItemLimit).
			Return(nil, expectedErr).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, "Monitor", price)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.AddItem")
		assert.Nil(t, itemDTO)
	})
}

func TestCartService_UpdateCartItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Shoes"
		newPrice := decimal.RequireFromString("5000.50")

		expectedEntityItem := &entity.CartItem{
			ID:      itemID,
			CartID:  cartID,
			Product: product,
			Price:   newPrice,
		}

		mockRepo.EXPECT().
			UpdateCartItem(gomock.Any(), cartID, itemID, product, newPrice).
			Return(expectedEntityItem, nil).
			Times(1)

		itemDTO, err := cartService.UpdateCartItem(context.Background(), cartID, itemID, product, newPrice)

		require.NoError(t, err)
		require.NotNil(t, itemDTO)
		assert.Equal(t, itemID, itemDTO.ID)
		assert.Equal(t, cartID, itemDTO.CartID)
		assert.Equal(t, product, itemDTO.Product)
		assertDecimalEqual(t, newPrice, itemDTO.Price)
	})

	t.Run("error - cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		nonExistentCartID := uuid.New()
		itemID := uuid.New()
		product := "Shoes"
		newPrice := decimal.RequireFromString("5000.50")

		mockRepo.EXPECT().
			UpdateCartItem(gomock.Any(), nonExistentCartID, itemID, product, newPrice).
			Return(nil, errs.ErrCartNotFound).
			Times(1)

		itemDTO, err := cartService.UpdateCartItem(context.Background(), nonExistentCartID, itemID, product, newPrice)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.ErrorContains(t, err, "s.UpdateItem")
		assert.Nil(t, itemDTO)
	})

	t.Run("error - cart item not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		nonExistentItemID := uuid.New()
		product := "NonExistentShoes"
		newPrice := decimal.RequireFromString("5000.50")

		mockRepo.EXPECT().
			UpdateCartItem(gomock.Any(), cartID, nonExistentItemID, product, newPrice).
			Return(nil, errs.ErrCartItemNotFound).
			Times(1)

		itemDTO, err := cartService.UpdateCartItem(context.Background(), cartID, nonExistentItemID, product, newPrice)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartItemNotFound)
		assert.ErrorContains(t, err, "s.UpdateItem")
		assert.Nil(t, itemDTO)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Shoes"
		newPrice := decimal.RequireFromString("5000.50")
		expectedErr := errors.New("db query failed")

		mockRepo.EXPECT().
			UpdateCartItem(gomock.Any(), cartID, itemID, product, newPrice).
			Return(nil, expectedErr).
			Times(1)

		itemDTO, err := cartService.UpdateCartItem(context.Background(), cartID, itemID, product, newPrice)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.UpdateItem")
		assert.Nil(t, itemDTO)
	})
}

func TestCartService_RemoveItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()

		mockRepo.EXPECT().
			RemoveCartItem(gomock.Any(), cartID, itemID).
			Return(nil).
			Times(1)

		err := cartService.RemoveItem(context.Background(), cartID, itemID)

		require.NoError(t, err)
	})

	t.Run("error - cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		nonExistentCartID := uuid.New()
		itemID := uuid.New()

		mockRepo.EXPECT().
			RemoveCartItem(gomock.Any(), nonExistentCartID, itemID).
			Return(errs.ErrCartNotFound).
			Times(1)

		err := cartService.RemoveItem(context.Background(), nonExistentCartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.ErrorContains(t, err, "s.RemoveItem")
	})

	t.Run("error - cart item not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		nonExistentItemID := uuid.New()

		mockRepo.EXPECT().
			RemoveCartItem(gomock.Any(), cartID, nonExistentItemID).
			Return(errs.ErrCartItemNotFound).
			Times(1)

		err := cartService.RemoveItem(context.Background(), cartID, nonExistentItemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartItemNotFound)
		assert.ErrorContains(t, err, "s.RemoveItem")
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()
		expectedErr := errors.New("unexpected database error")

		mockRepo.EXPECT().
			RemoveCartItem(gomock.Any(), cartID, itemID).
			Return(expectedErr).
			Times(1)

		err := cartService.RemoveItem(context.Background(), cartID, itemID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.RemoveItem")
	})
}

func TestCartService_CalculateDiscount(t *testing.T) {
	t.Run("success - discount 10% (price > 5000, items <= 3)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		price := decimal.RequireFromString("6000")
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: price},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, discount)
		assertDecimalEqual(t, decimal.RequireFromString("6000"), discount.TotalPrice)
		assertDecimalEqual(t, decimal.RequireFromString("5400"), discount.FinalPrice)
	})

	t.Run("success - discount 5% (price <= 5000, items > 3)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		price := decimal.RequireFromString("1000")
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: price},
				{Price: price},
				{Price: price},
				{Price: price},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, discount)
		assertDecimalEqual(t, decimal.RequireFromString("4000"), discount.TotalPrice)
		assertDecimalEqual(t, decimal.RequireFromString("3800"), discount.FinalPrice)
	})

	t.Run("success - combined condition discount 10% (price > 5000 AND items > 3)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		price := decimal.RequireFromString("1500")
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: price},
				{Price: price},
				{Price: price},
				{Price: price},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, discount)
		assertDecimalEqual(t, decimal.RequireFromString("6000"), discount.TotalPrice)
		assertDecimalEqual(t, decimal.RequireFromString("5400"), discount.FinalPrice)
	})

	t.Run("success - total just above threshold is discounted without precision loss", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: decimal.RequireFromString("5000.01")},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, discount)
		assertDecimalEqual(t, decimal.RequireFromString("5000.01"), discount.TotalPrice)
		assertDecimalEqual(t, decimal.RequireFromString("4500.009"), discount.FinalPrice)
	})

	t.Run("success - no discount at boundaries (price == 5000, items == 3)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: decimal.RequireFromString("2000")},
				{Price: decimal.RequireFromString("2000")},
				{Price: decimal.RequireFromString("1000")},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, discount)
		assertDecimalEqual(t, decimal.RequireFromString("5000"), discount.TotalPrice)
		assertDecimalEqual(t, decimal.RequireFromString("5000"), discount.FinalPrice)
	})

	t.Run("success - empty cart returns zero discount", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		expectedCart := &entity.Cart{
			ID:    cartID,
			Items: []entity.CartItem{},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		require.NotNil(t, discount)
		assertDecimalEqual(t, decimal.Zero, discount.TotalPrice)
		assertDecimalEqual(t, decimal.Zero, discount.FinalPrice)
	})

	t.Run("error - cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		nonExistentCartID := uuid.New()

		mockRepo.EXPECT().
			GetCart(gomock.Any(), nonExistentCartID).
			Return(nil, errs.ErrCartNotFound).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), nonExistentCartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.ErrorContains(t, err, "s.CalculateDiscount")
		assert.Nil(t, discount)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := newTestService(mockRepo)

		cartID := uuid.New()
		expectedErr := errors.New("unexpected database error")

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(nil, expectedErr).
			Times(1)

		discount, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.CalculateDiscount")
		assert.Nil(t, discount)
	})
}
