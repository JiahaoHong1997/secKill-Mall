package common

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	rdb      *redis.Client
	lockKey  = "my_lock"
	unlockCh = make(chan struct{})
)

type RedisPool struct {
	Rdb      *redis.Client
	LockKey  string
	UnlockCh chan struct{}
}

// 初始化连接
func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
		PoolSize: 100,
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func NewRedisPool() *RedisPool {
	return &RedisPool{
		Rdb:      rdb,
		LockKey:  lockKey,
		UnlockCh: unlockCh,
	}
}

func (r *RedisPool) RedisConn() *redis.Client {
	return r.Rdb
}

// 获取当前 goroutine 的协程ID
func GetCurrentGoroutineId() (int, error) {
	buf := make([]byte, 128)
	buf = buf[:runtime.Stack(buf, false)]
	stackInfo := string(buf)
	goIdStr := strings.TrimSpace(strings.Split(strings.Split(stackInfo, "[running]")[0], "goroutine")[1])
	goId, err := strconv.Atoi(goIdStr)
	if err != nil {
		return 0, errors.Wrap(err, "Got goroutineId failed!")
	}
	return goId, nil
}

func (r *RedisPool) Lock() {
	var resp *redis.BoolCmd
	for {
		goId, err := GetCurrentGoroutineId()
		if err != nil {
			log.Printf("original error: %T, %v", errors.Cause(err), errors.Cause(err))
			return
		}
		resp = r.Rdb.SetNX(r.LockKey, goId, 10*time.Second)
		lockSuccess, err := resp.Result()
		if err == nil && lockSuccess {
			fmt.Println("lock success!", goId)
			//抢锁成功，开启看门狗 并跳出，否则失败继续自旋
			go r.WatchDog(goId)
			return
		} else {
			//log.Println("lock failed!", err)
		}

		// 抢锁失败，继续自旋
	}
}

// 使用 lua 脚本保证 UnLock 的原子性
func (r *RedisPool) UnLock() {
	script := redis.NewScript(`
	if redis.call('get', KEYS[1]) == ARGV[1]
	then
		return redis.call('del', KEYS[1])
	else
		return 0
	end`)

	goId, err := GetCurrentGoroutineId()
	if err != nil {
		log.Println("unlock failed!", err)
		return
	}

	// 在确认当前锁是自己的锁之后，删除锁之前，这段时间内，锁可能会恰巧过期释放且被其他竞争者抢占，那么继续删除则删除的是别人的锁，会出现误删问题。
	resp := script.Run(r.Rdb, []string{r.LockKey}, goId)
	if result, err := resp.Result(); err != nil || result == 0 {
		fmt.Println("unlock failed!", err)
	} else {
		//删锁成功后，通知看门狗退出
		r.UnlockCh <- struct{}{}
	}
}

// 自动续期看门狗
func (r *RedisPool) WatchDog(goId int) {
	// 创建一个定时器NewTicker, 每隔8s触发一次
	expTicker := time.NewTicker(8 * time.Second)

	// 确认锁与锁续期打包原子化
	script := redis.NewScript(`
	if redis.call('get', KEYS[1]) == ARGV[1]
	then
		return redis.call('expire', KEYS[1], ARGV[2])
	else
		return 0
	end`)

	for {
		select {
		case <-expTicker.C:
			resp := script.Run(r.Rdb, []string{r.LockKey}, goId, 10)
			if result, err := resp.Result(); err != nil || result == int64(0) {
				log.Println("expire lock failed", err)
			}
		case <-r.UnlockCh: //任务完成后用户解锁通知看门狗退出
			return
		}
	}
}
