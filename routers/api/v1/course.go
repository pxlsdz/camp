package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/repository"
	"camp/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func CreateCourse(c *gin.Context) {

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

	CourseID := c.Query("CourseID")
	id, err := strconv.ParseInt(CourseID, 10, 64)

	if err != nil {
		c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.ParamInvalid})
		return
	}

	var course types.TCourse
	code := repository.GetTCourseByID(id, &course)
	c.JSON(http.StatusOK, types.GetCourseResponse{
		Code: code,
		Data: course,
	})
}
