package OrdersServices

import (
	"Demonstration-Service/internal/Application/Domain"
	"context"
)

type IPostService interface {
	AddOrder(order Domain.Order, ctx context.Context) error
}
