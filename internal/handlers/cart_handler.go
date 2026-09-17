package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CartHandler struct {
	service Service
}

func NewCartHandler(service Service) CartHandler {
	return CartHandler{service: service}
}

func (h *CartHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	cartResp, err := h.service.CreateCart(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(cartResp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *CartHandler) View(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	id := r.PathValue("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("invalid id", id, idInt)
		return
	}

	cartResp, err := h.service.ViewCart(ctx, idInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(cartResp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
