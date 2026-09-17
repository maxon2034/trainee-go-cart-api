package handlers

type CartDTO struct {
	ID    int           `json:"id"`
	Items []CartItemDTO `json:"items"`
}

type CartItemDTO struct {
	ID      int     `json:"id"`
	CartID  int     `json:"cart_id"`
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}
