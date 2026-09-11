package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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
	cartResp, err := h.service.CreateCart(context.Background())
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
