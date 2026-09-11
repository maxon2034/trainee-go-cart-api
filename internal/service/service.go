package service

type CartService struct {
	repo Repository
}

func NewService(rep Repository) *CartService { return &CartService{repo: rep} }
