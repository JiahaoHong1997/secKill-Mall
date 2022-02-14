package service

import (
	"github.com/pkg/errors"
	"seckill/common/encrypt"
	"seckill/dao"
	"seckill/models"
)

type IUserService interface {
	IsPwdSuccess(userName string, pwd string) (user *models.User, isOk bool, err error)
	AddUser(user *models.User) (int64, error)
}

type UserService struct {
	UserRepository dao.IUserRepository
}

func NewUserService(repository dao.IUserRepository) IUserService {
	return &UserService{
		UserRepository: repository,
	}
}

func (u *UserService) IsPwdSuccess(userName string, pwd string) (*models.User, bool, error) {
	user, err := u.UserRepository.Select(userName)
	if err != nil {
		return &models.User{}, false, errors.WithMessage(err, "no such user")
	}
	isOk, _ := encrypt.ValidatePassword(pwd, user.HashPassword)
	if !isOk {
		return &models.User{}, false, errors.New("password false")
	}
	return user, true, nil
}

func (u *UserService) AddUser(user *models.User) (int64, error) {
	pwdByte, err := encrypt.GeneratePassWord(user.HashPassword)
	if err != nil {
		return 0, errors.Wrap(err, "encryption failed")
	}
	user.HashPassword = string(pwdByte)
	return u.UserRepository.Insert(user)
}
