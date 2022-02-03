package routers

import (
	"camp/routers/api/v1"
	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.Engine) {
	g := r.Group("/api/v1")
	// 成员管理
	g.POST("/member/create", v1.CreateMember)
	g.GET("/member", v1.GetMember)
	g.GET("/member/list", v1.GetMemberList)
	g.POST("/member/update", v1.UpdateMember)
	g.POST("/member/delete", v1.DeleteMember)

	// 登录
	g.POST("/auth/login", v1.Login)
	g.POST("/auth/logout", v1.Logout)
	g.GET("/auth/whoami", v1.AuthMiddleWare(), v1.Whoami)

	// 排课
	g.POST("/course/create", v1.CreateCourse)
	g.GET("/course/get", v1.GetCourse)

	g.POST("/teacher/bind_course", v1.BindCourse)
	g.POST("/teacher/unbind_course", v1.UnbindCourse)
	g.GET("/teacher/get_course", v1.GetTeacherCourse)

	g.POST("/course/schedule", v1.ScheduleCourse)
	// g.GET("/course/schedule_test", v1.ScheduleCourseTest)

	// 抢课
	g.POST("/student/book_course", v1.BookCourse)
	g.GET("/student/course", v1.GetStudentCourse)

}
