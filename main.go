package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"seckill/controllers"
)

func main() {

	r := gin.Default()
	r.Use(gin.Recovery())
	r.LoadHTMLGlob("templates/**/*")

	store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))

	store.Options(sessions.Options{
		MaxAge: int(30*60),
		Path: "/",
	})
	r.Use(sessions.Sessions("sessionId", store))

	// 后台管理功能
	manageParty := r.Group("/manage")
	manageParty.Static("/assets", "./static/backend")
	// 商品管理
	productParty := manageParty.Group("/product")
	productParty.GET("/all", controllers.GetAllProduct)         // 获取所有商品信息
	productParty.GET("/manager", controllers.ManageProductByID) // 商品管理
	productParty.GET("/add", controllers.GetProductAdd)         // 商品添加页面
	productParty.GET("/delete", controllers.DeleteProductInfo)  // 删除指定商品信息
	productParty.POST("/update", controllers.UpdateProductInfo) // 修改指定商品信息
	productParty.POST("/add", controllers.AddProductInfo)       // 新增商品信息
	// 订单管理
	orderParty := manageParty.Group("/order")
	orderParty.GET("/all", controllers.GetAllOrder)         // 获取所有订单信息
	orderParty.GET("/manager", controllers.ManageOrderByID) // 订单管理
	orderParty.GET("/add", controllers.GetOrderAdd)         // 订单添加页面
	orderParty.GET("/delete", controllers.DeleteOrderInfo)  // 删除指定订单信息
	orderParty.POST("/update", controllers.UpdateOrderInfo) // 修改指定订单信息
	orderParty.POST("add", controllers.AddOrderInfo)        // 新增订单信息


	// 前台用户功能
	userParty := r.Group("/user")
	userParty.Static("/assets","./static/fronted")
	userParty.GET("/register", controllers.GetRegister)
	userParty.GET("/login", controllers.GetLogin)
	userParty.GET("/error", controllers.GetError)
	userParty.POST("/register", controllers.PostRegister)
	userParty.POST("/login", controllers.PostLogin)


	r.Run(":8080")
}
