package redis

import (
	"fmt"
	"order_service/internal/config"

	"github.com/redis/go-redis/v9"
)

func New(conf config.RedisConfig) (*redis.Client, error) {
	opts, err := redis.ParseURL(conf.ConnString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse reddis connString: %v", err)
	}
	client := redis.NewClient(opts)
	return client, nil
}
