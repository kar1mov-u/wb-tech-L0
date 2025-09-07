package cache

import (
	"context"
	"order_service/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderCacheRedis struct {
	client *redis.Client
}

func NewOrderCacheRedis(client *redis.Client) *OrderCacheRedis {
	return &OrderCacheRedis{client: client}
}

func (c *OrderCacheRedis) Set(ctx context.Context, id string, order models.Order) error {
	err := c.client.Set(ctx, id, order, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *OrderCacheRedis) Get(ctx context.Context, id string) (models.Order, bool, error) {
	var order models.Order
	err := c.client.Get(ctx, id).Scan(&order)
	if err != nil {
		return models.Order{}, false, err
	}
	return order, true, nil
}
