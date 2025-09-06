package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func New(connString string) (*redis.Client, error) {
	opts, err := redis.ParseURL(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse reddis connString: %v", err)
	}
	client := redis.NewClient(opts)
	return client, nil
}
