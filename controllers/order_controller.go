package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
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
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}
	c.HTML(http.StatusOK, "managerOrder.tmpl", gin.H{
		"order": order,
	})
}

func GetOrderAdd(c *gin.Context) {
	c.HTML(http.StatusOK, "addOrder.tmpl", nil)
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
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all") // 重定向
}

func AddOrderInfo(c *gin.Context) {
	order := &datamodels.Order{}
	c.Request.ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "secKillSystem"})
	if err := dec.Decode(c.Request.Form, order); err != nil {
		log.Printf("order AddOrderInfo: Failed to decode the form: %s", err)
	}

	_, err := orderService.InsertOrder(order)
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all")
}

func DeleteOrderInfo(c *gin.Context) {
	idString := c.Query("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		log.Printf("order DeleteOrderInfo: Failed to transform to int type: %s", err)
	}
	isOk, _ := orderService.DeleteOrderByID(id)
	if isOk {
		log.Printf("删除订单成功，ID为：" + idString)
	} else {
		log.Printf("删除订单失败，ID为：" + idString)
	}
	c.Redirect(http.StatusMovedPermanently, "all")
}
