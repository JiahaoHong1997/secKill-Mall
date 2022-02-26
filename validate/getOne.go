package main

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

var sum int32

// 预存商品数量
var productNum int32 = 200000

// 互斥锁
var mutex sync.Mutex

var isSync bool

// 获取秒杀商品
func GetOneProduct() bool {

	if isSync {
		mutex.Lock()
		defer mutex.Unlock()
		if sum < productNum {
			sum += 1
			return true
		}
		return false
	} else {
		atomic.AddInt32(&sum, 1)
		if sum >= 195000 {
			isSync = true
		}
		return true
	}
}

func GetProduct(w http.ResponseWriter, r *http.Request) {
	if GetOneProduct() {
		w.Write([]byte("true"))
		return
	}
	w.Write([]byte("false"))
	return
}

func main() {
	http.HandleFunc("/getOne", GetProduct)
	err := http.ListenAndServe(":8084", nil)
	if err != nil {
		log.Fatal("Err:", err)
	}
}
