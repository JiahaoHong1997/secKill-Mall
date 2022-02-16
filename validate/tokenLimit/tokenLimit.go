package tokenLimit

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
	"seckill/common/limit"
)

var (
	l *limit.TokenLimiter
	// 令牌桶容量
	burst = 1000
	// 令牌生成速率
	rate = 500
)

func init() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	limitPool := limit.NewLimitConn(ctx)
	l = limit.NewTokenLimiter(rate, burst, limitPool, "tokenLimit")
}

// 令牌桶算法，用于拦截请求，保证服务可用
func LimitT(w http.ResponseWriter, r *http.Request) error {
	if !l.Allow() {
		return errors.New("Request limit exceeded")
	}
	return nil
}
