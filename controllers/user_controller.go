package controllers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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
	log.Println(strings.Split(c.Request.Header.Get("X-Forwarded-For"), ","))
	log.Println(ip)
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
	var (
		userName = c.PostForm("userName")
		password = c.PostForm("password")
	)

	user, isOk, _ := userService.IsPwdSuccess(userName, password)
	if !isOk {
		c.Redirect(http.StatusMovedPermanently, "/user/login")
		return
	}
	session := sessions.Default(c)
	session.Set(userName, user.HashPassword)
	session.Save()
	common.GlobalCookie(c, "uid", strconv.FormatInt(user.ID, 10), 30*60)
	data, _ := c.Cookie("uid")
	log.Println(data)
	c.Redirect(http.StatusMovedPermanently, "/user/register")
}
