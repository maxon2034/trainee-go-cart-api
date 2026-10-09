package config

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := Load("../../config")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Cart.DiscountTotalPrice.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected discount_total_price 5000, got %s", cfg.Cart.DiscountTotalPrice)
	}
}
