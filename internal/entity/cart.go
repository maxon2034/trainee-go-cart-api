package entity

import (
	"time"
)

type Cart struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

type CartDTO struct {
	ID    int           `json:"id"`
	Items []CartItemDTO `json:"items"`
}
