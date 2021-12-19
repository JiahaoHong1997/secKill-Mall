package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"seckill/common"
	"seckill/datamodels"
	rabbitmq "seckill/rabbitmq"
	"seckill/repositories"
	"seckill/service"
	"strconv"
)

var productRepository repositories.IProduct
var productService service.IProductService
var orderRepository repositories.IOrderRepository
var orderService service.IOrderService
var rabbitMq *rabbitmq.RabbitMQ

func init() {
	db := common.DBConn()
	productRepository = repositories.NewProductManager("product", db)
	productService = service.NewProductService(productRepository)
	orderRepository = repositories.NewOrderManagerRepository("order", db)
	orderService = service.NewOrderService(orderRepository)
	rabbitMq = rabbitmq.NewRabbitMQSimple("secKillProduct")
}

// 秒杀页面
func GetDetail(c *gin.Context) {
	productString := c.Query("productID")
	productID, err := strconv.Atoi(productString)
	if err != nil {
		log.Println(err)
	}
	product, err := productService.GetProductByID(int64(productID))
	if err != nil {
		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace: %+v", err)
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

	message := datamodels.NewMessage(userID, productID)
	byteMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
	}

	err = rabbitMq.PublishSimple(string(byteMessage))
	if err != nil {
		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
	}
	c.String(200, "true")

	//product, err := productService.GetProductByID(int64(productID))
	//if err != nil {
	//	log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
	//	log.Printf("stack trace: %+v", err)
	//}

	//var orderID int64
	//showMessage := "抢购失败"
	//// 判断商品数量是否满足需求
	//// TODO:高并发需求还未实现
	//if product.ProductNum > 0 {
	//	// 扣除商品数量
	//	product.ProductNum -= 1
	//	err = productService.UpdateProduct(product)
	//	if err != nil {
	//		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
	//		log.Printf("stack trace: %+v", err)
	//	}
	//
	//	// 创建订单
	//	userID, err := strconv.Atoi(userString)
	//	if err != nil {
	//		log.Println("string false")
	//	}
	//	order := &datamodels.Order{
	//		UserId:      int64(userID),
	//		ProductId:   int64(productID),
	//		OrderStatus: datamodels.OrderSuccess,
	//	}
	//
	//	// 新建订单
	//	orderID, err = orderService.InsertOrder(order)
	//	if err != nil {
	//		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
	//		log.Printf("stack trace: %+v", err)
	//	}
	//	showMessage = "抢购成功"
	//}
	//c.HTML(http.StatusOK, "result.tmpl", gin.H{
	//	"showMessage": showMessage,
	//	"orderID":     orderID,
	//})
}
