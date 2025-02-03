package OrdersServices

import "Demonstration-Service/internal/Application/Domain"

type IGetService interface {
	GetById(id string) (Domain.Order, error)
}
