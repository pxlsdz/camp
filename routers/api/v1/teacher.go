package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func BindCourse(c *gin.Context) {

	//校验请求参数是否合法
	var json types.BindCourseRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.BindCourseResponse{Code: types.ParamInvalid})
		return
	}

	db := mysql.GetDb()

	//寻找teacher在数据库表中的行，似乎teacherID 默认正确，之后在做改变，似乎可以注释掉
	//var teacher models.Member
	//find_teacher := db.Limit(1).Where("id = ?", json.TeacherID).Find(&teacher)
	////找不到teacher
	//if find_teacher.RowsAffected != 1 {
	//	c.JSON(http.StatusOK, types.BindCourseResponse{Code: types.UserNotExisted})
	//	return
	//}

	var course models.Course
	find_course := db.Limit(1).Where("id = ?", json.CourseID).Find(&course)
	//找不到course
	if find_course.RowsAffected != 1 {
		c.JSON(http.StatusOK, types.BindCourseResponse{Code: types.CourseNotExisted})
		return
	}
	//course被绑定
	if course.TeacherID != 0 {
		c.JSON(http.StatusOK, types.BindCourseResponse{Code: types.CourseHasBound})
		return
	}

	//将传入的string类型转成int64
	teacher_id, _ := strconv.ParseInt(json.TeacherID, 10, 64)

	update := db.Model(&course).Update("teacher_id", teacher_id)
	if update.Error == nil {
		c.JSON(http.StatusOK, types.BindCourseResponse{Code: types.OK})
	} else {
		c.JSON(http.StatusOK, types.BindCourseResponse{Code: types.UnknownError})
	}
	return
}

func UnbindCourse(c *gin.Context) {

	//校验参数是否合法
	var json types.UnbindCourseRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.UnbindCourseResponse{Code: types.ParamInvalid})
		return
	}

	db := mysql.GetDb()

	var teacher models.Member
	find_teacher := db.Limit(1).Where("id = ?", json.TeacherID).Find(&teacher)
	//找不到teacher
	if find_teacher.RowsAffected != 1 {
		c.JSON(http.StatusOK, types.UnbindCourseResponse{Code: types.UserNotExisted})
		return
	}

	var course models.Course
	find_course := db.Limit(1).Where("id = ?", json.CourseID).Find(&course)
	//找不到course
	if find_course.RowsAffected != 1 {
		c.JSON(http.StatusOK, types.UnbindCourseResponse{Code: types.CourseNotExisted})
		return
	}
	//course未被绑定 （待修改）
	if course.TeacherID == 0 {
		c.JSON(http.StatusOK, types.UnbindCourseResponse{Code: types.CourseNotBind})
		return
	}

	//将传入的string类型转成int64
	teacher_id, _ := strconv.ParseInt(json.TeacherID, 10, 64)

	//传入teacherid与表中不同，没有操作权限
	if course.TeacherID != teacher_id {
		c.JSON(http.StatusOK, types.UnbindCourseResponse{Code: types.PermDenied})
		return
	}

	//单字段更新，设为未绑定
	update := db.Model(&course).Update("teacher_id", nil)
	if update.Error == nil {
		c.JSON(http.StatusOK, types.UnbindCourseResponse{Code: types.OK})
	} else {
		c.JSON(http.StatusOK, types.UnbindCourseResponse{Code: types.UnknownError})
	}
	return
}

func GetTeacherCourse(c *gin.Context) {

	//校验参数是否合法
	var json types.GetTeacherCourseRequest

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.GetTeacherCourseResponse{Code: types.ParamInvalid})
		return
	}

	db := mysql.GetDb()

	//寻找teacher
	var teacher models.Member
	find := db.Limit(1).Where("id = ?", json.TeacherID).Find(&teacher)
	//找不到teacher
	if find.RowsAffected != 1 {
		c.JSON(http.StatusOK, types.GetTeacherCourseResponse{Code: types.UserNotExisted})
		return
	}

	//寻找绑定的课程数据
	var courses []models.Course
	if err := db.Where("teacher_id = ?", json.TeacherID).Find(&courses).Error; err != nil {
		c.JSON(http.StatusOK, types.GetTeacherCourseResponse{Code: types.UnknownError})
		return
	}

	//转化成返回得格式
	var tcoures []*types.TCourse
	for _, course := range courses {
		tcoures = append(tcoures, &types.TCourse{
			CourseID:  strconv.FormatInt(course.ID, 10),
			TeacherID: strconv.FormatInt(course.TeacherID, 10),
			Name:      course.Name,
		})
	}

	c.JSON(http.StatusOK, types.GetTeacherCourseResponse{
		Code: types.OK,
		Data: struct{ CourseList []*types.TCourse }{
			CourseList: tcoures,
		},
	})

}
