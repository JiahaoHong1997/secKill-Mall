package manage

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"seckill/common"
	RabbitMQ "seckill/common/rabbitmq"
	"seckill/dao"
	db2 "seckill/dao/db"
	"seckill/models"
	"seckill/service"
	"strconv"
)

var productRepository dao.IProduct
var productService service.IProductService
var cacheMq *RabbitMQ.RabbitMQ

func init() { // 实例化
	db := db2.DBConn()
	rdb := db2.NewRedisConn()
	cache := db2.NewCachePool()
	productRepository = dao.NewProductManager("product", db, rdb, cache)
	productService = service.NewProductService(productRepository)
	cacheMq = RabbitMQ.NewRabbitMQSimple("cacheMq")
}

func GetAllProduct(c *gin.Context) {
	productArray, err := productService.GetAllProduct()
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}

	c.HTML(http.StatusOK, "view.tmpl", gin.H{
		"productArray": productArray,
	})
}

func ManageProductByID(c *gin.Context) {
	idString := c.Query("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		log.Printf("product ManageProductByID: Failed to transform to int type: %s", err)
	}
	product, err, cached := productService.GetProductByID(id)
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}

	// 更新缓存：将查询到的结构体送到消息队列中
	if !cached {
		cacheMessage := models.NewMessageCache(id, product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl, false)
		fmt.Println(cacheMessage)
		message, _ := json.Marshal(cacheMessage)
		err = cacheMq.PublishSimple(string(message))
		if err != nil {
			log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
			log.Printf("stack trace:%+v", err)
		}
	}
	c.HTML(http.StatusOK, "manager.tmpl", gin.H{
		"product": product,
	})
}

func GetProductAdd(c *gin.Context) {
	c.HTML(http.StatusOK, "add.tmpl", nil)
}

func UpdateProductInfo(c *gin.Context) {
	product := &models.Product{}
	c.Request.ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "secKillSystem"})
	if err := dec.Decode(c.Request.Form, product); err != nil {
		log.Printf("product UpdateProductInfo: Failed to decode the form: %s", err)
	}

	id, err := productService.UpdateProduct(product)
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}

	cacheMessage := models.NewMessageCache(id, product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl, false)
	message, _ := json.Marshal(cacheMessage)
	err = cacheMq.PublishSimple(string(message))
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all") // 重定向
}

func AddProductInfo(c *gin.Context) {
	product := &models.Product{}
	c.Request.ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "secKillSystem"})
	if err := dec.Decode(c.Request.Form, product); err != nil {
		log.Printf("product AddProductInfo: Failed to decode the form: %s", err)
	}
	id, err := productService.InsertProduct(product)
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}

	// 更新缓存：将查询到的结构体送到消息队列中
	cacheMessage := models.NewMessageCache(id, product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl, false)
	message, _ := json.Marshal(cacheMessage)
	err = cacheMq.PublishSimple(string(message))
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all")
}

func DeleteProductInfo(c *gin.Context) {
	idString := c.Query("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		log.Printf("product DeleteProduct: Failed to transform to int type: %s", err)
	}

	// 先删库
	isOk, _ := productService.DeleteProductID(id)
	if isOk {
		log.Printf("删除商品成功，ID为：" + idString)
	} else {
		log.Printf("删除商品失败，ID为：" + idString)
	}

	// 再删缓存
	cacheMessage := models.NewMessageCache(id, "", 0, "", "", true)
	message, _ := json.Marshal(cacheMessage)
	err = cacheMq.PublishSimple(string(message))
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all")
}

func GetProductAddSec(c *gin.Context) {
	idString := c.Query("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		log.Printf("product ManageProductByID: Failed to transform to int type: %s", err)
	}
	product, err, cached := productService.GetProductByID(id)
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace:%+v", err)
	}
	if !cached {
		cacheMessage := models.NewMessageCache(product.ID, product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl, false)
		message, _ := json.Marshal(cacheMessage)
		err = cacheMq.PublishSimple(string(message))
		if err != nil {
			log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
			log.Printf("stack trace:%+v", err)
		}
	}
	c.HTML(http.StatusOK, "addsec.tmpl", gin.H{
		"product": product,
	})
}

func AddProductSecInfo(c *gin.Context) {
	c.Request.ParseForm()
	productID := c.PostForm("ProductID")
	id, err := strconv.ParseInt(productID, 10, 64)
	if err != nil {
		log.Printf("product AddProductSecInfo: Failed to transform to int type: %s", err)
	}
	productNum := c.PostForm("ProductNum")
	num, err := strconv.ParseInt(productNum, 10, 64)
	if err != nil {
		log.Printf("product AddProductSecInfo: Failed to transform to int type: %s", err)
	}
	countDown := c.PostForm("Countdown")
	duration, err := strconv.ParseFloat(countDown, 64)
	if err != nil {
		log.Printf("product AddProductSecInfo: Failed to transform to int type: %s", err)
	}
	err = productService.InsertSecProduct(id, num, duration)
	if err != nil {
		log.Printf("original error:%T %v\n", errors.Cause(err), errors.Cause(err))
	}
	c.Redirect(http.StatusMovedPermanently, "all")
}
