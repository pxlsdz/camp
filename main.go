package main

// 声明 main 包，表明当前是一个可执行程序
import (
	"camp/infrastructure/stores/mysql"
	"camp/routers"
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

import "camp/routers/api/v1"

// main函数，是程序执行的入口
func main() {
	initConfig()

	// 连接mysql
	if err := mysql.Init(); err != nil {
		panic(err)
	}
	//把user这个接头体注册进来，后面跨路由才可以获取到user数据
	gob.Register(v1.User{})
	// 1.创建路由
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	//路由上加入session中间件
	r.Use(sessions.Sessions("camp-session", store))

	routers.RegisterRouter(r)
	// 2.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(":6000")
}

// viper工具读取配置文件
func initConfig() {

	viper.AddConfigPath("conf")
	viper.SetConfigName("config")

	//viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		panic(err)
	}
}
