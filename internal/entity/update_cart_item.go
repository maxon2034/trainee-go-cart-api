package entity

type UpdateCartItemRequest struct {
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}

type UpdateCartItemResponse struct {
	ID      int     `json:"id"`
	CartID  int     `json:"cart_id"`
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}
