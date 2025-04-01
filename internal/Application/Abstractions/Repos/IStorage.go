package Repos

import (
	"Demonstration-Service/internal/Application/Domain"
	"context"
)

type IStorage interface {
	IsExist(id string) bool
	Read(id string) (Domain.Order, error)
	Save(order Domain.Order, ctx context.Context) error
}
