package v1

import (
	"camp/infrastructure/goCache"
	"camp/infrastructure/mq/rabbitmq"
	"camp/infrastructure/stores/myRedis"
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/repository"
	"camp/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"net/http"
	"strconv"
)

func GetStudentCourse(c *gin.Context) {

	//参数校验
	StudentID := c.Query("StudentID")
	studentID, err := strconv.ParseInt(StudentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.ParamInvalid})
		return
	}

	ctx := context.Background()
	cli := myRedis.GetClient()

	// 学生布隆过滤器
	val, err := cli.Do(ctx, "BF.EXISTS", types.BStudentKey, studentID).Bool()
	if err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	} else if val == false {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentNotExisted})
		return
	}

	db := mysql.GetDb()

	//判断学生是否存在
	//逻辑和抢课函数一致
	val, err = cli.SIsMember(ctx, types.StudentKey, studentID).Result()
	if err != nil {
		c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.UnknownError})
		return
	} else if val == false {
		if code := repository.GetBoolStudentById(studentID); code != types.OK {
			c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: code})
			return
		}
		cli.SAdd(ctx, types.StudentKey, studentID)
	}

	//判断课程列表是否存在
	key := fmt.Sprintf(types.StudentHasCourseKey, studentID)
	var courseIDs []int64
	all, err := cli.SMembers(ctx, key).Result()

	if err != nil {
		c.JSON(http.StatusOK, types.GetCourseResponse{Code: types.UnknownError})
		return
	} else if all != nil && len(all) > 0 { //key存在于redis中
		for _, id := range all {
			courseID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.UnknownError})
				return
			}
			courseIDs = append(courseIDs, courseID)
		}
	} else {
		if err := db.Select("course_id").Where("student_id = ?", studentID).Model(&models.StudentCourse{}).Scan(&courseIDs).Error; err != nil {
			c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.UnknownError})
			return
		}

		if courseIDs == nil || len(courseIDs) == 0 {
			c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.StudentHasNoCourse})
			return
		}

		//课程列表写入redis
		t := make([]interface{}, len(courseIDs))
		for i, courseID := range courseIDs {
			t[i] = courseID
		}
		cli.SAdd(ctx, key, t)
	}

	//courseList := make([]types.TCourse, len(courseIDs))
	//for i, id := range courseIDs {
	//	repository.GetTCourseByID(id, &courseList[i])
	//}

	courseList, code := repository.GetTCourseByIDs(courseIDs)
	if code != types.OK {
		courseList = nil
	}

	c.JSON(http.StatusOK, types.GetStudentCourseResponse{
		Code: code,
		Data: struct{ CourseList []types.TCourse }{CourseList: courseList},
	})
	//db := mysql.GetDb()
	//var member models.Member
	//result := db.Take(&member, studentID)
	//
	//// 判断用户是否存在
	//if result.RowsAffected == 0 {
	//	c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.UserNotExisted})
	//	return
	//}
	//
	//// 判断用户是否已经删除
	//if member.Deleted == types.Deleted {
	//	c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.UserHasDeleted})
	//	return
	//}
	//
	//var courseList []types.TCourse
	//if err := db.Raw("select c.id as course_id, c.name, c.teacher_id from student_course sc join course c on  sc.course_id = c.id where sc.student_id = ?", studentID).Scan(&courseList).Error; err != nil {
	//	c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.UnknownError})
	//	return
	//
	//}

	//if err := db.Raw("SELECT id as course_id, name, teacher_id FROM course WHERE id IN (SELECT course_id FROM student_course WHERE student_id = ?)", studentID).Scan(&courseList).Error; err != nil {
	//	if errors.Is(err, gorm.ErrRecordNotFound) {
	//		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.StudentHasNoCourse})
	//		return
	//	} else {
	//		c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.UnknownError})
	//		return
	//	}
	//}

	//if courseList == nil || len(courseList) == 0 {
	//	c.JSON(http.StatusOK, types.GetStudentCourseResponse{Code: types.StudentHasNoCourse})
	//	return
	//}
	//c.JSON(http.StatusOK, types.GetStudentCourseResponse{
	//	Code: types.OK,
	//	Data: struct{ CourseList []types.TCourse }{CourseList: courseList},
	//})

}

