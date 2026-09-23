package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
)

type CartHandler struct {
	service Service
	logger  *slog.Logger
}

func NewCartHandler(s Service, l *slog.Logger) *CartHandler {
	return &CartHandler{s, l}
}

func (h *CartHandler) Create(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	cart, err := h.service.CreateCart(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errs.InternalServerError())
		h.logger.Error("errs in creating cart", slog.Any("errs", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(cart); err != nil {
		h.logger.Error("errs in forming response", slog.Any("errs", err))
		return
	}
	h.logger.Info("created cart", slog.Any("cart", cart))
}

func (h *CartHandler) View(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	id := r.PathValue("id")

	cartID, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errs.BadCartRequest())
		h.logger.Error("errs in parsing uuid", slog.Any("errs", err))
		return
	}

	cart, err := h.service.ViewCart(ctx, cartID)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write(errs.NotFound())
			h.logger.Info("cart not found", slog.Any("id", cartID.String()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errs.InternalServerError())
		h.logger.Error("errs in viewing cart", slog.Any("errs", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(cart); err != nil {
		h.logger.Error("errs in forming response", slog.Any("errs", err))
		return
	}
	h.logger.Info("viewed cart", slog.Any("cart", cart))
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {

	var itemRequest ItemRequestDTO

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	cartId := r.PathValue("cart_id")
	cartUUID, err := uuid.Parse(cartId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errs.BadCartRequest())
		h.logger.Error("error in parsing uuid", slog.Any("error", err))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&itemRequest); err != nil {
		h.logger.Error("error in parsing request", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errs.BadRequest())
		return
	}
	defer r.Body.Close()

	cartItem, err := h.service.AddItem(ctx, cartUUID, itemRequest.Product, itemRequest.Price)
	if err != nil {
		if errors.Is(err, errs.ErrCartNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write(errs.NotFound())
			h.logger.Info("cart not found", slog.Any("id", cartUUID.String()))
			return
		}
		if errors.Is(err, errs.ErrFullCart) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errs.FullCart())
			h.logger.Info("cart full", slog.Any("id", cartUUID.String()))
			return
		}
		if errors.Is(err, errs.ErrNegativePrice) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errs.NegativePrice())
			return
		}

		if errors.Is(err, errs.ErrEmptyProduct) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errs.EmptyProduct())
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errs.InternalServerError())
		h.logger.Error("error in adding item", slog.Any("error", err))
		return
	}
	fmt.Println(cartItem)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(cartItem); err != nil {
		h.logger.Error("error in forming response", slog.Any("error", err))
		return
	}
	h.logger.Info("added item", slog.Any("cartItem", cartItem))
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var itemRequest ItemRequestDTO

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	cartId := r.PathValue("cart_id")
	cartUUID, err := uuid.Parse(cartId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errs.BadCartRequest())
		h.logger.Error("error in parsing cart uuid", slog.Any("error", err))
		return
	}

	cartItemID := r.PathValue("id")
	cartItemUUID, err := uuid.Parse(cartItemID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errs.BadItemRequest())
		h.logger.Error("error in parsing item uuid", slog.Any("error", err))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&itemRequest); err != nil {
		h.logger.Error("error in parsing request", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errs.BadRequest())
		return
	}
	defer r.Body.Close()

	cartItemDTO, err := h.service.UpdateCartItem(ctx, cartItemUUID, itemRequest.Product, itemRequest.Price)
	if err != nil {
		if errors.Is(err, errs.ErrCartItemNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write(errs.NotFound())
			h.logger.Info("cart not found", slog.Any("id", cartUUID.String()))
			return
		}
		if errors.Is(err, errs.ErrEmptyProduct) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errs.EmptyProduct())
			return
		}
		if errors.Is(err, errs.ErrNegativePrice) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errs.NegativePrice())
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("error in adding item", slog.Any("error", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(cartItemDTO); err != nil {
		h.logger.Error("error in forming response", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errs.InternalServerError())
		return
	}
	h.logger.Info("updated item", slog.Any("cartItem", cartItemDTO))
}
