package main

import (
	"github.com/gin-gonic/gin"
	"seckill/controllers"
)

func main() {

	r := gin.Default()
	r.Use(gin.Recovery())
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	// 商品管理
	productParty := r.Group("/product")
	productParty.GET("/all", controllers.GetAllProduct)         // 获取所有商品信息
	productParty.GET("/manager", controllers.GetManager)        // 商品管理
	productParty.GET("/add", controllers.GetAdd)                // 商品添加页面
	productParty.GET("/delete", controllers.DeleteProductInfo)  // 删除指定商品信息
	productParty.POST("/update", controllers.UpdateProductInfo) // 修改指定商品信息
	productParty.POST("/add", controllers.AddProductInfo)       // 新增商品信息

	// 订单管理
	orderParty := r.Group("/order")

	r.Run(":8080")
}
