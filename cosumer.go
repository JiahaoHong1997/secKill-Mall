package main

import (
	"seckill/common"
	rabbitmq "seckill/rabbitmq"
	"seckill/repositories"
	"seckill/service"
)

func main() {
	db := common.DBConn()

	//创建product数据库操作实例
	product := repositories.NewProductManager("product", db)
	//创建product serivce
	productService := service.NewProductService(product)
	//创建Order数据库实例
	order := repositories.NewOrderManagerRepository("order", db)
	//创建order Service
	orderService := service.NewOrderService(order)

	rabbitmqConsumeSimple := rabbitmq.NewRabbitMQSimple("secKillProduct")
	rabbitmqConsumeSimple.ConsumeSimple(orderService, productService)
}
