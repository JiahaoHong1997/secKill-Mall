package service

import (
	"seckill/dao"
	"seckill/models"
)

type IOrderService interface {
	GetOrderByID(int64) (*models.Order, error)
	DeleteOrderByID(int64) (bool, error)
	UpdateOrder(*models.Order) error
	InsertOrder(*models.Order) (int64, error)
	GetAllOrder() ([]*models.Order, error)
	GetAllOrderInfo() (map[int]map[string]string, error)
	InsertOrderByMessage(*models.Message) (int64, error)
}

type OrderService struct {
	OrderRepository dao.IOrderRepository
}

func NewOrderService(repository dao.IOrderRepository) IOrderService {
	return &OrderService{OrderRepository: repository}
}

func (o *OrderService) GetOrderByID(orderID int64) (*models.Order, error) {
	return o.OrderRepository.SelectByKey(orderID)
}

func (o *OrderService) DeleteOrderByID(productID int64) (bool, error) {
	return o.OrderRepository.Delete(productID)
}

func (o *OrderService) UpdateOrder(order *models.Order) error {
	return o.OrderRepository.Update(order)
}

func (o *OrderService) InsertOrder(order *models.Order) (orderID int64, err error) {
	return o.OrderRepository.Insert(order)
}

func (o *OrderService) GetAllOrder() ([]*models.Order, error) {
	return o.OrderRepository.SelectAll()
}

func (o *OrderService) GetAllOrderInfo() (map[int]map[string]string, error) {
	return o.OrderRepository.SelectAllWithInfo()
}

func (o *OrderService) InsertOrderByMessage(message *models.Message) (orderID int64, err error) {
	order := &models.Order{
		UserId:      message.UserID,
		ProductId:   message.ProductID,
		OrderStatus: models.OrderSuccess,
	}
	return o.InsertOrder(order)
}
