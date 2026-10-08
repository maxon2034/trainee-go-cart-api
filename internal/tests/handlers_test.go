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
	"github.com/maxon2034/trainee-go-cart-api/internal/handlers"
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
					Return(service.CartDTO{
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
					Return(service.CartDTO{}, errors.New("db error")).
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
					Return(service.CartItemDTO{
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
			expectedStatus: http.StatusBadRequest, // СТРОГО 400! Упадет, если в хэндлере нет w.WriteHeader(http.StatusBadRequest)
		},
		{
			name: "Fail - Cart Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "Shoes", "price": 2500.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "Shoes", 2500.50).
					Return(service.CartItemDTO{}, errs.ErrCartNotFound).
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
					Return(service.CartItemDTO{}, errs.ErrFullCart).
					Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Fail - Blank Product Name",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "   ", "price": 2500.50}`,
			buildStubs: func(ms *mocks.MockService) {
				// Если валидация на уровне сервиса:
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "   ", 2500.50).
					Return(service.CartItemDTO{}, errs.ErrEmptyProduct).
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
					Return(service.CartItemDTO{}, errs.ErrNegativePrice).
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
					Return(service.CartDTO{
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
					Return(service.CartDTO{}, errs.ErrCartNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/non-uuid-string",
			buildStubs:     func(ms *mocks.MockService) { /* Сервис не вызовется */ },
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
			router.HandleFunc("GET /api/v1/carts/{id}", h.View)

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
					UpdateCartItem(gomock.Any(), itemID, "Shoes", 5000.50).
					Return(service.CartItemDTO{
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
				if res.ID != itemID || res.Product != "Shoes" || res.Price != 5000.50 {
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
			name: "Fail - Item Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "Shoes", "price": 5000.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), itemID, "Shoes", 5000.50).
					Return(service.CartItemDTO{}, errs.ErrCartItemNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				if !bytes.Equal(body, errs.NotFound()) {
					t.Errorf("Expected body %s, got %s", errs.NotFound(), body)
				}
			},
		},
		{
			name: "Fail - Empty Product Name",
			url:  "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body: `{"product": "", "price": 5000.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					UpdateCartItem(gomock.Any(), itemID, "", 5000.50).
					Return(service.CartItemDTO{}, errs.ErrEmptyProduct).
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
					UpdateCartItem(gomock.Any(), itemID, "Shoes", -100.0).
					Return(service.CartItemDTO{}, errs.ErrNegativePrice).
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
					UpdateCartItem(gomock.Any(), itemID, "Shoes", 5000.50).
					Return(service.CartItemDTO{}, errors.New("db connection failure")).
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
			router.HandleFunc("PUT /api/v1/carts/{cart_id}/items/{id}", h.UpdateItem)

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
