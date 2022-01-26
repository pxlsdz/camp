package main

import (
	"camp/infrastructure/stores/mysql"
	"camp/routers"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	initConfig()

	// 连接mysql
	if err := mysql.Init(); err != nil {
		panic(err)
	}

	// 1.创建路由
	r := gin.Default()

	routers.RegisterRouter(r)
	// 2.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(":8000")
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
