package test

import (
	"fmt"
	"seckill/common/lock"
	"seckill/dao/db"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRedisConn(t *testing.T) {
	redisPool := lock.NewRedisPool()
	_, err := redisPool.Rdb.Ping(redisPool.Ctx).Result()
	if err != nil {
		t.Errorf("Cannot cannect to redis Pool, %v", err)
	}
}

func TestGetCurrentGoroutineId(t *testing.T) {

	goId, err := lock.GetCurrentGoroutineId()
	fmt.Println(goId)
	if err != nil {
		t.Errorf("original error:%v", err)
	}
}

func TestUnLock(t *testing.T) {

	redisPool := lock.NewRedisPool()
	var wg sync.WaitGroup
	var count int

	for n := 0; n < 5; n++ {
		wg.Add(1)

		go func() {
			redisPool.Lock()
			goId, _ := lock.GetCurrentGoroutineId()
			fmt.Println("GoroutineId: ", goId)
			count++
			time.Sleep(15 * time.Second)
			result, _ := redisPool.Rdb.Get(redisPool.Ctx, redisPool.LockKey).Result()
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

func T(a int) {}

func BenchmarkConcurrent(b *testing.B) {

	redisPool := lock.NewRedisPool()
	var v atomic.Value
	var count int32
	type Map map[int]interface{}

	m := make(Map)
	for i := 0; i < 10; i++ {
		m[i] = i
	}
	v.Store(m)

	go func() {
		for {
			redisPool.Lock()

			m1 := v.Load().(Map)
			m2 := make(Map)

			for k, v := range m1 {
				m2[k] = v.(int) + 1
			}
			v.Store(m2)
		}
	}()

	var wg sync.WaitGroup
	for n := 0; n < 2000; n++ {
		wg.Add(1)
		go func() {
			for n := 0; n < b.N; n++ {
				T(v.Load().(Map)[1].(int))
				atomic.AddInt32(&count,1)
			}
			wg.Done()
		}()

	}
	wg.Wait()
	b.Logf("count:%v", count)
}

func BenchmarkAtomic(b *testing.B) {
	var v atomic.Value
	var l sync.Mutex
	type Map map[int]interface{}
	var count int32

	m := make(Map)
	for i:=0; i<10; i++ {
		m[i] = i
	}
	v.Store(m)

	// Copy-on-write 思想
	go func() {
		for {
			l.Lock()
			defer l.Unlock()
			m1 := v.Load().(Map)
			m2 := make(Map)

			for r, v := range m1 {
				m2[r] = v.(int)+10
			}
			v.Store(m2)
		}
	}()

	var wg sync.WaitGroup
	for n := 0; n < 2000; n++ {
		wg.Add(1)
		go func() {
			for n := 0; n <= b.N; n++ {
				T(v.Load().(Map)[1].(int))
				atomic.AddInt32(&count,1)
				//fmt.Println(cfg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	b.Logf("atomic count:%v\n", count)
}

func TestSecKill(t *testing.T) {
	proRdb := db.NewRedisConn()
	lockRdb := lock.NewRedisPool()
	var wg sync.WaitGroup

	for i:=0; i<1000; i++ {
		wg.Add(1)
		go func() {

			lockRdb.Lock()
			num, err := proRdb.Get(lockRdb.Ctx,"1").Result()
			if err != nil {
				t.Logf("cannot get product nums")
			}
			if num != "0" {

				rest, err := proRdb.DecrBy(lockRdb.Ctx,"1", 1).Result()
				if err != nil {
					t.Logf("decr failed")
				}
				t.Logf("rest product num:%v", rest)
			} else {
				t.Logf("no more product")
			}

			lockRdb.UnLock()
			wg.Done()
		}()
	}
	wg.Wait()
}