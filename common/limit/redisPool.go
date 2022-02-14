package limit

import (
	"context"
	"github.com/go-redis/redis/v8"
)


var (
	rdb	*redis.Client
)

type RedisPool struct {
	Rdb      *redis.Client
	Ctx      context.Context
}

func init() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6378",
		Password: "",
		PoolSize: 100,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}

func NewRedisConn() *RedisPool {
	return &RedisPool{
		Rdb: rdb,
		Ctx: context.Background(),
	}
}
