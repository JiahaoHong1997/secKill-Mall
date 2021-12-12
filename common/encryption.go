package common

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func ValidatePassword(userPassWord string, hashed string) (bool, error) { // 解密
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassWord)); err != nil {
		return false, errors.New("ValidatePassword: passWord is not correct")
	}
	return true, nil
}

func GeneratePassWord(userPassword string) ([]byte, error) { // 加密
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}
