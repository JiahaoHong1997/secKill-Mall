package models

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

type MessageCache struct {
	ID           int64
	ProductName  string
	ProductNum   int64
	ProductImage string
	ProductUrl   string
	IsDelete     bool
}

func NewMessageCache(id int64, name string, num int64, image string, url string, isDelete bool) *MessageCache {
	return &MessageCache{
		ID:           id,
		ProductName:  name,
		ProductNum:   num,
		ProductImage: image,
		ProductUrl:   url,
		IsDelete:     isDelete,
	}
}
