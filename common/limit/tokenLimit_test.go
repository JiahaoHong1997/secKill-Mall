package limit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)


func TestTokenLimit_Take(t *testing.T) {
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
	store := &RedisPool{
		Rdb: rdb,
		Ctx: ctx,
	}

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenLimiter(rate, burst, store, "tokenlimit")
	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if l.Allow() {
			allowed++
		}
	}
	fmt.Println(allowed)
	if allowed != burst + rate {
		t.Errorf("allowed:%v, burst:%v, rate;%v", allowed, burst, rate)
	}
}

func TestTokenLimit_TakeBurst(t *testing.T) {
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
	store := &RedisPool{
		Rdb: rdb,
		Ctx: ctx,
	}

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenLimiter(rate, burst, store, "tokenlimit")
	var allowed int
	for i := 0; i < total; i++ {
		if l.Allow() {
			allowed++
		}
	}
	fmt.Println(allowed)
	if allowed != burst {
		t.Errorf("allowed:%v, burst:%v, rate;%v", allowed, burst, rate)
	}
}
