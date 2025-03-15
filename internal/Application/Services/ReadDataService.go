package Services

import (
	"Demonstration-Service/internal/Application/Abstractions/Repos"
	"Demonstration-Service/internal/Application/Domain"
	"fmt"
	"log"
)

type ReadDataService struct {
	cashRepo Repos.IStorage
	dbRepo   Repos.IStorage
}

func NewReadDataService(cashService Repos.CashStorage, dbRepo Repos.DBStorage) *ReadDataService {
	return &ReadDataService{cashRepo: cashService, dbRepo: dbRepo}
}

func (service *ReadDataService) GetById(id string) (Domain.Order, error) {
	order, err := service.cashRepo.Read(id)
	if err == nil {
		return order, nil
	} else {
		log.Println("order not found in cache", err)
	}

	order, err = service.dbRepo.Read(id)
	if err != nil {
		return Domain.Order{}, fmt.Errorf("order not found in db for id: %s, %w", id, err)
	}

	err = service.cashRepo.Save(order)
	if err != nil {
		log.Printf("failed to save order in cash for id: %s, %v", id, err)
	}

	return order, nil
}
