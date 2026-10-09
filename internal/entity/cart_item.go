package entity

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CartItem struct {
	ID      uuid.UUID       `db:"id"`
	CartID  uuid.UUID       `db:"cart_id"`
	Product string          `db:"product"`
	Price   decimal.Decimal `db:"price"`
}
