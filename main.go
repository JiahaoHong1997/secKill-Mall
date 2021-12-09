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

	// 注册控制器
	productParty := r.Group("/product")
	productParty.GET("/all", controllers.GetAllHandler)
	productParty.GET("/manager", controllers.GetManager)
	productParty.GET("/add", controllers.GetAdd)
	productParty.GET("/delete", controllers.DeleteProductInfo)
	productParty.POST("/update", controllers.UpdateProductInfo)
	productParty.POST("/add", controllers.AddProductInfo)

	r.Run(":8080")
}
