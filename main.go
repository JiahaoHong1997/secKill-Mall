package main

import (
	"github.com/gin-gonic/gin"
	"seckill/controllers"
)

func main() {

	r := gin.Default()
	r.Use(gin.Recovery())
	r.LoadHTMLGlob("templates/**/*")
	r.Static("/assets", "./static")

	// 商品管理
	productParty := r.Group("/product")
	productParty.GET("/all", controllers.GetAllProduct)         // 获取所有商品信息
	productParty.GET("/manager", controllers.ManageProductByID) // 商品管理
	productParty.GET("/add", controllers.GetProductAdd)         // 商品添加页面
	productParty.GET("/delete", controllers.DeleteProductInfo)  // 删除指定商品信息
	productParty.POST("/update", controllers.UpdateProductInfo) // 修改指定商品信息
	productParty.POST("/add", controllers.AddProductInfo)       // 新增商品信息

	// 订单管理
	orderParty := r.Group("/order")
	orderParty.GET("/all", controllers.GetAllOrder)         // 获取所有订单信息
	orderParty.GET("/manager", controllers.ManageOrderByID) // 订单管理
	orderParty.GET("/add", controllers.GetOrderAdd)         // 订单添加页面
	orderParty.GET("/delete", controllers.DeleteOrderInfo)  // 删除指定订单信息
	orderParty.POST("/update", controllers.UpdateOrderInfo) // 修改指定订单信息
	orderParty.POST("add", controllers.AddOrderInfo)        // 新增订单信息

	r.Run(":8080")
}
