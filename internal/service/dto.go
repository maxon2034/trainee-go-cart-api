package service

import (
	"github.com/google/uuid"
	"github.com/maxon2034/trainee-go-cart-api/internal/entity"
)

type CartDTO struct {
	ID    uuid.UUID     `json:"id"`
	Items []CartItemDTO `json:"items"`
}

type CartItemDTO struct {
	ID      uuid.UUID `json:"id"`
	CartID  uuid.UUID `json:"cart_id"`
	Product string    `json:"product"`
	Price   float64   `json:"price"`
}

func ToDTO(cart *entity.Cart) (cartDTO CartDTO) {
	cartDTO.ID = cart.ID
	cartDTO.Items = make([]CartItemDTO, len(cart.Items))
	for i := range cart.Items {
		cartDTO.Items[i].ID = cart.Items[i].ID
		cartDTO.Items[i].CartID = cart.ID
		cartDTO.Items[i].Price = cart.Items[i].Price
		cartDTO.Items[i].Product = cart.Items[i].Product
	}
	return cartDTO
}
