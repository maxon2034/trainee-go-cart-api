package entity

type AddCartItemRequest struct {
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}

type AddCartItemResponse struct {
	ID      int     `json:"id"`
	CartID  int     `json:"cart_id"`
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}
