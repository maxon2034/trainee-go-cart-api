package repository

import (
	"github.com/jmoiron/sqlx"
)

type CartRepository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *CartRepository {
	return &CartRepository{DB: db}
}
