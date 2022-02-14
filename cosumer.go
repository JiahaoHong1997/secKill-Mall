package main

import (
	"seckill/common/rabbitmq"
	"seckill/dao"
	db2 "seckill/dao/db"
	"seckill/service"
)

func main() {
	db := db2.DBConn()

	//创建product数据库操作实例
	product := dao.NewProductManager("product", db)
	//创建product serivce
	productService := service.NewProductService(product)
	//创建Order数据库实例
	order := dao.NewOrderManagerRepository("order", db)
	//创建order Service
	orderService := service.NewOrderService(order)

	rabbitmqConsumeSimple := RabbitMQ.NewRabbitMQSimple("secKillProduct")
	rabbitmqConsumeSimple.ConsumeSimple(orderService, productService)
}
