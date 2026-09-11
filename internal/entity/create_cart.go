package entity

type CreateCartResponse struct {
	ID    int        `json:"id"`
	Items []CartItem `json:"items"`
}
