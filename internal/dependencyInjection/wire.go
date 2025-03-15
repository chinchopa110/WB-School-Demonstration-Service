//+build wireinject

package wireInject

import (
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
	"Demonstration-Service/internal/configs"
	"Demonstration-Service/internal/configs/grpcConfig"
	"Demonstration-Service/internal/configs/httpConfig"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func provideLogFilePath() string {
	return "logs/app.log"
}

func provideRedisConfig() configs.RedisConfig {
	return *configs.NewRedisConfig()
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
		configs.InitLogger,

		// База данных
		configs.GetUpSQL,

		// Redis
		configs.NewClient,

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
		grpcConfig.ServerGetUp,
		httpConfig.ServerGetUp,

		// Kafka consumer
		configs.NewKafkaConfig,
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
	GrpcServer    *grpcConfig.Server
	HttpServer    *httpConfig.Server
	KafkaConsumer *kafka.Consumer
}

func NewApplication(
	Logger *zap.Logger,
	db *sql.DB,
	redisClient *redis.Client,
	readService *Services.ReadDataService,
	addService *Services.ProcessDataService,
	GrpcServer *grpcConfig.Server,
	HttpServer *httpConfig.Server,
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