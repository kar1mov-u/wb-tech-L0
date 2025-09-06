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

func (c *OrderCacheRedis) Set(ctx context.Context, id string, order models.Order) error {
	err := c.client.Set(ctx, id, order, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}
