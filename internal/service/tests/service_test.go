package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
	"github.com/maxon2034/trainee-go-cart-api/internal/service"
	"github.com/maxon2034/trainee-go-cart-api/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCartService_CreateCart(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

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
		cartService := service.New(mockRepo, logger)

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
	logger := slog.New(slog.DiscardHandler)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Headphones"
		price := 120.50

		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{
					ID:      &itemID,
					CartID:  cartID,
					Product: &product,
					Price:   &price,
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

		require.NotNil(t, cartDTO.Items[0].ID)
		assert.Equal(t, itemID, *cartDTO.Items[0].ID)
		require.NotNil(t, cartDTO.Items[0].Product)
		assert.Equal(t, "Headphones", *cartDTO.Items[0].Product)
		require.NotNil(t, cartDTO.Items[0].Price)
		assert.Equal(t, 120.50, *cartDTO.Items[0].Price)
	})

	t.Run("cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

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
		cartService := service.New(mockRepo, logger)

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
	logger := slog.New(slog.DiscardHandler)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Mechanical Keyboard"
		price := 150.00

		expectedEntityItem := &entity.CartItem{
			ID:      &itemID,
			CartID:  cartID,
			Product: &product,
			Price:   &price,
		}

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, product, price, service.ItemLimit).
			Return(expectedEntityItem, nil).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, product, price)

		require.NoError(t, err)
		require.NotNil(t, itemDTO)
		require.NotNil(t, itemDTO.ID)
		assert.Equal(t, itemID, *itemDTO.ID)
		assert.Equal(t, cartID, itemDTO.CartID)
		require.NotNil(t, itemDTO.Product)
		assert.Equal(t, product, *itemDTO.Product)
		require.NotNil(t, itemDTO.Price)
		assert.Equal(t, price, *itemDTO.Price)
	})

	t.Run("error - full cart", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		product := "Mouse"
		price := 50.00

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
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		expectedErr := errors.New("db connection lost")

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, "Monitor", 300.00, service.ItemLimit).
			Return(nil, expectedErr).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, "Monitor", 300.00)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.AddItem")
		assert.Nil(t, itemDTO)
	})
}
func TestCartService_UpdateCartItem(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Shoes"
		newPrice := 5000.50

		expectedEntityItem := &entity.CartItem{
			ID:      &itemID,
			CartID:  cartID,
			Product: &product,
			Price:   &newPrice,
		}

		mockRepo.EXPECT().
			UpdateCartItem(gomock.Any(), cartID, itemID, product, newPrice).
			Return(expectedEntityItem, nil).
			Times(1)

		itemDTO, err := cartService.UpdateCartItem(context.Background(), cartID, itemID, product, newPrice)

		require.NoError(t, err)
		require.NotNil(t, itemDTO)
		require.NotNil(t, itemDTO.ID)
		assert.Equal(t, itemID, *itemDTO.ID)
		assert.Equal(t, cartID, itemDTO.CartID)
		require.NotNil(t, itemDTO.Product)
		assert.Equal(t, product, *itemDTO.Product)
		require.NotNil(t, itemDTO.Price)
		assert.Equal(t, newPrice, *itemDTO.Price)
	})

	t.Run("error - cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		nonExistentCartID := uuid.New()
		itemID := uuid.New()
		product := "Shoes"
		newPrice := 5000.50

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
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		nonExistentItemID := uuid.New()
		product := "NonExistentShoes"
		newPrice := 5000.50

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
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Shoes"
		newPrice := 5000.50
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
	logger := slog.New(slog.DiscardHandler)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

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
		cartService := service.New(mockRepo, logger)

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
		cartService := service.New(mockRepo, logger)

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
		cartService := service.New(mockRepo, logger)

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
	logger := slog.New(slog.DiscardHandler)

	t.Run("success - discount 10% (price > 5000, items <= 3)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		price := 6000.0
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: &price},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		id, total, percent, final, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		assert.Equal(t, cartID, id)
		assert.Equal(t, 6000.0, total)
		assert.Equal(t, 0.1, percent)
		assert.Equal(t, 5400.0, final)
	})

	t.Run("success - discount 5% (price <= 5000, items > 3)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		price := 1000.0
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: &price},
				{Price: &price},
				{Price: &price},
				{Price: &price},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		id, total, percent, final, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		assert.Equal(t, cartID, id)
		assert.Equal(t, 4000.0, total)
		assert.Equal(t, 0.05, percent)
		assert.Equal(t, 3800.0, final)
	})

	t.Run("success - combined condition discount 10% (price > 5000 AND items > 3)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		price := 1500.0
		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{Price: &price},
				{Price: &price},
				{Price: &price},
				{Price: &price},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		id, total, percent, final, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		assert.Equal(t, cartID, id)
		assert.Equal(t, 6000.0, total)
		assert.Equal(t, 0.1, percent)
		assert.Equal(t, 5400.0, final)
	})

	t.Run("success - empty cart returns zero discount", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		expectedCart := &entity.Cart{
			ID:    cartID,
			Items: []entity.CartItem{},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		id, total, percent, final, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.NoError(t, err)
		assert.Equal(t, cartID, id)
		assert.Equal(t, 0.0, total)
		assert.Equal(t, 0.0, percent)
		assert.Equal(t, 0.0, final)
	})

	t.Run("error - cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		nonExistentCartID := uuid.New()

		mockRepo.EXPECT().
			GetCart(gomock.Any(), nonExistentCartID).
			Return(nil, errs.ErrCartNotFound).
			Times(1)

		id, total, percent, final, err := cartService.CalculateDiscount(context.Background(), nonExistentCartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.ErrorContains(t, err, "s.CalculateDiscount")
		assert.Equal(t, uuid.Nil, id)
		assert.Equal(t, 0.0, total)
		assert.Equal(t, 0.0, percent)
		assert.Equal(t, 0.0, final)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo, logger)

		cartID := uuid.New()
		expectedErr := errors.New("unexpected database error")

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(nil, expectedErr).
			Times(1)

		id, total, percent, final, err := cartService.CalculateDiscount(context.Background(), cartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorContains(t, err, "s.CalculateDiscount")
		assert.Equal(t, uuid.Nil, id)
		assert.Equal(t, 0.0, total)
		assert.Equal(t, 0.0, percent)
		assert.Equal(t, 0.0, final)
	})
}
