package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/config"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
	"github.com/maxon2034/trainee-go-cart-api/internal/handlers"
	"go.uber.org/mock/gomock"

	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
	"github.com/maxon2034/trainee-go-cart-api/mocks"
)

func TestCartHandler_Create(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.ServerConfig{
		CtxDefaultTimeout: 5 * time.Second,
	}
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
					Return(&entity.Cart{
						ID:    cartID,
						Items: []entity.CartItem{},
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.CreateCartResponse
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
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}
				if res.Status != "INTERNAL_SERVER_ERROR" {
					t.Errorf("Expected status INTERNAL_SERVER_ERROR, got %s", res.Status)
				}
				if res.Message != "internal server error" {
					t.Errorf("Expected message 'internal server error', got %s", res.Message)
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

			h := handlers.NewCartHandler(mockSvc, discardLogger, cfg)

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

func TestCartHandler_View(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.ServerConfig{
		CtxDefaultTimeout: 5 * time.Second,
	}
	cartID := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()
	product1 := "Shoes"
	price1 := 2500.50
	product2 := "Socks"
	price2 := 1200.00

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
					Return(&entity.Cart{
						ID: cartID,
						Items: []entity.CartItem{
							{ID: &itemID1, CartID: cartID, Product: &product1, Price: &price1},
							{ID: &itemID2, CartID: cartID, Product: &product2, Price: &price2},
						},
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ViewCartResponse
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
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "CART_NOT_FOUND" {
					t.Errorf("Expected status CART_NOT_FOUND, got %s", res.Status)
				}
				if res.Message != "cart not found" {
					t.Errorf("Expected message 'cart not found', got %s", res.Message)
				}
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/non-uuid-string",
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
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

			h := handlers.NewCartHandler(mockSvc, discardLogger, cfg)

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

func TestCartHandler_AddItem(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.ServerConfig{
		CtxDefaultTimeout: 5 * time.Second,
	}
	cartID := uuid.New()
	itemID := uuid.New()
	product := "Shoes"
	price := 2500.50

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
					Return(&entity.CartItem{
						ID:      &itemID,
						CartID:  cartID,
						Product: &product,
						Price:   &price,
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.AddItemResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode response JSON: %v", err)
				}
				if res.ID != itemID || res.CartID != cartID || res.Product != "Shoes" || res.Price != 2500.50 {
					t.Errorf("Unexpected response data: %+v", res)
				}
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/invalid-uuid/items",
			body:           `{"product": "Shoes", "price": 2500.50}`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Malformed JSON Body",
			url:            "/api/v1/carts/" + cartID.String() + "/items",
			body:           `{"product": "Shoes", "price": }`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
				}
			},
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
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "CART_NOT_FOUND" {
					t.Errorf("Expected status CART_NOT_FOUND, got %s", res.Status)
				}
			},
		},
		{
			name: "Fail - Full Cart Limit Reached",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "Socks", "price": 100.0}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "Socks", 100.0).
					Return(nil, errs.ErrFullCart).
					Times(1)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "FULL_CART" {
					t.Errorf("Expected status FULL_CART, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Empty Product Name",
			url:            "/api/v1/carts/" + cartID.String() + "/items",
			body:           `{"product": "", "price": 2500.50}`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "EMPTY_PRODUCT" {
					t.Errorf("Expected status EMPTY_PRODUCT, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Negative Price",
			url:            "/api/v1/carts/" + cartID.String() + "/items",
			body:           `{"product": "Shoes", "price": -100.0}`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "INVALID_PRICE" {
					t.Errorf("Expected status INVALID_PRICE, got %s", res.Status)
				}
			},
		},
		{
			name: "Fail - Database / Unexpected Error",
			url:  "/api/v1/carts/" + cartID.String() + "/items",
			body: `{"product": "Shoes", "price": 2500.50}`,
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					AddItem(gomock.Any(), cartID, "Shoes", 2500.50).
					Return(nil, errors.New("db connection failure")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "INTERNAL_SERVER_ERROR" {
					t.Errorf("Expected status INTERNAL_SERVER_ERROR, got %s", res.Status)
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

			h := handlers.NewCartHandler(mockSvc, discardLogger, cfg)

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

func TestCartHandler_UpdateItem(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.ServerConfig{
		CtxDefaultTimeout: 5 * time.Second,
	}
	cartID := uuid.New()
	itemID := uuid.New()
	product := "Shoes"
	price := 5000.50

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
					Return(&entity.CartItem{
						ID:      &itemID,
						CartID:  cartID,
						Product: &product,
						Price:   &price,
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.UpdateItemResponse
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
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Invalid Item UUID",
			url:            "/api/v1/carts/" + cartID.String() + "/items/invalid-uuid",
			body:           `{"product": "Shoes", "price": 5000.50}`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Malformed JSON Body",
			url:            "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body:           `{"product": "Shoes", "price": }`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
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
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "CART_NOT_FOUND" {
					t.Errorf("Expected status CART_NOT_FOUND, got %s", res.Status)
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
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "ITEM_NOT_FOUND" {
					t.Errorf("Expected status ITEM_NOT_FOUND, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Empty Product Name",
			url:            "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body:           `{"product": "", "price": 5000.50}`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "EMPTY_PRODUCT" {
					t.Errorf("Expected status EMPTY_PRODUCT, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Negative Price",
			url:            "/api/v1/carts/" + cartID.String() + "/items/" + itemID.String(),
			body:           `{"product": "Shoes", "price": -100.0}`,
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "INVALID_PRICE" {
					t.Errorf("Expected status INVALID_PRICE, got %s", res.Status)
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
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "INTERNAL_SERVER_ERROR" {
					t.Errorf("Expected status INTERNAL_SERVER_ERROR, got %s", res.Status)
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

			h := handlers.NewCartHandler(mockSvc, discardLogger, cfg)

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
	cfg := config.ServerConfig{
		CtxDefaultTimeout: 5 * time.Second,
	}
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
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
				}
			},
		},
		{
			name:           "Fail - Invalid Item UUID",
			url:            "/api/v1/carts/" + cartID.String() + "/items/invalid-uuid",
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
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
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "CART_NOT_FOUND" {
					t.Errorf("Expected status CART_NOT_FOUND, got %s", res.Status)
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
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "ITEM_NOT_FOUND" {
					t.Errorf("Expected status ITEM_NOT_FOUND, got %s", res.Status)
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
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "INTERNAL_SERVER_ERROR" {
					t.Errorf("Expected status INTERNAL_SERVER_ERROR, got %s", res.Status)
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

			h := handlers.NewCartHandler(mockSvc, discardLogger, cfg)

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
	cfg := config.ServerConfig{
		CtxDefaultTimeout: 5 * time.Second,
	}
	cartID := uuid.New()

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
					Return(cartID, 6000.0, 0.1, 5400.0, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.CalculateDiscountResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode response JSON: %v", err)
				}
				if res.CartID != cartID || res.TotalPrice != 6000.0 || res.DiscountPercent != 0.1 || res.FinalPrice != 5400.0 {
					t.Errorf("Unexpected response data: %+v", res)
				}
			},
		},
		{
			name: "Success - Empty Cart (Zero Discount)",
			url:  "/api/v1/carts/" + cartID.String() + "/discount",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CalculateDiscount(gomock.Any(), cartID).
					Return(cartID, 0.0, 0.0, 0.0, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.CalculateDiscountResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode response JSON: %v", err)
				}
				if res.CartID != cartID || res.TotalPrice != 0.0 || res.DiscountPercent != 0.0 || res.FinalPrice != 0.0 {
					t.Errorf("Unexpected response data: %+v", res)
				}
			},
		},
		{
			name:           "Fail - Invalid Cart UUID",
			url:            "/api/v1/carts/invalid-uuid/discount",
			buildStubs:     func(ms *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "BAD_REQUEST" {
					t.Errorf("Expected status BAD_REQUEST, got %s", res.Status)
				}
			},
		},
		{
			name: "Fail - Cart Not Found",
			url:  "/api/v1/carts/" + cartID.String() + "/discount",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CalculateDiscount(gomock.Any(), cartID).
					Return(uuid.Nil, 0.0, 0.0, 0.0, errs.ErrCartNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "CART_NOT_FOUND" {
					t.Errorf("Expected status CART_NOT_FOUND, got %s", res.Status)
				}
			},
		},
		{
			name: "Fail - Database / Internal Error",
			url:  "/api/v1/carts/" + cartID.String() + "/discount",
			buildStubs: func(ms *mocks.MockService) {
				ms.EXPECT().
					CalculateDiscount(gomock.Any(), cartID).
					Return(uuid.Nil, 0.0, 0.0, 0.0, errors.New("db error")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body []byte) {
				var res handlers.ErrorResponse
				if err := json.Unmarshal(body, &res); err != nil {
					t.Fatalf("Failed to decode error response JSON: %v", err)
				}
				if res.Status != "INTERNAL_SERVER_ERROR" {
					t.Errorf("Expected status INTERNAL_SERVER_ERROR, got %s", res.Status)
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

			h := handlers.NewCartHandler(mockSvc, discardLogger, cfg)

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
