package controllers

import (
	"github.com/JiahaoHong1997/altria-web"
	"net/http"
	"seckill/common"
	"seckill/repositories"
	"seckill/service"
)

var productRepository repositories.IProduct
var productService service.IProductService

func init() {
	db := common.DBConn()
	productRepository = repositories.NewProductManager("product", db)
	productService = service.NewProductService(productRepository)
}

func GetAllHandler(c *altria.Context) {
	productArray, _ := productService.GetAllProduct()
	c.HTML(http.StatusOK, "view.tmpl", altria.H{
		"productArray": productArray,
	})
}
