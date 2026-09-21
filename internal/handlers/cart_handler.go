package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	errs "github.com/maxon2034/trainee-go-cart-api/internal/error"
	"github.com/maxon2034/trainee-go-cart-api/internal/repository"
)

var errResp errs.ErrorResponse

type CartHandler struct {
	service Service
	logger  *slog.Logger
}

func NewCartHandler(s Service, l *slog.Logger) *CartHandler {
	return &CartHandler{s, l}
}

func (h *CartHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	cart, err := h.service.CreateCart(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp.InternalServerError())
		h.logger.Error("error in creating cart", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(cart); err != nil {
		h.logger.Error("error in forming response", slog.Any("error", err))
		return
	}
	h.logger.Info("created cart", slog.Any("cart", cart))
}

func (h *CartHandler) View(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.Error("Invalid method")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	id := r.PathValue("id")

	uuid, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errResp.BadRequest())
		h.logger.Error("error in parsing uuid", slog.Any("error", err))
		return
	}

	cart, err := h.service.ViewCart(ctx, uuid)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write(errResp.NotFound())
			h.logger.Info("cart not found", slog.Any("id", uuid.String()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errResp.InternalServerError())
		h.logger.Error("error in viewing cart", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(cart); err != nil {
		h.logger.Error("error in forming response", slog.Any("error", err))
		return
	}
	h.logger.Info("viewed cart", slog.Any("cart", cart))
}
