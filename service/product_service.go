package service

import (
	"seckill/datamodels"
	"seckill/repositories"
)

type IProductService interface {
	GetProductByID(int64) (*datamodels.Product, error)
	GetAllProduct() ([]*datamodels.Product, error)
	DeleteProductID(int64) (bool, error)
	InsertProduct(*datamodels.Product) (int64, error)
	UpdateProduct(*datamodels.Product) error
	SubNumberOne(int64) error
}

type ProductService struct {
	productRepository repositories.IProduct
}

func NewProductService(product repositories.IProduct) IProductService {
	return &ProductService{productRepository: product}
}

func (p *ProductService) GetProductByID(productID int64) (*datamodels.Product, error) {
	return p.productRepository.SelectByKey(productID)
}

func (p *ProductService) GetAllProduct() ([]*datamodels.Product, error) {
	return p.productRepository.SelectAll()
}

func (p *ProductService) DeleteProductID(productID int64) (bool, error) {
	return p.productRepository.Delete(productID)
}

func (p *ProductService) InsertProduct(product *datamodels.Product) (int64, error) {
	return p.productRepository.Insert(product)
}

func (p *ProductService) UpdateProduct(product *datamodels.Product) error {
	return p.productRepository.Update(product)
}

func (p *ProductService) SubNumberOne(productID int64) error {
	return p.productRepository.SubProductNum(productID)
}
