package models

type Product struct {
	ID           int64  `json:"id" sql:"ID" secKillSystem:"ID"`
	ProductName  string `json:"ProductName" sql:"productName" secKillSystem:"ProductName"`
	ProductNum   int64  `json:"ProductNum" sql:"productNum" secKillSystem:"ProductNum"`
	ProductImage string `json:"ProductImage" sql:"productImage" secKillSystem:"ProductImage"`
	ProductUrl   string `json:"ProductUrl" sql:"productUrl" secKillSystem:"ProductUrl"`
}
