package OrdersServices

import "Demonstration-Service/internal/Application/Domain"

type IPostService interface {
	AddOrder(order Domain.Order) error
}
