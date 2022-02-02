package routers

import (
	"github.com/gin-gonic/gin"
)
import "camp/routers/api/v1"

func RegisterRouter(r *gin.Engine) {
	g := r.Group("/api/v1")
	// 成员管理
	g.POST("/member/create", v1.CreateMember)
	g.GET("/member", v1.GetMember)
	g.GET("/member/list", v1.GetMemberList)
	g.POST("/member/update", v1.UpdateMember)
	g.POST("/member/delete", v1.DeleteMember)

	// 登录

	g.POST("/auth/login")
	g.POST("/auth/logout")
	g.GET("/auth/whoami")

	// 排课
	g.POST("/course/create")
	g.GET("/course/get")

	g.POST("/teacher/bind_course", v1.BindCourse)
	g.POST("/teacher/unbind_course", v1.UnbindCourse)
	g.GET("/teacher/get_course", v1.GetTeacherCourse)

	g.POST("/course/schedule", v1.ScheduleCourse)
	// g.GET("/course/schedule_test", v1.ScheduleCourseTest)

	// 抢课
	g.POST("/student/book_course")
	g.GET("/student/course")

}
