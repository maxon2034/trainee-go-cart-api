package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/errs"
	"github.com/shopspring/decimal"
)

func writeError(w http.ResponseWriter, l *slog.Logger, code int, status string, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(ErrorResponse{Status: status, Message: message})
	if err != nil {
		l.Error("encode error response", slog.Any("error", err))
	}
}

func processError(w http.ResponseWriter, l *slog.Logger, err error) {
	switch {
	case errors.Is(err, errs.ErrCartNotFound):
		writeError(w, l, http.StatusNotFound, "CART_NOT_FOUND", "cart not found")
		l.Info("not found", slog.Any("error", "cart not found"))
	case errors.Is(err, errs.ErrFullCart):
		writeError(w, l, http.StatusBadRequest, "FULL_CART", "full cart")
		l.Info("bad request", slog.Any("error", "full cart"))
	case errors.Is(err, errs.ErrCartItemNotFound):
		writeError(w, l, http.StatusNotFound, "ITEM_NOT_FOUND", "cart item not found")
		l.Info("not found", slog.Any("error", "cart item not found"))
	default:
		writeError(w, l, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
	}
}

func writeResponse(w http.ResponseWriter, l *slog.Logger, code int, payload interface{}) bool {
	resp, err := json.Marshal(payload)
	if err != nil {
		writeError(w, l, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		l.Error("encode response", slog.Any("error", err))
		return false
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(resp)
	if err != nil {
		l.Error("write response", slog.Any("error", err))
		return false
	}
	return true
}

func parseUUID(w http.ResponseWriter, r *http.Request, l *slog.Logger, pathValue string) (uuid.UUID, bool) {
	ID := r.PathValue(pathValue)
	UUID, err := uuid.Parse(ID)
	if err != nil {
		writeError(w, l, http.StatusBadRequest, "BAD_REQUEST", "invalid UUID: "+ID)
		l.Info("invalid uuid", slog.Any("path value", pathValue), slog.Any("uuid", ID))
		return uuid.Nil, false
	}
	return UUID, true
}

func validateItemRequest(w http.ResponseWriter, l *slog.Logger, product string, price decimal.Decimal) bool {
	product = strings.TrimSpace(product)
	if product == "" {
		writeError(w, l, http.StatusBadRequest, "EMPTY_PRODUCT", "empty product")
		l.Info("bad request", slog.Any("error", "empty product"))
		return false
	}
	if price.LessThanOrEqual(decimal.Zero) {
		writeError(w, l, http.StatusBadRequest, "INVALID_PRICE", "price must be greater than zero")
		l.Info("bad request", slog.Any("error", "invalid price"))
		return false
	}
	return true
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 200)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	return dec.Decode(v)
}
