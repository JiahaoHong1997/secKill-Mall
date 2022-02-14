package models

type UserInfo struct {
	ID         int64  `json:"id" form:"ID" sql:"ID"`
	UserName   string `json:"userName" form:"userName" sql:"userName"`
	UserGender int    `json:"userGender" form:"userGender" sql:"userGender"`
	LoginIp    string `json:"loginIp" form:"loginIp" sql:"loginIp"`
}

const (
	female = 0
	male   = 1
)

// TODO:
// 1.新建存放消息的结构体UserInfo，用户ID，用户名称，用户性别，用户登陆IP；
// 2.登陆成功后创建结构体UserInfo的实例存放用户数据；
// 3.转化UserInfo结构体到json，并且对json进行加密写入到cookie中；
// 4.获取cookie中加密字符串解密，并且把解密数据映射到UserInfo中；
