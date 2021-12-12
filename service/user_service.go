package service

import (
	"github.com/pkg/errors"
	"seckill/common"
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
	isOk, _ := common.ValidatePassword(pwd, user.HashPassword)
	if !isOk {
		return &datamodels.User{}, false, errors.New("password false")
	}
	return user, true, nil
}

func (u *UserService) AddUser(user *datamodels.User) (int64, error) {
	pwdByte, err := common.GeneratePassWord(user.HashPassword)
	if err != nil {
		return 0, errors.Wrap(err, "encryption failed")
	}
	user.HashPassword = string(pwdByte)
	return u.UserRepository.Insert(user)
}