//var localCapOverMap map[int64]bool

func BookCourse(c *gin.Context) {

	// 参数校验
	var requestJson types.BookCourseRequest
	if err := c.ShouldBindJSON(&requestJson); err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.ParamInvalid})
		return
	}

	courseID, err := strconv.ParseInt(requestJson.CourseID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.ParamInvalid})
		return
	}

	studentID, err := strconv.ParseInt(requestJson.StudentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.ParamInvalid})
		return
	}

	cache := goCache.GetCache()
	_, err = cache.Get([]byte(requestJson.CourseID))
	if err == nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotAvailable})
		return
	}

	ctx := context.Background()
	cli := myRedis.GetClient()

	// 学生布隆过滤器
	val, err := cli.Do(ctx, "BF.EXISTS", types.BStudentKey, studentID).Bool()
	if err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	} else if val == false {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentNotExisted})
		return
	}

	// 课程布隆过滤器
	val, err = cli.Do(ctx, "BF.EXISTS", types.BCourseKey, courseID).Bool()
	if err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	} else if val == false {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotExisted})
		return
	}

	db := mysql.GetDb()

	// 被删除的用户一直攻击 需要做特殊出来，缓存
	// 判断学生是否存在
	val, err = cli.SIsMember(ctx, types.StudentKey, studentID).Result()
	if err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	} else if val == false {
		if code := repository.GetBoolStudentById(studentID); code != types.OK {
			c.JSON(http.StatusOK, types.BookCourseResponse{Code: code})
			return
		}
		// 更新缓存
		cli.SAdd(ctx, types.StudentKey, studentID)
	}

	// 判断课程是否存在
	capRedis, err := cli.Get(ctx, fmt.Sprintf(types.CourseKey, courseID)).Result()
	if err == myRedis.Nil {
		var capCnt int
		if code := repository.GetCapCourseById(courseID, &capCnt); code != types.OK {
			c.JSON(http.StatusOK, types.BookCourseResponse{Code: code})
			return
		}
		cli.Set(ctx, fmt.Sprintf(types.CourseKey, courseID), capCnt, -1)
	} else if err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	} else {
		capCnt, _ := strconv.ParseInt(capRedis, 10, 64)
		if capCnt <= 0 {
			cache.Set([]byte(requestJson.CourseID), []byte("1"), 3600)
			c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotAvailable})
			return
		}
	}

	// 申请锁
	pool := goredis.NewPool(cli)
	rs := redsync.New(pool)

	key := fmt.Sprintf(types.StudentIDCourseIDKey, studentID, courseID)
	lockKey := fmt.Sprintf(types.StudentLock, studentID)
	mutex := rs.NewMutex(lockKey)

	if err := mutex.LockContext(ctx); err != nil {
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	}

	// 选课记录布隆过滤器
	val, err = cli.Do(ctx, "BF.EXISTS", types.BStudentHasCourseKey, key).Bool()
	if err != nil {
		if _, err := mutex.UnlockContext(ctx); err != nil {
		}
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	} else if val == true {
		// 标志学生是否含有该课程
		flag := false
		// 判断redis 是否存在
		_, err := cli.Get(ctx, fmt.Sprintf(types.StudentHCourseKey, studentID, courseID)).Result()

		if err == myRedis.Nil {
			// 导入mysql学生选课记录
			var count int64
			if err := db.Model(&models.StudentCourse{}).Where("student_id = ? AND course_id = ?", studentID, courseID).Limit(1).Count(&count).Error; err != nil {
				if _, err := mutex.UnlockContext(ctx); err != nil {
				}
				c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
				return
			}
			if count != 0 {
				flag = true
			}
		} else if err != nil {
			if _, err := mutex.UnlockContext(ctx); err != nil {
			}
			c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
			return
		} else {
			flag = true
		}
		if flag {
			if _, err := mutex.UnlockContext(ctx); err != nil {
			}
			c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentHasCourse})
			return
		}
	}

	//  预扣减库存
	stock, err := cli.Decr(ctx, fmt.Sprintf(types.CourseKey, courseID)).Result()
	if err != nil {
		if _, err := mutex.UnlockContext(ctx); err != nil {
		}
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	}

	if stock < 0 {
		if _, err := mutex.UnlockContext(ctx); err != nil {

		}
		cache.Set([]byte(requestJson.CourseID), []byte("1"), 3600)
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotAvailable})
		return
	}

	_, err = cli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		// 删除课程缓存
		pipe.Del(ctx, fmt.Sprintf(types.StudentHasCourseKey, studentID))
		pipe.SetNX(ctx, fmt.Sprintf(types.StudentHCourseKey, studentID, courseID), 1, types.RedisWriteExpiration)
		pipe.Do(ctx, "BF.ADD", types.BStudentHasCourseKey, fmt.Sprintf(types.StudentIDCourseIDKey, studentID, courseID))
		return nil
	})

	if err != nil {
		cli.Incr(ctx, fmt.Sprintf(types.CourseKey, courseID))
		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
		return
	}

	if _, err := mutex.UnlockContext(ctx); err != nil {
	}

	// 消息队列减少课程数据库的库存以及创建数据库表
	//创建消息体

	studentCourse := models.StudentCourse{
		StudentID: studentID,
		CourseID:  courseID,
	}
	//类型转化
	byteMessage, _ := json.Marshal(studentCourse)

	rabbitMQ := rabbitmq.GetRabbitMQ()
	err = rabbitMQ.PublishSimple(string(byteMessage))
	//if err != nil {
	//}

	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.OK})

	// redis lua脚本实现检验该学生是否已经有该课和课程数量是否足够
	// 缓存设计待讨论
	// 学生是否存在与商品是否存在 是否和减少库存是一个原子性质操作？
	//res, err := cli.EvalSha(ctx, myRedis.LuaHash, []string{fmt.Sprintf(types.StudentHasCourseKey, studentID, courseId), fmt.Sprintf(types.CourseKey, courseId), fmt.Sprintf(types.StudentKey, studentID)}).Result()
	//
	//if err != nil || res == int64(-1) {
	//	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
	//	return
	//}
	//if res == int64(4) {
	//	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentNotExisted})
	//	return
	//}
	//
	//if res == int64(3) {
	//	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentHasCourse})
	//	return
	//}
	//if res == int64(2) {
	//	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotExisted})
	//	return
	//}
	//if res == int64(0) {
	//	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotAvailable})
	//	localCapOverMap[courseId] = true
	//	return
	//}
	//// 消息队列减少课程数据库的库存以及创建数据库表
	////创建消息体
	//
	//studentCourse := models.StudentCourse{
	//	StudentID: studentID,
	//	CourseID:  courseId,
	//}
	////类型转化
	//byteMessage, _ := json.Marshal(studentCourse)
	////if err != nil {
	////
	////}
	//rabbitMQ := rabbitmq.GetRabbitMQ()
	//err = rabbitMQ.PublishSimple(string(byteMessage))
	////if err != nil {
	////}
	//
	//c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.OK})
	//
	//return

	//// 加锁
	//if res, err := cli.SetNX(ctx, fmt.Sprintf("%sl%s", requestJson.StudentID, requestJson.CourseID), time.Now().Unix(), time.Minute).Result(); err != nil || !res {
	//	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
	//	return
	//}

	//// 判断学生是否已经拥有该课程
	//if _, err := cli.Get(ctx, fmt.Sprintf(types.StudentHasCourseKey, requestJson.StudentID, requestJson.CourseID)).Result(); err != redis.Nil {
	//	if err == nil {
	//		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.StudentHasCourse})
	//	} else {
	//		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
	//	}
	//	return
	//}
	//
	// 预减库存
	//courseId, _ := strconv.ParseInt(requestJson.CourseID, 10, 64)
	//stock, err := cli.Decr(ctx, fmt.Sprintf(types.CourseKey, courseId)).Result()
	//if err != nil {
	//	if err == redis.Nil {
	//		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotExisted})
	//	} else {
	//		c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.UnknownError})
	//	}
	//	return
	//}
	//if stock < 0 {
	//	c.JSON(http.StatusOK, types.BookCourseResponse{Code: types.CourseNotAvailable})
	//}

}
