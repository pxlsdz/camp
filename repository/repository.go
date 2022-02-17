package repository

import (
	redis "camp/infrastructure/stores/myRedis"
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
	"context"
	"errors"
	"fmt"
	redis3 "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"strconv"
	"time"
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
	//转化为redis中对应的key
	key := fmt.Sprintf(types.TCourseKey, id)

	//从redis中获取hmap，会返回一个map集合
	val, err := client.HGetAll(ctx, key).Result()
	// 判断查询是否出错
	if err != nil {
		return types.UnknownError
	} else if len(val) != 0 { //如果集合里面有元素，则该键存在于redis中
		tCourse.CourseID = fmt.Sprintf("%d", id)
		tCourse.TeacherID = val["TeacherID"]
		tCourse.Name = val["Name"]
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
		client.Expire(ctx, key, 600*time.Second)

		tCourse.CourseID = fmt.Sprintf("%d", id)
		tCourse.TeacherID = fmt.Sprintf("%d", course.TeacherID)
		tCourse.Name = course.Name
	}
	return types.OK
}

func GetTCourseByIDs(ids []int64) (tCourses []types.TCourse, code types.ErrNo) {
	ctx := context.Background()
	client := redis.GetClient()
	db := mysql.GetDb()
	//转化为redis中对应的keys
	length := len(ids)
	//初始化
	tCourses = make([]types.TCourse, 0, length)
	keys := make([]string, length)
	//转化id位redis对应的key
	for i, id := range ids {
		keys[i] = fmt.Sprintf(types.TCourseKey, id)
	}

	//初始化redis中不存在的id列表
	idsNotExistRedis := make([]int64, 0)

	//管道
	pipeline := client.Pipeline()

	for _, key := range keys {
		//从redis中获取hmap，会返回一个map集合
		pipeline.HGetAll(ctx, key)
	}

	//执行
	exec, err := pipeline.Exec(ctx)

	if err != nil {
		if err != redis.Nil {
			code = types.UnknownError
			return
		}
		// 注意这里如果某一次获取时出错（常见的redis.Nil），返回的err即不为空
		// 如果需要处理redis.Nil为默认值，此处不能直接return
	}

	for index, cmdRes := range exec {
		// 此处断言类型为在for循环内执行的命令返回的类型,上面HGet返回的即为*redis2.StringStringMapCmd
		// 处理方式和直接调用同样处理即可
		cmd, ok := cmdRes.(*redis3.StringStringMapCmd)
		if ok {
			val, err := cmd.Result()
			if err != nil {
				code = types.UnknownError
				return
			}
			//判断是否返回的为空，为空map则不存在
			if len(val) != 0 {
				tCourse := types.TCourse{
					fmt.Sprintf("%d", ids[index]),
					val["Name"],
					val["TeacherID"],
				}
				tCourses = append(tCourses, tCourse)
			} else { //key不存在于redis中
				idsNotExistRedis = append(idsNotExistRedis, ids[index])
			}
		}
	}
	if len(idsNotExistRedis) != 0 {
		//处理未在redis找到的课程id
		var courses []models.Course
		//先在数据库中查询
		if err := db.Find(&courses, idsNotExistRedis).Error; err != nil {
			// 判断课程是否存在
			if errors.Is(err, gorm.ErrRecordNotFound) {
				code = types.CourseNotExisted
			} else {
				code = types.UnknownError
			}
			return
		}

		//存在，写入redis中,设置十分钟过期
		for _, course := range courses {
			key := fmt.Sprintf(types.TCourseKey, course.ID)
			client.HSet(ctx, key, map[string]interface{}{
				"Name":      course.Name,
				"TeacherID": course.TeacherID,
			})
			client.Expire(ctx, key, 600*time.Second)
			tCourse := types.TCourse{
				strconv.FormatInt(course.ID, 10),
				course.Name,
				strconv.FormatInt(course.TeacherID, 10),
			}
			tCourses = append(tCourses, tCourse)
		}
	}
	code = types.OK
	return
}
