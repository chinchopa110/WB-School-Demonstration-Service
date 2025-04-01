package Services

import (
	"Demonstration-Service/internal/Application/Abstractions/Repos"
	"Demonstration-Service/internal/Application/Domain"
	"Demonstration-Service/internal/Application/Services/validator"
	"context"
	"errors"
	"fmt"
	"sync"
)

type ProcessDataService struct {
	cashRepo Repos.IStorage
	dbRepo   Repos.IStorage
}

func NewProcessDataService(cashService Repos.CashStorage, dbRepo Repos.DBStorage) *ProcessDataService {
	return &ProcessDataService{cashRepo: cashService, dbRepo: dbRepo}
}

func (service *ProcessDataService) AddOrder(order Domain.Order, ctx context.Context) error {
	if err := validator.ValidateOrder(order); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)
	defer close(errChan)

	go func() {
		defer wg.Done()
		err := service.cashRepo.Save(order, ctx)
		if err != nil {
			errChan <- fmt.Errorf("cash repo error: %w", err)
			return
		}
		errChan <- nil
	}()

	go func() {
		defer wg.Done()
		err := service.dbRepo.Save(order, ctx)
		if err != nil {
			errChan <- fmt.Errorf("db repo error: %w", err)
			return
		}
		errChan <- nil
	}()
	wg.Wait()

	var allErrors error
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			allErrors = errors.Join(allErrors, err)
		}
	}

	if allErrors != nil {
		return fmt.Errorf("errors while saving order: %w", allErrors)
	}

	return nil
}
