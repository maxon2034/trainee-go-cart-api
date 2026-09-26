package service

type CartService struct {
	repo Repository
}

func New(rep Repository) *CartService { return &CartService{repo: rep} }
