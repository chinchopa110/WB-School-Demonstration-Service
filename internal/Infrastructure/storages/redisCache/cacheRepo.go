package redisCache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"Demonstration-Service/internal/Application/Domain"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound = errors.New("order not found in cache")
)

type RedisRepository struct {
	client     *redis.Client
	expiration time.Duration
}

func NewRedisRepository(client *redis.Client, expiration time.Duration) *RedisRepository {
	return &RedisRepository{
		client:     client,
		expiration: expiration,
	}
}

func (r *RedisRepository) IsExist(id string) bool {
	ctx := context.Background()

	exists, err := r.client.Exists(ctx, id).Result()
	if err != nil {
		return false
	}
	return exists > 0
}

func (r *RedisRepository) Read(id string) (Domain.Order, error) {
	ctx := context.Background()

	val, err := r.client.Get(ctx, id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Domain.Order{}, ErrNotFound
		}
		return Domain.Order{}, fmt.Errorf("failed to get order from redis: %w", err)
	}

	var order Domain.Order
	err = json.Unmarshal([]byte(val), &order)
	if err != nil {
		return Domain.Order{}, fmt.Errorf("failed to unmarshal order from redis: %w", err)
	}

	return order, nil
}

func (r *RedisRepository) Save(order Domain.Order) error {
	ctx := context.Background()

	orderBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = r.client.Set(ctx, order.OrderUID, orderBytes, r.expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set order to redis: %w", err)
	}

	return nil
}
