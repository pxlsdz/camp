package routers

import (
	"camp/middleware"
	"camp/routers/api/v1"
	"github.com/gin-gonic/gin"
	"time"
)

func RegisterRouter(r *gin.Engine) {
	// 测试并发
	//r.GET("/ping", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "pong",
	//	})
	//})

	// 登录
	auth := r.Group("/api/v1/auth")

	auth.POST("/login", v1.Login)
	auth.POST("/logout", v1.Logout)
	auth.GET("/whoami", middleware.LoginAuth(), v1.Whoami)

	// 成员管理
	member := r.Group("/api/v1/member")

	member.POST("/create", middleware.AdminAuth(), v1.CreateMember)
	member.GET("", v1.GetMember)
	member.GET("/list", v1.GetMemberList)
	member.POST("/update", v1.UpdateMember)
	member.POST("/delete", v1.DeleteMember)

	teacher := r.Group("/api/v1/teacher")
	teacher.POST("/bind_course", v1.BindCourse)
	teacher.POST("/unbind_course", v1.UnbindCourse)
	teacher.GET("/get_course", v1.GetTeacherCourse)

	course := r.Group("/api/v1/course")
	// 排课
	course.POST("/create", v1.CreateCourse)
	course.GET("/get", v1.GetCourse)

	course.POST("/schedule", v1.ScheduleCourse)
	// g.GET("/course/schedule_test", v1.ScheduleCourseTest)

	student := r.Group("/api/v1/student")
	// 抢课
	student.POST("/book_course", middleware.RateLimitMiddleware(1*time.Second, 200000, 100000), v1.BookCourse)
	student.GET("/course", middleware.RateLimitMiddleware(1*time.Second, 200000, 100000), v1.GetStudentCourse)

	//// 测试MQ
	//r.GET("/mq", func(c *gin.Context) {
	//	StudentID := c.Query("StudentID")
	//	CourseID := c.Query("CourseID")
	//	studentID, err := strconv.ParseInt(StudentID, 10, 64)
	//	if err != nil {
	//		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.ParamInvalid})
	//		return
	//	}
	//	courseID, err := strconv.ParseInt(CourseID, 10, 64)
	//	if err != nil {
	//		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.ParamInvalid})
	//		return
	//	}
	//	// 消息队列减少课程数据库的库存以及创建数据库表
	//	//创建消息体
	//
	//	studentCourse := models.StudentCourse{
	//		StudentID: studentID,
	//		CourseID:  courseID,
	//	}
	//	//类型转化
	//	byteMessage, _ := json.Marshal(studentCourse)
	//
	//	rabbitMQ := rabbitmq.GetRabbitMQ()
	//	rabbitMQ.PublishSimple(string(byteMessage))
	//
	//	c.JSON(200, gin.H{
	//		"message": "pong",
	//	})
	//})

}
