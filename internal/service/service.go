package service

type Service struct {
	repo Repository
}

func NewService(rep Repository) *Service { return &Service{repo: rep} }
