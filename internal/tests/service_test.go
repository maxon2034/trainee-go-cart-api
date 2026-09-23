package tests

import (
	"context"
	"errors"
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
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

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

		// 5. Проверяем результат
		require.NoError(t, err)
		assert.Equal(t, expectedID, cartDTO.ID)
		assert.Empty(t, cartDTO.Items)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		expectedErr := errors.New("db insert error")

		mockRepo.EXPECT().
			AddCart(gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		cartDTO, err := cartService.CreateCart(context.Background())

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Empty(t, cartDTO)
	})
}

func TestCartService_ViewCart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()

		expectedCart := &entity.Cart{
			ID: cartID,
			Items: []entity.CartItem{
				{
					ID:      itemID,
					CartID:  cartID,
					Product: "Headphones",
					Price:   120.50,
				},
			},
		}

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(expectedCart, nil).
			Times(1)

		cartDTO, err := cartService.ViewCart(context.Background(), cartID)

		require.NoError(t, err)
		assert.Equal(t, cartID, cartDTO.ID)
		assert.Len(t, cartDTO.Items, 1)

		assert.Equal(t, itemID, cartDTO.Items[0].ID)
		assert.Equal(t, "Headphones", cartDTO.Items[0].Product)
		assert.Equal(t, 120.50, cartDTO.Items[0].Price)
	})

	t.Run("cart not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(nil, errs.ErrCartNotFound).
			Times(1)

		cartDTO, err := cartService.ViewCart(context.Background(), cartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrCartNotFound)
		assert.Empty(t, cartDTO)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()
		expectedErr := errors.New("db connection timeout")

		mockRepo.EXPECT().
			GetCart(gomock.Any(), cartID).
			Return(nil, expectedErr).
			Times(1)

		cartDTO, err := cartService.ViewCart(context.Background(), cartID)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Empty(t, cartDTO)
	})
}

func TestCartService_AddItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()
		itemID := uuid.New()
		product := "Mechanical Keyboard"
		price := 150.00

		expectedEntityItem := &entity.CartItem{
			ID:      itemID,
			CartID:  cartID,
			Product: product,
			Price:   price,
		}

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, product, price).
			Return(expectedEntityItem, nil).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, product, price)

		require.NoError(t, err)
		assert.Equal(t, itemID, itemDTO.ID)
		assert.Equal(t, cartID, itemDTO.CartID)
		assert.Equal(t, product, itemDTO.Product)
		assert.Equal(t, price, itemDTO.Price)
	})

	t.Run("error - full cart", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()
		product := "Mouse"
		price := 50.00

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, product, price).
			Return(nil, errs.ErrFullCart).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, product, price)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrFullCart)
		assert.Empty(t, itemDTO)
	})

	t.Run("error - empty product", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, "", 50.00).
			Return(nil, errs.ErrEmptyProduct).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, "", 50.00)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrEmptyProduct)
		assert.Empty(t, itemDTO)
	})

	t.Run("error - negative price", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, "Laptop", -100.00).
			Return(nil, errs.ErrNegativePrice).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, "Laptop", -100.00)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrNegativePrice)
		assert.Empty(t, itemDTO)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		cartService := service.New(mockRepo)

		cartID := uuid.New()
		expectedErr := errors.New("db connection lost")

		mockRepo.EXPECT().
			AddCartItem(gomock.Any(), cartID, "Monitor", 300.00).
			Return(nil, expectedErr).
			Times(1)

		itemDTO, err := cartService.AddItem(context.Background(), cartID, "Monitor", 300.00)

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Empty(t, itemDTO)
	})
}
