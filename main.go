package main

import (
	"camp/infrastructure/mq/rabbitmq"
	"camp/infrastructure/stores/mysql"
	"camp/infrastructure/stores/redis"
	"camp/models"
	"camp/routers"
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
	defer simple.Destory()

	//把user这个接头体注册进来，后面跨路由才可以获取到user数据
	gob.Register(types.User{})

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
	ctx := context.Background()
	cli := redis.GetClient()
	db := mysql.GetDb()

	// 导入mysql课程库存记录
	var courses []models.Course
	if err := db.Where("deleted = ?", types.Default).Find(&courses).Error; err != nil {
		panic(err)
	}
	for _, course := range courses {
		if err := cli.Set(ctx, fmt.Sprintf(types.CourseKey, course.ID), course.Cap, -1).Err(); err != nil {
			panic(err)
		}
	}

	// 导入mysql选课记录
	var studentCourses []models.StudentCourse
	if err := db.Find(&studentCourses).Error; err != nil {
		panic(err)
	}
	for _, studentCourses := range studentCourses {
		if err := cli.Set(ctx, fmt.Sprintf(types.StudentHasCourseKey, studentCourses.StudentID, studentCourses.CourseID), 1, -1).Err(); err != nil {
			panic(err)
		}
	}

	// 导入mysql学生成员记录
	var students []models.Member
	if err := db.Where("user_type", types.Student).Find(&students).Error; err != nil {
		panic(err)
	}
	for _, student := range students {
		if err := cli.Set(ctx, fmt.Sprintf(types.StudentKey, student.ID), 1, -1).Err(); err != nil {
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
