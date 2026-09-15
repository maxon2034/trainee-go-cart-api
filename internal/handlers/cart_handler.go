package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	id := r.URL.Query().Get("id")

	cartResp, err := h.service.ViewCart(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(cartResp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
