package models

type Order struct {
	ID          int64 `sql:"ID" secKillSystem:"ID"`
	UserId      int64 `sql:"userID" secKillSystem:"UserId"`
	ProductId   int64 `sql:"productID" secKillSystem:"ProductId"`
	OrderStatus int   `sql:"orderStatus" secKillSystem:"OrderStatus"`
}

const (
	OrderWait    = iota // 0
	OrderSuccess        // 1
	OrderFailed         // 2
)
