package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func AuthMiddleWare() gin.HandlerFunc {

	return func(c *gin.Context) {
		// 获取客户端cookie并校验
		if cookie, err := c.Cookie("camp-session"); err == nil && cookie != "" {
			c.Next()
			return
		}
		// 返回错误,未授权
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unAuthorized"})
		// 若验证不通过，不再调用后续的函数处理
		c.Abort()
		return
	}
}

func Login(c *gin.Context) {

	fmt.Println("login")
	var json types.LoginRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, "json解析失败")
		return
	}

	username := json.Username
	password := json.Password

	db := mysql.GetDb()
	var member models.Member
	err := db.Where("username = ?", username).First(&member)
	//err := db.Select(&member, "select username from member where username='guoguo'")
	//result := db.Select("select username from member where username=?", username)
	if err.Error != nil || member.Password != password {
		c.JSON(http.StatusOK, types.LoginResponse{Code: types.WrongPassword})
		return
	}

	// 登录成功，设置cookie
	c.SetCookie("camp-session", strconv.FormatInt(member.ID, 10), 120, "/",
		"127.0.0.1", false, true)

	c.JSON(http.StatusOK, types.LoginResponse{
		Code: types.OK,
		Data: struct {
			UserID string
		}{strconv.FormatInt(member.ID, 10)},
	})
}

func Logout(c *gin.Context) {
	//建立同名cookie进行覆盖
	c.SetCookie("camp-session", "value_cookie", -1, "/",
		"127.0.0.1", false, true)
	// 返回信息
	c.JSON(http.StatusOK, types.LogoutResponse{
		Code: types.OK,
	})
}

func Whoami(c *gin.Context) {

	cookie_uid, _ := c.Cookie("camp-session")
	db := mysql.GetDb()
	var member models.Member

	db.Where("id = ?", cookie_uid).First(&member)
	c.JSON(http.StatusOK, types.WhoAmIResponse{
		Code: types.OK,
		Data: types.TMember{
			UserID:   strconv.FormatInt(member.ID, 10),
			Nickname: member.Nickname,
			Username: member.Username,
			UserType: member.UserType,
		},
	})
}
