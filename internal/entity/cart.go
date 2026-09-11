package entity

import (
	"time"
)

type Cart struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}
