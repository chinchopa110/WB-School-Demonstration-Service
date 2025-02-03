package configs

import (
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topic   string
	Reader  *kafka.Reader
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "Orders",
		GroupID: "my-consumer-group",
	}
}

func (c *KafkaConfig) InitConsumer() (*kafka.Reader, error) {
	config := kafka.ReaderConfig{
		Brokers:  c.Brokers,
		GroupID:  c.GroupID,
		Topic:    c.Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
	}

	reader := kafka.NewReader(config)

	c.Reader = reader
	return reader, nil
}

func (c *KafkaConfig) CloseConsumer() error {
	if c.Reader != nil {
		if err := c.Reader.Close(); err != nil {
			return fmt.Errorf("failed to close kafka reader: %w", err)
		}
	}
	return nil
}
