package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"seckill/common"
	"seckill/datamodels"
	"seckill/repositories"
	"seckill/service"
	"strconv"
	"strings"
)

var userRepository repositories.IUserRepository
var userService service.IUserService

func init() {
	db := common.DBConn()
	userRepository = repositories.NewUserRepository("user", db)
	userService = service.NewUserService(userRepository)
}

func GetRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.tmpl", nil)
}

func GetLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", nil)
}

func GetError(c *gin.Context) {
	c.HTML(http.StatusOK, "error.tmpl", gin.H{
		"Message": "访问的页面出错",
	})
}

func PostRegister(c *gin.Context) {
	var (
		nikName  = c.PostForm("nickName")
		userName = c.PostForm("userName")
		password = c.PostForm("password")
	)

	var ip string
	for _, ip = range strings.Split(c.Request.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			break
		}
	}

	user := &datamodels.User{
		UserName:     userName,
		NickName:     nikName,
		HashPassword: password,
		UserIp:       ip,
	}
	_, err := userService.AddUser(user)
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/user/error")
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/user/login")
}

func PostLogin(c *gin.Context) {
	// 1.获取表单信息
	var (
		userName = c.PostForm("userName")
		password = c.PostForm("password")
	)

	// 2.验证账号密码
	user, isOk, _ := userService.IsPwdSuccess(userName, password)
	if !isOk {
		c.Redirect(http.StatusMovedPermanently, "/user/login")
		return
	}
	// 3.写入用户ID到 cookie 中
	c.SetCookie("uid", strconv.FormatInt(user.ID, 10), 30*60, "/", "127.0.0.1", false, true)
	uidByte := []byte(strconv.FormatInt(user.ID, 10))
	uidString, err := common.EnPwdCode(uidByte)
	if err != nil {
		log.Printf("origin error: %T, %v", errors.Cause(err), errors.Cause(err))
		log.Printf("stack trace: %+v", err)
	}
	c.SetCookie("sign", uidString, 30*60, "/", "127.0.0.1", false, true)

	c.Redirect(http.StatusMovedPermanently, "/product/")
}
