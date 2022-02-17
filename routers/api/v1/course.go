package v1

import (
	"camp/infrastructure/stores/myRedis"
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/repository"
	"camp/types"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

	//如果cap小于0， 返回参数不合法
	if json.Cap < 0 {
		c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.ParamInvalid})
		return
	}
	var course models.Course
	var count int64
	if err := db.Model(&course).Where("name = ?", json.Name).Limit(1).Count(&count).Error; err != nil {
		c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.UnknownError})
		return
	}
	if count == 1 {
		c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.UnknownError})
		return
	}

	course = models.Course{
		Name:    json.Name,
		Cap:     json.Cap,
		Deleted: types.Default,
	}

	if err := db.Create(&course).Error; err != nil {
		c.JSON(http.StatusOK, types.CreateCourseResponse{Code: types.UnknownError})
		return
	}

	cli := myRedis.GetClient()
	ctx := context.Background()
	cli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, fmt.Sprintf(types.CourseKey, course.ID), course.Cap, types.RedisWriteExpiration)
		pipe.Do(ctx, "BF.ADD", types.BCourseKey, course.ID)
		return nil
	})

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
