package run

import (
	"Demonstration-Service/internal/dependencyInjection"
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

func Run() {
	ctx := context.Background()

	app, err := dependencyInjection.InitializeApplication(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		return
	}
	defer func() {
		if err := app.Logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	app.Logger.Info("Starting application...")

	var wg sync.WaitGroup

	// Запуск gRPC сервера асинхронно
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("Starting gRPC server")
		if err := app.GrpcServer.Start(":50051"); err != nil {
			app.Logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Запуск HTTP сервера асинхронно
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("Starting HTTP server")
		if err := app.HttpServer.Start(":8080"); err != nil {
			app.Logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Запуск Kafka consumer асинхронно
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("Starting Kafka consumer")
		if err := app.KafkaConsumer.Listen(ctx); err != nil {
			app.Logger.Error("Failed to start consume messages", zap.Error(err))
		}
	}()

	app.Logger.Info("All systems are up and running")
	wg.Wait()
	app.Logger.Info("Shutting down...")

	app.GrpcServer.Stop()
	app.HttpServer.Stop()
}