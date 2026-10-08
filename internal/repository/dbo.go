package repository

import (
	"time"

	"github.com/google/uuid"
)

type CartDBO struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

type CartItemDBO struct {
	ID      uuid.UUID `db:"id"`
	CartID  uuid.UUID `db:"cart_id"`
	Product string    `db:"product"`
	Price   float64   `db:"price"`
}
