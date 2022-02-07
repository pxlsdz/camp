package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Login(c *gin.Context) {

	var jsonLogin types.LoginRequest
	if err := c.ShouldBindJSON(&jsonLogin); err != nil {
		c.JSON(http.StatusBadRequest, types.LoginResponse{Code: types.ParamInvalid})
		return
	}

	username := jsonLogin.Username
	password := jsonLogin.Password

	db := mysql.GetDb()
	var member models.Member

	//err := db.Where("username = ?", username).First(&member)
	err := db.Limit(1).Where("username = ?", username).Find(&member)

	if err.Error != nil || member.Password != password || member.Deleted == types.Deleted {
		c.JSON(http.StatusOK, types.LoginResponse{Code: types.WrongPassword})
		return
	}

	// 创建session
	session := sessions.Default(c)
	//注意类型的转换
	loginUser := types.User{UserId: member.ID, UserType: types.UserType(int(member.UserType))}
	session.Set("camp-session", loginUser)
	// 设置session的参数
	options := sessions.Options{}
	options.Path = "/"
	// domain：域名，本地调试，127.0.0.1;正式,180.184.74.13
	options.Domain = "180.184.74.13"
	//options.Domain = "127.0.0.1"
	//maxAge: x<0,立即删除cookie; x=0,无限时间; x>0,x秒之后过期
	options.MaxAge = 0
	session.Options(options)
	session.Save()

	c.JSON(http.StatusOK, types.LoginResponse{
		Code: types.OK,
		Data: struct {
			UserID string
		}{strconv.FormatInt(member.ID, 10)},
	})
	return
}

func Logout(c *gin.Context) {
	//建立同名cookie进行覆盖
	session := sessions.Default(c)
	user := session.Get("camp-session")
	if _, ok := user.(types.User); ok == true {
		session.Set("camp-session", user)
		// 设置session的参数
		options := sessions.Options{}
		options.Path = "/"
		// domain：域名，本地调试，127.0.0.1;正式,180.184.74.13
		options.Domain = "180.184.74.13"
		//options.Domain = "127.0.0.1"
		//maxAge: x<0,立即删除cookie; x=0,无限时间; x>0,x秒之后过期
		options.MaxAge = -1
		session.Options(options)
		session.Save()
		// 返回信息
		c.JSON(http.StatusOK, types.LogoutResponse{Code: types.OK})
		return
	}

	// 返回错误,未授权
	c.JSON(http.StatusUnauthorized, types.LogoutResponse{Code: types.LoginRequired})
	// 若验证不通过，不再调用后续的函数处理
	c.Abort()
	return
}

func Whoami(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("camp-session")
	if value, ok := user.(types.User); ok == true {
		sessionUid := value.UserId
		db := mysql.GetDb()
		var member models.Member
		//db.Where("id = ?", sessionUid).First(&member)
		db.Limit(1).Where("id = ?", sessionUid).Find(&member)
		c.JSON(http.StatusOK, types.WhoAmIResponse{
			Code: types.OK,
			Data: types.TMember{
				UserID:   strconv.FormatInt(member.ID, 10),
				Nickname: member.Nickname,
				Username: member.Username,
				UserType: member.UserType,
			},
		})
		return
	}

	// 返回错误,未授权
	c.JSON(http.StatusUnauthorized, types.WhoAmIResponse{Code: types.LoginRequired})
	// 若验证不通过，不再调用后续的函数处理
	c.Abort()
	return
}
