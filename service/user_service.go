package service

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"seckill/datamodels"
	"seckill/repositories"
)

type IUserService interface {
	IsPwdSuccess(userName string, pwd string) (user *datamodels.User, isOk bool, err error)
	AddUser(user *datamodels.User) (int64, error)

}

type UserService struct {
	UserRepository repositories.IUserRepository
}

func NewUserService(repository repositories.IUserRepository) IUserService {
	return &UserService{
		UserRepository: repository,
	}
}

func (u *UserService) IsPwdSuccess(userName string, pwd string) (*datamodels.User, bool, error) {
	user, err := u.UserRepository.Select(userName)
	if err != nil {
		return &datamodels.User{}, false, errors.WithMessage(err, "no such user")
	}
	isOk, _ := ValidatePassword(pwd, user.HashPassword)
	if !isOk {
		return &datamodels.User{}, false, errors.New("password false")
	}
	return user, true, nil
}

func ValidatePassword(userPassWord string, hashed string) (bool, error) {	// 解密
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassWord)); err != nil {
		return false, errors.New("ValidatePassword: passWord is not correct")
	}
	return true, nil
}

func GeneratePassWord(userPassword string) ([]byte, error) {	// 加密
	return bcrypt.GenerateFromPassword([]byte(userPassword),bcrypt.DefaultCost)
}

func (u *UserService) AddUser(user *datamodels.User) (int64, error) {
	pwdByte, err := GeneratePassWord(user.HashPassword)
	if err != nil {
		return 0, errors.Wrap(err, "encryption failed")
	}
	user.HashPassword = string(pwdByte)
	return u.UserRepository.Insert(user)
}

