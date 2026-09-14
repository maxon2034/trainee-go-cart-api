package handlers

import "github.com/maxon2034/trainee-go-cart-api/internal/entity"

type Service interface {
	CreateCart() (entity.CreateCartResponse, error)
	ViewCart()
	AddItem()
	UpdateItem()
	RemoveItem()
	CalculatePrice()
}
