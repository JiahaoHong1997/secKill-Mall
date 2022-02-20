package bloom

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Bloom struct{
	Store *redis.Client
	Key string
	HashFuncs []F //保存hash函数
}
func NewBloom(store *redis.Client) *Bloom{
	return &Bloom{Store:store, Key:"bloom", HashFuncs:NewFunc()}
}
func (b *Bloom)Add(str string) error{
	var err error
	for _,f := range b.HashFuncs {
		offset := f(str)
		_, err := b.Store.Do(context.Background(),"setbit", b.Key, offset,1).Result()
		if err != nil {
			return err
		}
	}
	return err
}
func (b *Bloom)Exist(str string) bool{
	var a int64 = 1
	for _,f := range b.HashFuncs {
		offset := f(str)
		bitValue,_ := b.Store.Do(context.Background(),"getbit", b.Key, offset).Result()
		if bitValue != a {
			return false
		}
	}
	return true
}
