package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"seckill/common/rabbitmq"
	"seckill/dao"
	db2 "seckill/dao/db"
	"seckill/models"
	"seckill/service"
	"strconv"
)

var productRepository dao.IProduct
var productService service.IProductService
var orderRepository dao.IOrderRepository
var orderService service.IOrderService
var rabbitMq *RabbitMQ.RabbitMQ
var cacheMq *RabbitMQ.RabbitMQ

func init() {
	db := db2.DBConn()
	rdb := db2.NewRedisConn()
	cache := db2.NewCachePool()
	productRepository = dao.NewProductManager("product", db, rdb, cache)
	productService = service.NewProductService(productRepository)
	orderRepository = dao.NewOrderManagerRepository("order", db)
	orderService = service.NewOrderService(orderRepository)
	rabbitMq = RabbitMQ.NewRabbitMQSimple("secKillProduct")
	cacheMq = RabbitMQ.NewRabbitMQSimple("cacheMq")
}

// 秒杀页面
func GetDetail(c *gin.Context) {
	productString := c.Query("productID")
	productID, err := strconv.Atoi(productString)
	if err != nil {
		log.Println(err)
	}
	product, err, cached := productService.GetProductByID(int64(productID))
	if err != nil {
		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace: %+v", err)
	}
	if !cached {
		cacheMessage := models.NewMessageCache(product.ID, product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl, false)
		message, _ := json.Marshal(cacheMessage)
		err = cacheMq.PublishSimple(string(message))
		if err != nil {
			log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
			log.Printf("stack trace:%+v", err)
		}
	}
	c.HTML(http.StatusOK, "user_view.tmpl", gin.H{
		"product": product,
	})
}

func GetOrder(c *gin.Context) {

	productString := c.Query("productID")
	userString, err := c.Cookie("uid")
	if err != nil {
		log.Println(errors.New("cookie false"))
	}
	productID, err := strconv.ParseInt(productString, 10, 64)
	if err != nil {
		log.Println(errors.New("string false"))
	}

	userID, err := strconv.ParseInt(userString, 10, 64)
	if err != nil {
		log.Println(errors.New("string false"))
	}

	message := models.NewMessage(userID, productID)
	byteMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
	}

	err = rabbitMq.PublishSimple(string(byteMessage))
	if err != nil {
		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
	}
	c.String(200, "true")
}
