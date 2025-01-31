package Repos

import "Demonstration-Service/internal/Application/Domain"

type IStorage interface {
	IsExist(id int) bool
	Read(id int) (Domain.Order, error)
	Save(order Domain.Order) error
}
