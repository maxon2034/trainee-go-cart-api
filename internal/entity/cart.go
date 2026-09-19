package entity

import "github.com/google/uuid"

type Cart struct {
	ID    uuid.UUID
	Items []CartItem
}
