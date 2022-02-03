package v1

import (
	"camp/infrastructure/mq/rabbitmq"
	"camp/infrastructure/stores/mysql"
	"camp/infrastructure/stores/redis"
	"camp/models"
	"camp/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetStudentCourse(c *gin.Context) {
	//TODO: 登陆验证和权限验证

	//参数校验
	var jsonRequest types.GetStudentCourseRequest
	if err := c.ShouldBindJSON(&jsonRequest); err != nil {
		c.JSON(http.StatusBadRequest, types.GetStudentCourseResponse{Code: types.ParamInvalid})
		return
	}

	//这里用student_courses or studentCourses
	var studentCourses []models.StudentCourse
	db := mysql.GetDb()
	if err := db.Where("student_id = ?", jsonRequest.StudentID).Find(&studentCourses).Error; err != nil {
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

func BookCourse(c *gin.Context) {
	//TODO:登录验证和权限认证
	//TODO:判断学生id和登录id怎么区分

	// 参数校验
	var requestJson types.BookCourseRequest
	if err := c.ShouldBindJSON(&requestJson); err != nil {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.ParamInvalid})
		return
	}

	ctx := context.Background()
	cli := redis.GetClient()
	// redis lua脚本实现检验是否已经有该课和课程数量是否足够
	// 缓存设计待讨论
	courseId, _ := strconv.ParseInt(requestJson.CourseID, 10, 64)
	res, err := cli.EvalSha(ctx, redis.LuaHash, []string{fmt.Sprintf(types.StudentHasCourseKey, requestJson.StudentID, requestJson.CourseID), fmt.Sprintf(types.CourseKey, courseId)}).Result()

	if err != nil || res == int64(-1) {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	}
	if res == int64(3) {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentHasCourse})
		return
	}
	if res == int64(2) {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotExisted})
		return
	}
	if res == int64(0) {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotAvailable})
		return
	}
	// 消息队列减少课程数据库的库存以及创建数据库表
	//创建消息体
	studentID, _ := strconv.ParseInt(requestJson.StudentID, 10, 64)
	studentCourse := models.StudentCourse{
		StudentID: studentID,
		CourseID:  courseId,
	}
	//类型转化
	byteMessage, _ := json.Marshal(studentCourse)
	//if err != nil {
	//
	//}
	rabbitMQ := rabbitmq.GetRabbitMQ()
	err = rabbitMQ.PublishSimple(string(byteMessage))
	//if err != nil {
	//}

	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.OK})

	return

	//// 加锁
	//if res, err := cli.SetNX(ctx, fmt.Sprintf("%sl%s", requestJson.StudentID, requestJson.CourseID), time.Now().Unix(), time.Minute).Result(); err != nil || !res {
	//	c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.UnknownError})
	//	return
	//}

	//// 判断学生是否已经拥有该课程
	//if _, err := cli.Get(ctx, fmt.Sprintf(types.StudentHasCourseKey, requestJson.StudentID, requestJson.CourseID)).Result(); err != redis.Nil {
	//	if err == nil {
	//		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.StudentHasCourse})
	//	} else {
	//		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.UnknownError})
	//	}
	//	return
	//}
	//
	//// 预减库存
	//courseId, _ := strconv.ParseInt(requestJson.CourseID, 10, 64)
	//stock, err := cli.Decr(ctx, fmt.Sprintf(types.CourseKey, courseId)).Result()
	//if err != nil {
	//	if err == redis.Nil {
	//		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.CourseNotExisted})
	//	} else {
	//		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.UnknownError})
	//	}
	//	return
	//}
	//if stock < 0 {
	//	c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.CourseNotAvailable})
	//}

}
