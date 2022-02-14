package service

import (
	"seckill/dao"
	"seckill/models"
)

type IProductService interface {
	GetProductByID(int64) (*models.Product, error)
	GetAllProduct() ([]*models.Product, error)
	DeleteProductID(int64) (bool, error)
	InsertProduct(*models.Product) (int64, error)
	UpdateProduct(*models.Product) error
	SubNumberOne(int64) error
	InsertSecProduct(int64, int64, float64) error
}

type ProductService struct {
	productRepository dao.IProduct
}

func NewProductService(product dao.IProduct) IProductService {
	return &ProductService{productRepository: product}
}

func (p *ProductService) GetProductByID(productID int64) (*models.Product, error) {
	return p.productRepository.SelectByKey(productID)
}

func (p *ProductService) GetAllProduct() ([]*models.Product, error) {
	return p.productRepository.SelectAll()
}

func (p *ProductService) DeleteProductID(productID int64) (bool, error) {
	return p.productRepository.Delete(productID)
}

func (p *ProductService) InsertProduct(product *models.Product) (int64, error) {
	return p.productRepository.Insert(product)
}

func (p *ProductService) UpdateProduct(product *models.Product) error {
	return p.productRepository.Update(product)
}

func (p *ProductService) SubNumberOne(productID int64) error {
	return p.productRepository.SubProductNum(productID)
}

func (p *ProductService) InsertSecProduct(productID int64, productNum int64, duration float64) error {
	return p.productRepository.AddSecProduct(productID, productNum, duration)
}
