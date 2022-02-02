package v1

import (
	"camp/infrastructure/mq/rabbitmq"
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

	courseId, _ := strconv.ParseInt(requestJson.CourseID, 10, 64)
	res, err := cli.EvalSha(ctx, redis.LuaHash, []string{fmt.Sprintf(types.StudentHasCourseKey, requestJson.StudentID, requestJson.CourseID), fmt.Sprintf(types.CourseKey, courseId)}).Result()
	if err != nil || res == -1 {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.UnknownError})
		return
	}
	if res == 3 {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.StudentHasCourse})
		return
	}
	if res == 2 {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.CourseNotExisted})
		return
	}
	if res == 0 {
		c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.CourseNotAvailable})
		return
	}

	// 消息队列下单
	//创建消息体
	studentID, _ := strconv.ParseInt(requestJson.StudentID, 10, 64)
	message := models.Message{
		StudentID: studentID,
		CourseId:  courseId,
	}
	//类型转化
	byteMessage, _ := json.Marshal(message)
	//if err != nil {
	//	//p.Ctx.Application().Logger().Debug(err)
	//}
	rabbitMQ := rabbitmq.GetRabbitMQ()
	err = rabbitMQ.PublishSimple(string(byteMessage))
	if err != nil {
		//p.Ctx.Application().Logger().Debug(err)
	}

	c.JSON(http.StatusBadRequest, types.BookCourseResponse{Code: types.OK})

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
