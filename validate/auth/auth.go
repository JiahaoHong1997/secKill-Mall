package auth

import (
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"seckill/common/encrypt"
)

// 统一验证拦截器，每个接口都需要提前验证
func Auth(w http.ResponseWriter, r *http.Request) error {
	err := CheckUserInfo(r)
	if err != nil {
		return err
	}
	return nil
}

// 身份校验函数
func CheckUserInfo(r *http.Request) error {
	// 1.获取uid cookie
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return errors.New("uid got failed")
	}

	// 2.获取用户加密串
	signCookie, err := r.Cookie("sign")
	if err != nil {
		return errors.New("sign got failed")
	}

	val, _ := url.QueryUnescape(signCookie.Value)
	signByte, err := encrypt.DePwdCode(val)
	if err != nil {
		return errors.New("加密串已被串改")
	}

	if checkInfo(uidCookie.Value, string(signByte)) {
		return nil
	}
	return errors.New("身份校验失败！")
}

func checkInfo(checkStr string, signStr string) bool {
	if checkStr == signStr {
		return true
	}
	return false
}
