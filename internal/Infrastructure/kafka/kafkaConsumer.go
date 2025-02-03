package kafka

import (
	"Demonstration-Service/internal/Infrastructure/post"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"Demonstration-Service/internal/Application/Domain"
	"Demonstration-Service/internal/configs"
)

type Consumer struct {
	config  *configs.KafkaConfig
	reader  *kafka.Reader
	logger  *zap.Logger
	service *post.ProcessService
}

func NewKafkaConsumer(config *configs.KafkaConfig, service *post.ProcessService) *Consumer {
	logger, err := configs.InitLogger("logs/kafka.log")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	return &Consumer{
		config:  config,
		logger:  logger,
		service: service,
	}
}

func (kc *Consumer) Listen(ctx context.Context) error {
	kc.logger.Info("Starting Kafka consumer...")

	var err error
	kc.reader, err = kc.config.InitConsumer()
	if err != nil {
		kc.logger.Error("failed to init kafka consumer", zap.Error(err))
		return fmt.Errorf("failed to init kafka consumer: %w", err)
	}

	kc.logger.Info("Successfully connected to Kafka broker.")

	defer func() {
		if err := kc.config.CloseConsumer(); err != nil {
			kc.logger.Error("failed to close kafka reader", zap.Error(err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			kc.logger.Info("Kafka consumer context done.")
			return nil
		default:
			message, err := kc.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					kc.logger.Info("Kafka consumer context canceled or deadline exceeded.")
					return nil
				}
				kc.logger.Error("failed to fetch kafka message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			var order Domain.Order
			if err := json.Unmarshal(message.Value, &order); err != nil {
				kc.logger.Error("failed to unmarshal kafka message", zap.Error(err))
				continue
			}

			err = kc.service.ProcessMessage(order)
			if err != nil {
				kc.logger.Error("failed to process message", zap.Error(err), zap.Any("order", order))
			}

			if err := kc.reader.CommitMessages(ctx, message); err != nil {
				kc.logger.Error("failed to commit kafka message", zap.Error(err))
			}
		}
	}
}
