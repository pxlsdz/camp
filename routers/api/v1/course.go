package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreateCourse(c *gin.Context) {
	//TODO: 权限验证

	//参数校验
	var json types.CreateCourseRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.ParamInvalid})
		return
	}

	db := mysql.GetDb()

	//TeacherID is Null
	course := models.Course{
		Name:    json.Name,
		Cap:     json.Cap,
		Deleted: types.Default,
	}
	//如果cap小于0， 返回参数不合法
	if course.Cap < 0 {
		c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.ParamInvalid})
		return
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

	CourseID := c.Query("CourseID")
	id, err := strconv.ParseInt(CourseID, 10, 64)
	//var json types.GetCourseRequest
	//if err := c.ShouldBindJSON(&json); err != nil {
	//	c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.ParamInvalid})
	//	return
	//}
	if err != nil {
		c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.ParamInvalid})
		return
	}

	db := mysql.GetDb()
	var course models.Course
	//判断课程是否存在
	if err := db.Take(&course, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.CourseNotExisted})
			return
		} else {
			c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.UnknownError})
			return
		}
	}
	//判断课程是否删除 --省略删除课程这一步，无状态码

	c.JSON(http.StatusOK, types.GetCourseResponse{
		Code: types.OK,
		Data: types.TCourse{
			CourseID:  strconv.FormatInt(course.ID, 10),
			Name:      course.Name,
			TeacherID: strconv.FormatInt(course.TeacherID, 10),
		},
	})
}
