package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
	"github.com/maxon2034/trainee-go-cart-api/internal/service"
	"github.com/maxon2034/trainee-go-cart-api/mocks"
)

// --- 1. TEST CREATE CART ---
func TestCartHandler_Create(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cartID := uuid.New()

	tests := []struct {
		name           string
		buildStubs     func(ms *mocks.MockService)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Success - Create Empty Cart",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CreateCart(gomock.Any()).
					Return(&service.CartDTO{
						ID:    cartID,
						Items: []service.CartItemDTO{},
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var res service.CartDTO
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if res.ID != cartID {
					t.Errorf("Expected cart ID %s, got %s", cartID, res.ID)
				}
				if res.Items == nil || len(res.Items) != 0 {
					t.Errorf("Expected empty items array, got %+v", res.Items)
				}
			},
		},
		{
			name: "Internal Error On Service Failure",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CreateCart(gomock.Any()).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockService(ctrl)
			tt.buildStubs(mockSvc)

			h := handlers.NewCartHandler(mockSvc, discardLogger)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/carts", nil)
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("STATUS MISMATCH: expected %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

// --- 2. TEST ADD TO CART ---
func TestCartHandler_AddItem(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cartID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name           string
		url            string
		body           string
		buildStubs     func(ms *mocks.MockService)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Success - Item Added",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "Shoes", "price": 2500.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "Shoes", 2500.50).
					Return(&service.CartItemDTO{
						ID:      itemID,
						CartID:  cartID,
						Product: "Shoes",
						Price:   2500.50,
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var res service.CartItemDTO
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode response JSON: %v", err)
				}
				if res.CartID != cartID || res.Product != "Shoes" || res.Price != 2500.50 {
					t.Errorf("Unexpected response data: %+v", res)
				}
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/invalid-uuid/items",
			body:           `{"product": "Shoes", "price": 2500.50}`,
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Fail - Malformed JSON Body",
			url:            "/api/v1/carts/" + cartID.String() + "/items",
			body:           `{"product": "Shoes", "price": }`,
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Fail - Cart Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "Shoes", "price": 2500.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "Shoes", 2500.50).
					Return(nil, errs.ErrCartNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Fail - Cart Limit Reached (Already contains 5 products)",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "Socks", "price": 100.0}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "Socks", 100.0).
					Return(nil, errs.ErrFullCart).
					Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Fail - Blank Product Name",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "   ", "price": 2500.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "   ", 2500.50).
					Return(nil, errs.ErrEmptyProduct).
					Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Fail - Non-Positive Price (Zero or Negative)",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "Shoes", "price": -10.0}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "Shoes", -10.0).
					Return(nil, errs.ErrNegativePrice).
					Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockService(ctrl)
			tt.buildStubs(mockSvc)

			h := handlers.NewCartHandler(mockSvc, discardLogger)

			router := http.NewServeMux()
			router.HandleFunc("POST /api/v1/carts/{cart_id}/items", h.AddItem)

			req := httptest.NewRequest(http.MethodPost, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("STATUS MISMATCH: expected %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

// --- 3. TEST VIEW CART ---
func TestCartHandler_View(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cartID := uuid.New()

	tests := []struct {
		name           string
		url            string
		buildStubs     func(ms *mocks.MockService)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Success - View Cart With Items",
			url:  "/api/v1/carts/" + cartID.String(),
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					ViewCart(gomock.Any(), cartID).
					Return(&service.CartDTO{
						ID: cartID,
						Items: []service.CartItemDTO{
							{ID: uuid.New(), CartID: cartID, Product: "Shoes", Price: 2500.50},
							{ID: uuid.New(), CartID: cartID, Product: "Socks", Price: 1200.00},
						},
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var res service.CartDTO
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode response JSON: %v", err)
				}
				if res.ID != cartID {
					t.Errorf("Expected cart ID %s, got %s", cartID, res.ID)
				}
				if len(res.Items) != 2 {
					t.Fatalf("Expected 2 items in cart, got %d", len(res.Items))
				}
				if res.Items[0].Product != "Shoes" || res.Items[1].Product != "Socks" {
					t.Errorf("Items payload mismatched: %+v", res.Items)
				}
			},
		},
		{
			name: "Fail - Cart Does Not Exist",
			url:  "/api/v1/carts/" + cartID.String(),
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					ViewCart(gomock.Any(), cartID).
					Return(nil, errs.ErrCartNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.CartNotFound()) {
					t.Errorf("Expected body %s, got %s", errs.CartNotFound(), body)
				}
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/non-uuid-string",
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.BadCartRequest()) {
					t.Errorf("Expected body %s, got %s", errs.BadCartRequest(), body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockService(ctrl)
			tt.buildStubs(mockSvc)

			h := handlers.NewCartHandler(mockSvc, discardLogger)

			router := http.NewServeMux()
			router.HandleFunc("GET /api/v1/carts/{cart_id}", h.View)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("STATUS MISMATCH: expected %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

func TestCartHandler_UpdateItem(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cartID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name           string
		url            string
		body           string
		buildStubs     func(ms *mocks.MockService)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Success - Item Updated",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "Shoes", "price": 5000.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), cartID, itemID, "Shoes", 5000.50).
					Return(&service.CartItemDTO{
						ID:      itemID,
						CartID:  cartID,
						Product: "Shoes",
						Price:   5000.50,
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var res service.CartItemDTO
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode response JSON: %v", err)
				}
				if res.ID != itemID || res.CartID != cartID || res.Product != "Shoes" || res.Price != 5000.50 {
					t.Errorf("Unexpected response data: %+v", res)
				}
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/invalid-uuid/items/" + itemID.String(),
			body:           `{"product": "Shoes", "price": 5000.50}`,
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.BadCartRequest()) {
					t.Errorf("Expected body %s, got %s", errs.BadCartRequest(), body)
				}
			},
		},
		{
			name:           "Fail - Invalid Item UUID",
			url:            "/api/v1/carts/" + cartID.String() + "/items/invalid-uuid",
			body:           `{"product": "Shoes", "price": 5000.50}`,
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.BadItemRequest()) {
					t.Errorf("Expected body %s, got %s", errs.BadItemRequest(), body)
				}
			},
		},
		{
			name:           "Fail - Malformed JSON Body",
			url:            "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body:           `{"product": "Shoes", "price": }`,
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.BadRequest()) {
					t.Errorf("Expected body %s, got %s", errs.BadRequest(), body)
				}
			},
		},
		{
			name: "Fail - Cart Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "Shoes", "price": 5000.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), cartID, itemID, "Shoes", 5000.50).
					Return(nil, errs.ErrCartNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.CartNotFound()) {
					t.Errorf("Expected body %s, got %s", errs.CartNotFound(), body)
				}
			},
		},
		{
			name: "Fail - Item Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "Shoes", "price": 5000.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), cartID, itemID, "Shoes", 5000.50).
					Return(nil, errs.ErrCartItemNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.ItemNotFound()) {
					t.Errorf("Expected body %s, got %s", errs.ItemNotFound(), body)
				}
			},
		},
		{
			name: "Fail - Empty Product Name",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "", "price": 5000.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), cartID, itemID, "", 5000.50).
					Return(nil, errs.ErrEmptyProduct).
					Times(1)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.EmptyProduct()) {
					t.Errorf("Expected body %s, got %s", errs.EmptyProduct(), body)
				}
			},
		},
		{
			name: "Fail - Negative Price",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "Shoes", "price": -100.0}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), cartID, itemID, "Shoes", -100.0).
					Return(nil, errs.ErrNegativePrice).
					Times(1)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.NegativePrice()) {
					t.Errorf("Expected body %s, got %s", errs.NegativePrice(), body)
				}
			},
		},
		{
			name: "Fail - Database / Unexpected Error",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "Shoes", "price": 5000.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), cartID, itemID, "Shoes", 5000.50).
					Return(nil, errors.New("db connection failure")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockService(ctrl)
			tt.buildStubs(mockSvc)

			h := handlers.NewCartHandler(mockSvc, discardLogger)

			router := http.NewServeMux()
			router.HandleFunc("PUT /api/v1/carts/{cart_id}/items/{item_id}", h.UpdateItem)

			req := httptest.NewRequest(http.MethodPut, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("STATUS MISMATCH: expected %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

func TestCartHandler_DeleteItem(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cartID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name           string
		url            string
		buildStubs     func(ms *mocks.MockService)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Success - Item Deleted",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					RemoveItem(gomock.Any(), cartID, itemID).
					Return(nil).
					Times(1)
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(t *testing.T, body []byte) {
				if len(body) != 0 {
					t.Errorf("Expected empty response body for 204 No Content, got: %s", body)
				}
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/invalid-uuid/items/" + itemID.String(),
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.BadCartRequest()) {
					t.Errorf("Expected body %s, got %s", errs.BadCartRequest(), body)
				}
			},
		},
		{
			name:           "Fail - Invalid Item UUID",
			url:            "/api/v1/carts/" + cartID.String() + "/items/invalid-uuid",
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.BadItemRequest()) {
					t.Errorf("Expected body %s, got %s", errs.BadItemRequest(), body)
				}
			},
		},
		{
			name: "Fail - Cart Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					RemoveItem(gomock.Any(), cartID, itemID).
					Return(errs.ErrCartNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.CartNotFound()) {
					t.Errorf("Expected body %s, got %s", errs.CartNotFound(), body)
				}
			},
		},
		{
			name: "Fail - Item Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					RemoveItem(gomock.Any(), cartID, itemID).
					Return(errs.ErrCartItemNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.ItemNotFound()) {
					t.Errorf("Expected body %s, got %s", errs.ItemNotFound(), body)
				}
			},
		},
		{
			name: "Fail - Database / Unexpected Error",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					RemoveItem(gomock.Any(), cartID, itemID).
					Return(errors.New("db delete error")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockService(ctrl)
			tt.buildStubs(mockSvc)

			h := handlers.NewCartHandler(mockSvc, discardLogger)

			router := http.NewServeMux()
			router.HandleFunc("DELETE /api/v1/carts/{cart_id}/items/{item_id}", h.DeleteItem)

			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("STATUS MISMATCH: expected %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

func TestCartHandler_CalculateDiscount(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cartID := uuid.New()

	validCartDiscount := &entity.CartDiscount{
		CartID:          cartID,
		TotalPrice:      6000.0,
		DiscountPercent: 0.1,
		FinalPrice:      5400.0,
	}

	tests := []struct {
		name           string
		url            string
		buildStubs     func(ms *mocks.MockService)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "Success - Calculated Discount",
			url:  "/api/v1/carts/" + cartID.String() + "/discount",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CalculateDiscount(gomock.Any(), cartID).
					Return(validCartDiscount, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp entity.CartDiscount
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response body should be valid JSON")
				assert.Equal(t, validCartDiscount.CartID, resp.CartID)
				assert.Equal(t, validCartDiscount.TotalPrice, resp.TotalPrice)
				assert.Equal(t, validCartDiscount.DiscountPercent, resp.DiscountPercent)
				assert.Equal(t, validCartDiscount.FinalPrice, resp.FinalPrice)
			},
		},
		{
			name: "Success - Empty Cart (Zero Discount)",
			url:  "/api/v1/carts/" + cartID.String() + "/discount",
			buildStubs: func(ms *mocks.MockService) {
				emptyCartDiscount := &entity.CartDiscount{
					CartID:          cartID,
					TotalPrice:      0,
					DiscountPercent: 0,
					FinalPrice:      0,
				}
				ms.EXPECT().
					CalculateDiscount(gomock.Any(), cartID).
					Return(emptyCartDiscount, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp entity.CartDiscount
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.Equal(t, cartID, resp.CartID)
				assert.Equal(t, 0.0, resp.TotalPrice)
				assert.Equal(t, 0.0, resp.DiscountPercent)
				assert.Equal(t, 0.0, resp.FinalPrice)
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/invalid-uuid/discount",
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.BadCartRequest()) {
					t.Errorf("Expected body %s, got %s", errs.BadCartRequest(), body)
				}
			},
		},
		{
			name: "Fail - Cart Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/discount",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CalculateDiscount(gomock.Any(), cartID).
					Return(nil, errs.ErrCartNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.CartNotFound()) {
					t.Errorf("Expected body %s, got %s", errs.CartNotFound(), body)
				}
			},
		},
		{
			name: "Fail - Database / Internal Error",
			url:  "/api/v1/carts/" + cartID.String() + "/discount",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CalculateDiscount(gomock.Any(), cartID).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body []byte) {
				assert.Empty(t, body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockService(ctrl)
			tt.buildStubs(mockSvc)

			h := handlers.NewCartHandler(mockSvc, discardLogger)

			router := http.NewServeMux()
			router.HandleFunc("GET /api/v1/carts/{cart_id}/discount", h.CalculateDiscount)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("STATUS MISMATCH: expected %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}
