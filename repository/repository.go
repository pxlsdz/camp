package repository

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"errors"
	"gorm.io/gorm"
)

// GetBoolStudentById 判断学生是否存在、是否删除
func GetBoolStudentById(id int64) types.ErrNo {
	db := mysql.GetDb()
	var student models.Member
	if err := db.Take(&student, id).Where("user_type = ?", types.Student).Error; err != nil {
		// 判断学生是否存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.UserNotExisted
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

// GetBoolMemberById 判断用户是否存在、是否删除
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

// GetBoolMemberById 判断用户是否存在、是否删除
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
