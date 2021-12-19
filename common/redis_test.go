package common

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRedisConn(t *testing.T) {
	redisPool := NewRedisPool()
	_, err := redisPool.Rdb.Ping().Result()
	if err != nil {
		t.Errorf("Cannot cannect to redis Pool, %v", err)
	}
}

func TestGetCurrentGoroutineId(t *testing.T) {

	goId, err := GetCurrentGoroutineId()
	fmt.Println(goId)
	if err != nil {
		t.Errorf("original error:%v", err)
	}
}

func TestUnLock(t *testing.T) {

	redisPool := NewRedisPool()
	var wg sync.WaitGroup
	var count int

	for n := 0; n < 5; n++ {
		wg.Add(1)

		go func() {
			redisPool.Lock()
			goId, _ := GetCurrentGoroutineId()
			fmt.Println("GoroutineId: ", goId)
			count++
			time.Sleep(15 * time.Second)
			result, _ := redisPool.Rdb.Get(redisPool.LockKey).Result()
			fmt.Println("GoroutineId: ", result)
			redisPool.UnLock()
			wg.Done()
		}()
	}
	wg.Wait()
	if count != 5 {
		t.Errorf("failed")
	}
}
