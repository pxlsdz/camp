package main

import (
	"camp/infrastructure/mq/rabbitmq"
	"camp/infrastructure/stores/mysql"
	"camp/infrastructure/stores/redis"
	"camp/models"
	"camp/routers"
	"camp/routers/api/v1"
	"camp/types"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 读取配置文件
	initConfig()
	// Read()
	// 连接mysql
	if err := mysql.Init(); err != nil {
		panic(err)
	}

	// 连接redis
	if err := redis.Init(); err != nil {
		panic(err)
	}

	initCourseCap()

	simple := rabbitmq.NewRabbitMQSimple("miaosha")
	simple.ConsumeSimple()

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
	r.Run(":8000")
}

func initCourseCap() {
	var courses []models.Course
	db := mysql.GetDb()
	if err := db.Where("deleted = ?", types.Default).Find(&courses).Error; err != nil {
		panic(err)
	}
	ctx := context.Background()
	cli := redis.GetClient()
	for _, course := range courses {
		if err := cli.Set(ctx, fmt.Sprintf(types.CourseKey, course.ID), course.Cap, -1).Err(); err != nil {
			panic(err)
		}
	}

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
