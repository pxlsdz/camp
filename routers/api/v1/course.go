package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func CreateCourse(c *gin.Context) {
	//TODO: 权限验证

	//参数校验
	var json types.CreateCourseRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, types.CreateMemberResponse{Code: types.ParamInvalid})
		return
	}

	//检验课程是否存在 ? 没有对应返回的状态码
	db := mysql.GetDb()
	var course models.Course
	//find := db.Limit(1).Where("coursename = ?",json.Name).Find(&course)
	//if find.RowsAffected == 1 {
	//	c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.})
	//}

	course = models.Course{
		Name:    json.Name,
		Cap:     json.Cap,
		Deleted: types.Default,
	}

	if err := db.Create(&course).Error; err != nil {
		c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.UnknownError})
		return
	}

	c.JSON(http.StatusOK, types.CreateCourseResponse{
		Code: types.OK,
		Data: struct {
			CourseID string
		}{strconv.FormatInt(course.ID, 10)},
	})
}

func GetCourse(c *gin.Context) {
	//TODO : 权限验证

	var json types.GetCourseRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, types.GetCourseResponse{Code: types.ParamInvalid})
		return
	}

	id, err := strconv.ParseInt(json.CourseID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.ParamInvalid})
		return
	}

	db := mysql.GetDb()
	var course models.Course
	result := db.First(&course, id)
	//判断课程是否存在
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.CourseNotExisted})
		return
	}
	//判断课程是否删除
	//接口设计没有要求删除课程，且无课程已删除的状态码，可否删去？
	if course.Deleted == types.Deleted {
		c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.CourseNotExisted})
		return
	}

	c.JSON(http.StatusOK, types.GetCourseResponse{
		Code: types.OK,
		Data: types.TCourse{
			CourseID: strconv.FormatInt(course.ID, 10),
			Name:     course.Name,
		},
	})
}
