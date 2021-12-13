package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v, _ := c.Cookie("uid"); v == "" {
			c.Redirect(http.StatusMovedPermanently, "/user/login")
		}
	}
}
