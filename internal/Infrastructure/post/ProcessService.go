package post

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Application/Domain"
	"context"
)

type ProcessService struct {
	service OrdersServices.IPostService
}

func NewProcessService(service OrdersServices.IPostService) *ProcessService {
	return &ProcessService{service: service}
}

func (ps *ProcessService) ProcessMessage(order Domain.Order, ctx context.Context) error {
	return ps.service.AddOrder(order, ctx)
}
