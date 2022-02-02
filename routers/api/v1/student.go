package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetStudentCourse(c *gin.Context) {
	//TODO: 登陆验证和权限验证

	//参数校验
	var json types.GetStudentCourseRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, types.GetStudentCourseResponse{Code: types.ParamInvalid})
		return
	}

	//这里用student_courses or studentCourses
	var studentCourses []models.StudentCourse
	db := mysql.GetDb()
	if err := db.Where("student_id = ?", json.StudentID).Find(&studentCourses).Error; err != nil {
		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.UnknownError})
		return
	}

	//表的关联有没有更好的方法？
	var courseList []types.TCourse
	for _, studentCourse := range studentCourses {
		var course models.Course
		//课程没有删除的逻辑，是否还需要判断课程是否存在？
		_ = db.Find(&course, studentCourse.CourseID)
		courseList = append(courseList, types.TCourse{
			CourseID:  strconv.FormatInt(course.ID, 10),
			Name:      course.Name,
			TeacherID: strconv.FormatInt(course.TeacherID, 10),
		})
	}

	c.JSON(http.StatusOK, types.GetStudentCourseResponse{
		Code: types.OK,
		Data: struct{ CourseList []types.TCourse }{CourseList: courseList},
	})

}
