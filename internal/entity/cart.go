package entity

import (
	"time"
)

type Cart struct {
	ID        int       `db:"id"`
	createdAt time.Time `db:"created_at"`
}
