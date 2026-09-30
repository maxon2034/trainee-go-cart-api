package entity

import (
	"github.com/google/uuid"
)

type CartItem struct {
	ID      *uuid.UUID `db:"id"`
	CartID  uuid.UUID  `db:"cart_id"`
	Product *string    `db:"product"`
	Price   *float64   `db:"price"`
}
