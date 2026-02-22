package redisclient

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func New(addr string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}
