package main

import (
	"github.com/JiahaoHong1997/altria-web"
	"seckill/controllers"
)

func main() {

	r := altria.Saber()
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	// 注册控制器
	productParty := r.Group("/product")
	productParty.GET("/allproduct", controllers.GetAllHandler)

	r.Run(":8080")
}
