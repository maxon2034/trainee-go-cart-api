package handlers

type ItemRequestDTO struct {
	Product string  `json:"product"`
	Price   float64 `json:"price"`
}
