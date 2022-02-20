package bloom

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
)

func TestBloom(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6378",
		Password: "",
		PoolSize: 100,
	})

	bloom := NewBloom(rdb)
	bloom.Add("newClient") //往过滤器写入数据
	b := bloom.Exist("aaa") //判断是否存在这个值
	fmt.Println(b)
	b = bloom.Exist("newClient")
	fmt.Println(b)

	b = bloom.Exist("colar")
	fmt.Println(b)
	bloom.Add("colar")
	b = bloom.Exist("colar")
	fmt.Println(b)
}