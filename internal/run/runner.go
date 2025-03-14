package run

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"Demonstration-Service/internal/Application/Services"
	"Demonstration-Service/internal/Infrastructure/kafka"
	"Demonstration-Service/internal/Infrastructure/post"
	"Demonstration-Service/internal/Infrastructure/storages/dataAccess"
	"Demonstration-Service/internal/Infrastructure/storages/redisCache"
	"Demonstration-Service/internal/configs"
	"Demonstration-Service/internal/configs/grpcConfig"
	"Demonstration-Service/internal/configs/httpConfig"
)

func Run() {
	//TODO: переписать с google/wire
	

	var wg sync.WaitGroup
	ctx := context.Background()

	logger, err := configs.InitLogger("logs/app.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	logger.Info("Starting application...")

	// 1. Инициализация и запуск базы данных
	db, err := configs.GetUpSQL()
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Could not close database connection", zap.Error(err))
		}
	}()
	sqlRepo := dataAccess.NewOrdersRepo(db)
	logger.Info("Database connected")

	// 2. Инициализация и запуск Redis
	redisCfg := configs.NewRedisConfig()
	redisClient, err := configs.NewClient(ctx, *redisCfg)
	if err != nil {
		logger.Fatal("Failed to create Redis client", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Could not close redis connection", zap.Error(err))
		}
	}()
	redisRepo := redisCache.NewRedisRepository(redisClient, 30*time.Second)
	logger.Info("Redis connected")

	// 3. Инициализация сервиса данных
	readService := Services.NewReadDataService(redisRepo, sqlRepo)
	addService := Services.NewProcessDataService(redisRepo, sqlRepo)
	logger.Info("Services initialized")

	// 4. Запуск gRPC сервера асинхронно
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting gRPC server")
		grpcConfig.ServerGetUp(readService)

	}()

	// 5. Запуск HTTP сервера асинхронно
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server")
		httpConfig.ServerGetUp(readService)
	}()

	// 6. Запуск Kafka consumer асинхронно
	kc := configs.NewKafkaConfig()
	consumer := kafka.NewKafkaConsumer(kc, post.NewProcessService(addService))
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting Kafka consumer")
		if err := consumer.Listen(ctx); err != nil {
			logger.Error("Failed to start consume messages", zap.Error(err))
		}
	}()

	// 7. Ожидание завершения всех горутин
	logger.Info("All systems are up and running")
	wg.Wait()
	logger.Info("Shutting down...")
}
