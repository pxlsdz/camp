package main

import (
	"camp/infrastructure/goCache"
	"camp/infrastructure/mq/rabbitmq"
	"camp/infrastructure/stores/myRedis"
	"camp/infrastructure/stores/mysql"
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
	gin.SetMode(gin.ReleaseMode)

	// 读取配置文件
	initConfig()
	// Read()
	// 连接mysql
	if err := mysql.Init(); err != nil {
		panic(err)
	}

	// 连接redis
	if err := myRedis.Init(); err != nil {
		panic(err)
	}

	initCourseCap()

	goCache.Init()

	// 消费者1
	simple1 := rabbitmq.NewRabbitMQSimple("miaosha")
	simple1.ConsumeSimple()
	defer simple1.Destory()

	// 消费者2
	simple2 := rabbitmq.NewRabbitMQSimple("miaosha")
	simple2.ConsumeSimple()
	defer simple2.Destory()

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

	r.Run(":80")
}

func initCourseCap() {
	ctx := context.Background()
	cli := myRedis.GetClient()
	db := mysql.GetDb()

	// 导入mysql课程库存记录
	var courses []models.Course
	if err := db.Select("id, cap").Where("deleted = ?", types.Default).Find(&courses).Error; err != nil {
		panic(err)
	}

	pipe := cli.Pipeline()
	for _, course := range courses {
		pipe.Set(ctx, fmt.Sprintf(types.CourseKey, course.ID), course.Cap, types.RedisWriteExpiration)
		// 加入课程布隆过滤器
		pipe.Do(ctx, "BF.ADD", types.BCourseKey, course.ID)
	}

	// 导入mysql选课记录
	var studentCourses []models.StudentCourse
	if err := db.Select("student_id, course_id").Find(&studentCourses).Error; err != nil {
		panic(err)
	}

	s2c := make(map[int64][]interface{})
	for _, studentCourses := range studentCourses {
		s2c[studentCourses.StudentID] = append(s2c[studentCourses.StudentID], studentCourses.CourseID)
		// 加入选课记录布隆过滤器
		pipe.Do(ctx, "BF.ADD", types.BStudentHasCourseKey, fmt.Sprintf(types.StudentIDCourseIDKey, studentCourses.StudentID, studentCourses.CourseID))
	}
	for k, v := range s2c {
		pipe.SAdd(ctx, fmt.Sprintf(types.StudentHasCourseKey, k), v)
	}

	// 导入mysql学生成员记录
	var studentIDs []int64
	if err := db.Select("id").Where("user_type = ? AND deleted = ?", types.Student, types.Default).Model(&models.Member{}).Scan(&studentIDs).Error; err != nil {
		panic(err)
	}
	t := make([]interface{}, len(studentIDs))
	for i, v := range studentIDs {
		t[i] = v
		// 加入学生布隆过滤器
		pipe.Do(ctx, "BF.ADD", types.BStudentKey, v)
	}
	if studentIDs != nil && len(studentIDs) > 0 {
		pipe.SAdd(ctx, types.StudentKey, t)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		panic(err)
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
