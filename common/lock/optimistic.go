package lock
import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
)

var (
	rdb_op *redis.Client
	// maxRetries represents the maximum number of retry attempts after each client transaction is interrupted.
	// When this value is too small, the inventory may not be emptied
	maxRetries = 50
	// clientNum represents the number of client initiating a shopping request, this value is for test
	clientNum = 10000
	// repositories is the number of commodity
	repositories = 200
)

type RedisPoolOp struct {
	Rdb      *redis.Client
	Ctx      context.Context
}

// 初始化连接
func init() {
	rdb_op = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
		PoolSize: 100,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}

func NewRedisPoolOp() *RedisPoolOp {
	return &RedisPoolOp{
		Rdb: rdb_op,
		Ctx: context.Background(),
	}
}

func main() {
	store := NewRedisPoolOp()
	rdb := store.Rdb
	ctx := store.Ctx

	// Increment transactionally increments key using GET and SET commands.
	increment := func(key string) error {
		// Transactional function.
		txf := func(tx *redis.Tx) error {
			// Get current value or zero.
			n, err := tx.Get(ctx, key).Int()
			if err != nil && err != redis.Nil {
				return err
			}

			// Actual opperation (local in optimistic lock).
			n++
			if n > repositories {
				return errors.New("more than repositories limit")
			}
			// Operation is committed only if the watched keys remain unchanged.
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, key, n, 0)
				return nil
			})
			return err
		}

		for i := 0; i < maxRetries; i++ {
			err := rdb.Watch(ctx, txf, key)
			if err == nil {
				// Success.
				return nil
			}
			if err == redis.TxFailedErr {
				// Optimistic lock lost. Retry.
				continue
			}
			// Return any other error.
			return err
		}

		return errors.New("increment reached maximum number of retries")
	}

	var wg sync.WaitGroup


	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	for i := 0; i < clientNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := increment("counter3"); err != nil {
				fmt.Println("increment error:", err)
			}
		}()
	}
	wg.Wait()

	n, err := rdb.Get(ctx,"counter3").Int()
	fmt.Println("ended with", n, err)
}
