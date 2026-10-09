package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/maxon2034/trainee-go-cart-api/internal/config"
)

type CartHandler struct {
	service Service
	logger  *slog.Logger
	cfg     config.ServerConfig
}

func NewCartHandler(s Service, l *slog.Logger, cfg config.ServerConfig) *CartHandler {
	return &CartHandler{s, l, cfg}
}

func (h *CartHandler) Create(w http.ResponseWriter, r *http.Request) {
	var createCartResponse CreateCartResponse

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxDefaultTimeout)
	defer cancel()

	cart, err := h.service.CreateCart(ctx)
	if err != nil {
		writeError(w, h.logger, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		h.logger.Error("error in creating cart", slog.Any("error", err))
		return
	}

	createCartResponse.ID = cart.ID
	createCartResponse.Items = make([]CreateCartItemResponse, 0)

	ok := writeResponse(w, h.logger, http.StatusCreated, createCartResponse)
	if !ok {
		return
	}
	h.logger.Info("created cart", slog.Any("cart", cart.ID))
}

func (h *CartHandler) View(w http.ResponseWriter, r *http.Request) {
	var viewCartResponse ViewCartResponse

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxDefaultTimeout)
	defer cancel()

	cartUUID, ok := parseUUID(w, r, h.logger, "cart_id")
	if !ok {
		return
	}

	cart, err := h.service.ViewCart(ctx, cartUUID)
	if err != nil {
		processError(w, h.logger, err)
		return
	}

	viewCartResponse.ID = cart.ID
	for _, item := range cart.Items {
		viewCartResponse.Items = append(viewCartResponse.Items, ViewCartItemResponse{
			ID:      item.ID,
			CartID:  item.CartID,
			Product: item.Product,
			Price:   item.Price,
		})
	}

	ok = writeResponse(w, h.logger, http.StatusOK, viewCartResponse)
	if !ok {
		return
	}
	h.logger.Info("viewed cart", slog.Any("cart", cart))
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {

	var addItemRequest AddItemRequest
	var addItemResponse AddItemResponse

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxDefaultTimeout)
	defer cancel()

	cartUUID, ok := parseUUID(w, r, h.logger, "cart_id")
	if !ok {
		return
	}

	if err := decodeBody(w, r, &addItemRequest); err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "BAD_REQUEST", "bad request")
		h.logger.Info("decode item request", slog.Any("error", err))
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			h.logger.Error("close body", slog.Any("error", err))
		}
	}()

	ok = validateItemRequest(w, h.logger, addItemRequest.Product, addItemRequest.Price)
	if !ok {
		return
	}

	cartItem, err := h.service.AddItem(ctx, cartUUID, addItemRequest.Product, addItemRequest.Price)
	if err != nil {
		processError(w, h.logger, err)
		return
	}

	addItemResponse.ID = cartItem.ID
	addItemResponse.CartID = cartItem.CartID
	addItemResponse.Product = addItemRequest.Product
	addItemResponse.Price = addItemRequest.Price

	ok = writeResponse(w, h.logger, http.StatusCreated, addItemResponse)
	if !ok {
		return
	}

	h.logger.Info("added item", slog.Any("cartItem", cartItem))
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var updateItemRequest UpdateItemRequest
	var updateItemResponse UpdateItemResponse

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxDefaultTimeout)
	defer cancel()

	cartUUID, ok := parseUUID(w, r, h.logger, "cart_id")
	if !ok {
		return
	}
	cartItemUUID, ok := parseUUID(w, r, h.logger, "item_id")
	if !ok {
		return
	}

	if err := decodeBody(w, r, &updateItemRequest); err != nil {
		writeError(w, h.logger, http.StatusBadRequest, "BAD_REQUEST", "bad request")
		h.logger.Info("decode item request", slog.Any("error", err))
		return
	}
	defer r.Body.Close()

	ok = validateItemRequest(w, h.logger, updateItemRequest.Product, updateItemRequest.Price)
	if !ok {
		return
	}

	cartItem, err := h.service.UpdateCartItem(ctx, cartUUID, cartItemUUID, updateItemRequest.Product, updateItemRequest.Price)
	if err != nil {
		processError(w, h.logger, err)
		return
	}

	if cartItem.Product != updateItemRequest.Product {
		h.logger.Error("product mismatch", slog.Any("expected product", updateItemRequest.Product), slog.Any("product", cartItem.Product))
		writeError(w, h.logger, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}
	if !cartItem.Price.Equal(updateItemRequest.Price) {
		h.logger.Error("price mismatch", slog.Any("expected price", updateItemRequest.Product), slog.Any("price", cartItem.Product))
		writeError(w, h.logger, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	updateItemResponse.ID = cartItem.ID
	updateItemResponse.CartID = cartItem.CartID
	updateItemResponse.Product = updateItemRequest.Product
	updateItemResponse.Price = updateItemRequest.Price

	ok = writeResponse(w, h.logger, http.StatusOK, updateItemResponse)
	if !ok {
		return
	}
	h.logger.Info("updated item", slog.Any("cartItem", cartItem))
}

func (h *CartHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxDefaultTimeout)
	defer cancel()

	cartUUID, ok := parseUUID(w, r, h.logger, "cart_id")
	if !ok {
		return
	}

	cartItemUUID, ok := parseUUID(w, r, h.logger, "item_id")
	if !ok {
		return
	}

	err := h.service.RemoveItem(ctx, cartUUID, cartItemUUID)
	if err != nil {
		processError(w, h.logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("deleted item", slog.Any("cartItem", cartItemUUID))
}

func (h *CartHandler) CalculateDiscount(w http.ResponseWriter, r *http.Request) {
	var calculateDiscountResponse CalculateDiscountResponse
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	cartUUID, ok := parseUUID(w, r, h.logger, "cart_id")
	if !ok {
		return
	}

	cartDiscount, err := h.service.CalculateDiscount(ctx, cartUUID)
	if err != nil {
		processError(w, h.logger, err)
		return
	}

	calculateDiscountResponse.CartID = cartDiscount.CartID
	calculateDiscountResponse.TotalPrice = cartDiscount.TotalPrice
	calculateDiscountResponse.DiscountPercent = cartDiscount.DiscountPercent
	calculateDiscountResponse.FinalPrice = cartDiscount.FinalPrice

	ok = writeResponse(w, h.logger, http.StatusOK, calculateDiscountResponse)
	if !ok {
		return
	}
	h.logger.Info("calculated discount", slog.Any("cart", calculateDiscountResponse))
}
