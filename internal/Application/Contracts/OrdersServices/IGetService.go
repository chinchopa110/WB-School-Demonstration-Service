package OrdersServices

import "Demonstration-Service/internal/Application/Domain"

type IGetService interface {
	GetById(id int) (Domain.Order, error)
}

//TODO: будет выходить в слой презентации
