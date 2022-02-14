package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"seckill/controllers"
	"seckill/controllers/manage"
	"seckill/middleware"
	"syscall"
	"time"
)

func main() {

	r := gin.Default()
	r.Use(gin.Recovery())
	r.LoadHTMLGlob("templates/**/*")

	// 后台管理功能
	manageParty := r.Group("/manage")
	manageParty.Static("/assets", "./static/backend")
	// 商品管理
	productParty := manageParty.Group("/product")
	productParty.GET("/all", manage.GetAllProduct)         // 获取所有商品信息
	productParty.GET("/info", manage.ManageProductByID) // 商品管理
	productParty.GET("/add", manage.GetProductAdd)         // 商品添加页面
	productParty.GET("/delete", manage.DeleteProductInfo)  // 删除指定商品信息
	productParty.POST("/update", manage.UpdateProductInfo) // 修改指定商品信息
	productParty.POST("/add", manage.AddProductInfo)       // 新增商品信息
	productParty.GET("/addsec", manage.GetProductAddSec)   // 获取加入秒杀页面
	productParty.POST("/updatesec", manage.AddProductSecInfo) // 加入指定商品到秒杀池
	// 订单管理
	orderParty := manageParty.Group("/order")
	orderParty.GET("/all", manage.GetAllOrder)         // 获取所有订单信息
	orderParty.GET("/manager", manage.ManageOrderByID) // 订单管理
	orderParty.GET("/add", manage.GetOrderAdd)         // 订单添加页面
	orderParty.GET("/delete", manage.DeleteOrderInfo)  // 删除指定订单信息
	orderParty.POST("/update", manage.UpdateOrderInfo) // 修改指定订单信息
	orderParty.POST("/add", manage.AddOrderInfo)       // 新增订单信息

	// 前台用户登录注册功能
	userParty := r.Group("/user")
	userParty.Static("/assets", "./static/fronted")
	userParty.GET("/register", controllers.GetRegister)   // 注册页面
	userParty.GET("/login", controllers.GetLogin)         // 登录页面
	userParty.GET("/error", controllers.GetError)         // 访问错误页面
	userParty.POST("/register", controllers.PostRegister) // 新用户注册请求
	userParty.POST("/login", controllers.PostLogin)       // 用户登录功能

	// 用户购物功能
	userProductParty := r.Group("/product")
	userProductParty.Use(middleware.Auth())
	userProductParty.GET("/detail", controllers.GetDetail)
	userProductParty.GET("/order", controllers.GetOrder)

	serverOut := make(chan struct{}) // 访问 /shutdown，优雅退出
	r.GET("/shutdown", func(c *gin.Context) {
		c.String(200, "The http server will shutdown in 5 seconds!")
		serverOut <- struct{}{}
	})
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		if err := server.ListenAndServe(); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		select {
		case <-ctx.Done():
			log.Printf("errgroup exit...")
		case <-serverOut:
			log.Println("request `/shutdown`, http server will stop in 5 seconds!")
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		return err
	})

	g.Go(func() error { // CTRL + C，直接退出
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sig := <-signals:
				return errors.New("\nget signal " + sig.String() + ", application will shutdown\n")
			}
		}
	})

	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
}
