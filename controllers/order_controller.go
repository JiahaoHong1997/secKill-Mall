package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"seckill/common"
	"seckill/datamodels"
	"seckill/repositories"
	"seckill/service"
	"strconv"
)

var orderRepository repositories.IOrderRepository
var orderService service.IOrderService

func init() {
	db := common.DBConn()
	orderRepository = repositories.NewOrderManagerRepository("order", db)
	orderService = service.NewOrderService(orderRepository)
}

func GetAllOrder(c *gin.Context) {
	orderArray, err := orderService.GetAllOrderInfo()
	if err != nil {
		log.Printf("order controller: failed to query all order information, %v", err)
	}

	c.HTML(http.StatusOK, "viewOrder.tmpl", gin.H{
		"order": orderArray,
	})
}

func ManageOrderByID(c *gin.Context) {
	idString := c.Query("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		log.Printf("order ManageOrderByID: Failed to transform to int type: %s", err)
	}
	order, err := orderService.GetOrderByID(id)
	if err != nil {
		log.Printf("order ManageOrderByID: Failed to get order id: %s", err)
	}
	c.HTML(http.StatusOK, "managerOrder.tmpl", gin.H{
		"order": order,
	})
}

func UpdateOrderInfo(c *gin.Context) {
	order := &datamodels.Order{}
	c.Request.ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "secKillSystem"})
	if err := dec.Decode(c.Request.Form, order); err != nil {
		log.Printf("order UpdateOrderInfo: Failed to decode the form: %s", err)
	}
	err := orderService.UpdateOrder(order)
	if err != nil {
		log.Printf("order UpdateOrderInfo: Failed to update to order: %s", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all") // 重定向
}
