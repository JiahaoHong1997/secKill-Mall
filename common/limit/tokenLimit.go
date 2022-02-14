package limit

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const (
	script = `local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)
local last_tokens = tonumber(redis.call("get", KEYS[1]))
if last_tokens == nil then
    last_tokens = capacity
end
local last_refreshed = tonumber(redis.call("get", KEYS[2]))
if last_refreshed == nil then
    last_refreshed = 0
end
local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end
redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)
return allowed`
	tokenFormat     = "{%s}.tokens"
	timestampFormat = "{%s}.ts"
)

type TokenLimiter struct {
	rate         int
	burst        int
	store        *RedisPool
	tokenKey     string
	timestampKey string
	redisAlive   uint32
}

func NewTokenLimiter(rate, burst int, store *RedisPool, key string) *TokenLimiter {
	tokenKey := fmt.Sprintf(tokenFormat, key)
	timestampKey := fmt.Sprintf(timestampFormat, key)

	return &TokenLimiter{
		rate:         rate,
		burst:        burst,
		store:        store,
		tokenKey:     tokenKey,
		timestampKey: timestampKey,
		redisAlive:   1,
	}
}

func (lim *TokenLimiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

func (lim *TokenLimiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(now, n)
}

func (lim *TokenLimiter) reserveN(now time.Time, n int) bool {

	rs := redis.NewScript(script)
	resp := rs.Run(lim.store.Ctx,
		lim.store.Rdb,
		[]string {
			lim.tokenKey,
			lim.timestampKey,
		},
		strconv.Itoa(lim.rate),
		strconv.Itoa(lim.burst),
		strconv.FormatInt(now.Unix(), 10),
		strconv.Itoa(n),)

	result, err := resp.Result()
	if err != nil || result.(int64) == 0 {
		return false
	}

	return result.(int64) == 1
}