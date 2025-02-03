package Repos

import "Demonstration-Service/internal/Application/Domain"

type IStorage interface {
	IsExist(id string) bool
	Read(id string) (Domain.Order, error)
	Save(order Domain.Order) error
}
