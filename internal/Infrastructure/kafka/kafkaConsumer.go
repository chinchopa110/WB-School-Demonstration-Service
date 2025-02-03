package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"time"

	"Demonstration-Service/internal/Application/Domain"
	"Demonstration-Service/internal/configs"
)

type Consumer struct {
	config      *configs.KafkaConfig
	messageChan chan Domain.Order
	reader      *kafka.Reader
	logger      *zap.Logger
}

func NewKafkaConsumer(config *configs.KafkaConfig, msgChan chan Domain.Order, logger *zap.Logger) *Consumer {
	return &Consumer{
		config:      config,
		messageChan: msgChan,
		logger:      logger,
	}
}

func (kc *Consumer) Listen(ctx context.Context) error {
	var err error
	kc.reader, err = kc.config.InitConsumer()
	if err != nil {
		return fmt.Errorf("failed to init kafka consumer: %w", err)
	}

	defer func() {
		if err := kc.config.CloseConsumer(); err != nil {
			kc.logger.Error("failed to close kafka reader", zap.Error(err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			message, err := kc.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
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
			kc.messageChan <- order

			if err := kc.reader.CommitMessages(ctx, message); err != nil {
				kc.logger.Error("failed to commit kafka message", zap.Error(err))
			}
		}
	}
}
