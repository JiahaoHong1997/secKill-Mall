package db

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var (
	rdb		*redis.Client
)

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6380",
		Password: "",
		PoolSize: 100,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}

func NewRedisConn() *redis.Client {
	return rdb
}


