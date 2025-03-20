//go:build wireinject
// +build wireinject

package dependencyInjection

import (
	configs2 "Demonstration-Service/internal/run/configs"
	grpcConfig2 "Demonstration-Service/internal/run/configs/grpcConfig"
	httpConfig2 "Demonstration-Service/internal/run/configs/httpConfig"
	"context"
	"database/sql"
	"time"

	"Demonstration-Service/internal/Application/Abstractions/Repos"
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Application/Services"
	"Demonstration-Service/internal/Infrastructure/kafka"
	"Demonstration-Service/internal/Infrastructure/post"
	"Demonstration-Service/internal/Infrastructure/storages/dataAccess"
	"Demonstration-Service/internal/Infrastructure/storages/redisCache"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func provideLogFilePath() string {
	return "logs/app.log"
}

func provideRedisConfig() configs2.RedisConfig {
	return *configs2.NewRedisConfig()
}

func provideContext() context.Context {
	return context.Background()
}

func provideCacheExpiration() time.Duration {
	return 30 * time.Minute
}

func InitializeApplication(ctx context.Context) (*Application, error) {
	wire.Build(
		provideLogFilePath,
		provideRedisConfig,
		provideCacheExpiration,

		// Логгер
		configs2.InitLogger,

		// База данных
		configs2.GetUpSQL,

		// Redis
		configs2.NewClient,

		// Репозитории
		dataAccess.NewOrdersRepo,
		redisCache.NewRedisRepository,

		// Привязка интерфейсов
		wire.Bind(new(Repos.CashStorage), new(*redisCache.RedisRepository)),
		wire.Bind(new(Repos.DBStorage), new(*dataAccess.OrdersRepo)),
		wire.Bind(new(OrdersServices.IGetService), new(*Services.ReadDataService)),
		wire.Bind(new(OrdersServices.IPostService), new(*Services.ProcessDataService)),

		// Сервисы
		Services.NewReadDataService,
		Services.NewProcessDataService,

		// gRPC и HTTP серверы
		grpcConfig2.ServerGetUp,
		httpConfig2.ServerGetUp,

		// Kafka consumer
		configs2.NewKafkaConfig,
		kafka.NewKafkaConsumer,
		post.NewProcessService,

		// Сборка приложения
		NewApplication,
	)
	return &Application{}, nil
}

type Application struct {
	Logger        *zap.Logger
	db            *sql.DB
	redisClient   *redis.Client
	readService   *Services.ReadDataService
	addService    *Services.ProcessDataService
	GrpcServer    *grpcConfig2.Server
	HttpServer    *httpConfig2.Server
	KafkaConsumer *kafka.Consumer
}

func NewApplication(
	Logger *zap.Logger,
	db *sql.DB,
	redisClient *redis.Client,
	readService *Services.ReadDataService,
	addService *Services.ProcessDataService,
	GrpcServer *grpcConfig2.Server,
	HttpServer *httpConfig2.Server,
	KafkaConsumer *kafka.Consumer,
) *Application {
	return &Application{
		Logger:        Logger,
		db:            db,
		redisClient:   redisClient,
		readService:   readService,
		addService:    addService,
		GrpcServer:    GrpcServer,
		HttpServer:    HttpServer,
		KafkaConsumer: KafkaConsumer,
	}
}
