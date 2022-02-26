package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
)

var sum int32

// 预存商品数量
var productNum [100000]int64

// 互斥锁
var mutex sync.Mutex

var isSync bool

// 获取秒杀商品
func GetOneProduct(productId int64, requestNum int64) bool {

	if isSync {
		mutex.Lock()
		defer mutex.Unlock()
		if productNum[productId] >= requestNum {
			productNum[productId] -= requestNum
			return true
		}
		return false
	} else {
		atomic.AddInt64(&productNum[productId], -1*requestNum)
		if sum <= 5000 {
			isSync = true
		}
		return true
	}
}

func GetProduct(w http.ResponseWriter, r *http.Request) {

	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 || len(queryForm["requestNum"]) <= 0 {
		w.Write([]byte("false"))
		w.WriteHeader(http.StatusForbidden)
		return
	}
	productString := queryForm["productID"][0]
	requestNumString := queryForm["requestNum"][0]
	productId, _ := strconv.ParseInt(productString, 10, 64)
	requestNum, _ := strconv.ParseInt(requestNumString, 10, 64)

	if GetOneProduct(productId, requestNum) {
		w.Write([]byte("true"))
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Write([]byte("false"))
	w.WriteHeader(http.StatusForbidden)
	return
}

func AddProduct(w http.ResponseWriter, r *http.Request) {

	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 || len(queryForm["productNum"]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	productNumString := queryForm["productNum"][0]

	productId, _ := strconv.ParseInt(productString, 10, 64)
	proNum, _ := strconv.ParseInt(productNumString, 10, 64)
	productNum[productId] = proNum
	w.Write([]byte("success"))
	w.WriteHeader(http.StatusOK)
	return
}

func main() {
	http.HandleFunc("/getProduct", GetProduct)
	http.HandleFunc("/addProduct", AddProduct)
	err := http.ListenAndServe(":8084", nil)
	if err != nil {
		log.Fatal("Err:", err)
	}
}
