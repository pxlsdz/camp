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

	//参数校验
	var jsonRequest types.GetStudentCourseRequest
	if err := c.ShouldBindJSON(&jsonRequest); err != nil {
		c.JSON(http.StatusBadRequest, types.GetStudentCourseResponse{Code: types.ParamInvalid})
		return
	}

	db := mysql.GetDb()
	studentID, err := strconv.ParseInt(jsonRequest.StudentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.GetStudentCourseResponse{Code: types.ParamInvalid})
		return
	}

	//判断学生是否存在
	//是否需要判断是否是学生，教师，管理员？此处未判断
	var student models.Member
	result := db.First(&student, studentID)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.StudentNotExisted})
		return
	}

	var courseList []types.TCourse
	//关联表
	type Result struct {
		ID        int64
		Name      string
		TeacherID int64
	}
	var results []Result
	db.Raw("SELECT c.id, c.name, c.teacher_id FROM course c WHERE c.id IN (SELECT sc.course_id FROM student_course sc WHERE sc.student_id = ?)", jsonRequest.StudentID).Scan(&results)
	if len(results) == 0 {
		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.StudentHasNoCourse})
		return
	}
	for _, result := range results {
		courseList = append(courseList, types.TCourse{
			CourseID:  strconv.FormatInt(result.ID, 10),
			Name:      result.Name,
			TeacherID: strconv.FormatInt(result.TeacherID, 10),
		})
	}
	c.JSON(http.StatusOK, types.GetStudentCourseResponse{
		Code: types.OK,
		Data: struct{ CourseList []types.TCourse }{CourseList: courseList},
	})
}

var localCapOverMap map[int64]bool

func BookCourse(c *gin.Context) {

	// 参数校验
	var requestJson types.BookCourseRequest
	if err := c.ShouldBindJSON(&requestJson); err != nil {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.ParamInvalid})
		return
	}

	courseId, err := strconv.ParseInt(requestJson.CourseID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.ParamInvalid})
		return
	}

	studentID, err := strconv.ParseInt(requestJson.StudentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.ParamInvalid})
		return
	}

	_, ok := localCapOverMap[courseId]
	if ok {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotAvailable})
		return
	}

	ctx := context.Background()
	cli := redis.GetClient()
	// redis lua脚本实现检验该学生是否已经有该课和课程数量是否足够
	// 缓存设计待讨论
	// 学生是否存在与商品是否存在 是否和减少库存是一个原子性质操作？
	res, err := cli.EvalSha(ctx, redis.LuaHash, []string{fmt.Sprintf(types.StudentHasCourseKey, studentID, courseId), fmt.Sprintf(types.CourseKey, courseId), fmt.Sprintf(types.StudentKey, studentID)}).Result()

	if err != nil || res == int64(-1) {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	}
	if res == int64(4) {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentNotExisted})
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
		localCapOverMap[courseId] = true
		return
	}
	// 消息队列减少课程数据库的库存以及创建数据库表
	//创建消息体

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
