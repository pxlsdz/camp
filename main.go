package main

import (
	"camp/infrastructure/mq/rabbitmq"
	"camp/infrastructure/stores/mysql"
	"camp/infrastructure/stores/redis"
	"camp/models"
	"camp/routers"
	"camp/types"
	"context"
	"fmt"
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

	// 1.创建路由
	r := gin.Default()

	routers.RegisterRouter(r)
	// 2.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(":12333")
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
