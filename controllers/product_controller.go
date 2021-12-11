package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"seckill/common"
	"seckill/datamodels"
	"seckill/repositories"
	"seckill/service"
	"strconv"
)

var productRepository repositories.IProduct
var productService service.IProductService

func init() { // 实例化
	db := common.DBConn()
	productRepository = repositories.NewProductManager("product", db)
	productService = service.NewProductService(productRepository)
}

func GetAllProduct(c *gin.Context) {
	productArray, _ := productService.GetAllProduct()

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
	product, err := productService.GetProductByID(id)
	if err != nil {
		log.Printf("product: Failed to get product id: %s", err)
	}
	c.HTML(http.StatusOK, "manager.tmpl", gin.H{
		"product": product,
	})
}

func GetAdd(c *gin.Context) {
	c.HTML(http.StatusOK, "add.tmpl", nil)
}

func UpdateProductInfo(c *gin.Context) {
	product := &datamodels.Product{}
	c.Request.ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "secKillSystem"})
	if err := dec.Decode(c.Request.Form, product); err != nil {
		log.Printf("product UpdateProductInfo: Failed to decode the form: %s", err)
	}
	err := productService.UpdateProduct(product)
	if err != nil {
		log.Printf("product: Failed to update to product: %s", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all") // 重定向

}

func AddProductInfo(c *gin.Context) {
	product := &datamodels.Product{}
	c.Request.ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "secKillSystem"})
	if err := dec.Decode(c.Request.Form, product); err != nil {
		log.Printf("product AddProductInfo: Failed to decode the form: %s", err)
	}
	_, err := productService.InsertProduct(product)
	if err != nil {
		log.Printf("product: Failed to add product: %s", err)
	}
	c.Redirect(http.StatusMovedPermanently, "all")

}

func DeleteProductInfo(c *gin.Context) {
	idString := c.Query("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		log.Printf("product DeleteProduct: Failed to transform to int type: %s", err)
	}
	isOk := productService.DeleteProductID(id)
	if isOk {
		log.Printf("删除商品成功，ID为：" + idString)
	} else {
		log.Printf("删除商品失败，ID为：" + idString)
	}
	c.Redirect(http.StatusMovedPermanently, "all")
}
