package datamodels

// 定义简单消息体
type Message struct {
	ProductID int64
	UserID    int64
}

// 创建消息体
func NewMessage(userId int64, productId int64) *Message {
	return &Message{
		ProductID: productId,
		UserID:    userId,
	}
}
