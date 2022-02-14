package repository

import (
	redis "camp/infrastructure/stores/myRedis"
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// GetBoolStudentById 判断学生是否存在、是否删除
func GetBoolStudentById(id int64) types.ErrNo {
	db := mysql.GetDb()
	var student models.Member
	if err := db.Take(&student, id).Where("user_type = ?", types.Student).Error; err != nil {
		// 判断学生是否存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.StudentNotExisted
		} else {
			return types.UnknownError
		}
	}
	// 判断学生是否已经删除
	if student.Deleted == types.Deleted {
		return types.UserHasDeleted
	}
	return types.OK
}

// GetMemberById 判断用户是否存在、是否删除
func GetMemberById(id int64, member *models.Member) types.ErrNo {
	db := mysql.GetDb()
	if err := db.Take(&member, id).Error; err != nil {
		// 判断用户是否存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.UserNotExisted
		} else {
			return types.UnknownError
		}
	}
	// 判断用户是否已经删除
	if member.Deleted == types.Deleted {
		return types.UserHasDeleted
	}
	return types.OK
}

// GetBoolMemberById 判断课程是否存在并获取课程容量
func GetCapCourseById(id int64, cap *int) types.ErrNo {
	db := mysql.GetDb()
	var course models.Course
	if err := db.Take(&course, id).Error; err != nil {
		// 判断课程是否存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.CourseNotExisted
		} else {
			return types.UnknownError
		}
	}
	*cap = course.Cap
	return types.OK
}

func GetTCourseByID(id int64, tCourse *types.TCourse) types.ErrNo {

	ctx := context.Background()
	client := redis.GetClient()
	db := mysql.GetDb()

	key := fmt.Sprintf(types.TCourseKey, id)

	n, err1 := client.Exists(ctx, key).Result()
	if err1 != nil {
		return types.UnknownError
	} else if n > 0 { //key存在于redis中
		val, err := client.HGetAll(ctx, key).Result()
		// 判断查询是否出错
		if err != nil {
			return types.UnknownError
		} else {
			tCourse.CourseID = fmt.Sprintf("%d", id)
			tCourse.TeacherID = val["TeacherID"]
			tCourse.Name = val["Name"]
		}
	} else { //key不存在于redis中

		var course models.Course
		//先在数据库中查询
		if err := db.Take(&course, id).Error; err != nil {
			// 判断课程是否存在
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.CourseNotExisted
			} else {
				return types.UnknownError
			}
		}

		//存在，写入redis中,设置十分钟过期
		client.HSet(ctx, key, map[string]interface{}{
			"Name":      course.Name,
			"TeacherID": course.TeacherID,
		})
		client.Expire(ctx, key, 600)

		tCourse.CourseID = fmt.Sprintf("%d", id)
		tCourse.TeacherID = fmt.Sprintf("%d", course.TeacherID)
		tCourse.Name = course.Name

	}

	return types.OK
}
