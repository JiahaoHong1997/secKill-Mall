package common

import (
	"github.com/gin-gonic/gin"
)

//设置全局cookie
func GlobalCookie(c *gin.Context, name string, value string, maxAge int)  {
	c.SetCookie(name,value,maxAge,"/","127.0.0.1",false,true)
}
