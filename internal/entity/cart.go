package entity

import "github.com/google/uuid"

type Cart struct {
	ID    uuid.UUID  `db:"id"`
	Items []CartItem `db:"items"`
}
