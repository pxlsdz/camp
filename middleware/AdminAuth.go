package middleware

import (
	"camp/types"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("camp-session")
		if value, ok := user.(types.User); ok == true {
			if value.UserType != types.Admin {
				// 返回错误,未授权
				c.JSON(http.StatusOK, gin.H{"Code": types.PermDenied})
				// 若验证不通过，不再调用后续的函数处理
				c.Abort()
				return
			}
			c.Next()
			return
		}
		// 返回错误,未登录
		c.JSON(http.StatusOK, gin.H{"Code": types.LoginRequired})
		// 若验证不通过，不再调用后续的函数处理
		c.Abort()
		return
	}
}
