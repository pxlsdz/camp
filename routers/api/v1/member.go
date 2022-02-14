package v1

import (
	"camp/infrastructure/stores/myRedis"
	"camp/infrastructure/stores/mysql"
	"errors"
	"gorm.io/gorm"

	"camp/models"
	"camp/repository"
	"camp/types"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strconv"
)

func CreateMember(c *gin.Context) {
	// 参数校验
	var json types.CreateMemberRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.ParamInvalid})
		return
	}

	if PasswordCheck(json.Password) != nil {
		c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.ParamInvalid})
		return
	}

	// 检验用户名是否存在
	db := mysql.GetDb()
	var member models.Member
	err := db.Where("username = ?", json.Username).Take(&member).Error
	if err == nil {
		c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.UserHasExisted})
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) == false {
		c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.UnknownError})
		return
	}

	member = models.Member{
		Nickname: json.Nickname,
		Username: json.Username,
		Password: json.Password,
		UserType: json.UserType,
		Deleted:  types.Default,
	}

	if err := db.Create(&member).Error; err != nil {
		e1 := fmt.Sprintf("%s", err)
		e2 := fmt.Sprintf("Error 1062: Duplicate entry '%s' for key 'username'", member.Username)
		if e1 == e2 {
			c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.UserHasExisted})
			return
		}
		c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.UnknownError})
		return
	}

	// 加入redis
	if member.UserType == types.Student {
		cli := myRedis.GetClient()
		ctx := context.Background()
		cli.SAdd(ctx, fmt.Sprintf(types.StudentKey), member.ID)

	}

	c.JSON(http.StatusOK, types.CreateMemberResponse{
		Code: types.OK,
		Data: struct {
			UserID string
		}{strconv.FormatInt(member.ID, 10)},
	})
}

func GetMember(c *gin.Context) {

	UserID := c.Query("UserID")
	id, err := strconv.ParseInt(UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: types.ParamInvalid})
		return
	}

	var member models.Member
	if code := repository.GetMemberById(id, &member); code != types.OK {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: code})
		return
	}

	c.JSON(http.StatusOK, types.GetMemberResponse{
		Code: types.OK,
		Data: types.TMember{
			UserID:   strconv.FormatInt(member.ID, 10),
			Nickname: member.Nickname,
			Username: member.Username,
			UserType: member.UserType,
		},
	})
}

func UpdateMember(c *gin.Context) {

	// 参数校验
	var json types.UpdateMemberRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.UpdateMemberResponse{Code: types.ParamInvalid})
		return
	}

	id, err := strconv.ParseInt(json.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, types.UpdateMemberResponse{Code: types.ParamInvalid})
		return
	}

	var member models.Member
	if code := repository.GetMemberById(id, &member); code != types.OK {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: code})
		return
	}

	db := mysql.GetDb()
	if err := db.Model(&member).Update("nickname", json.Nickname).Error; err != nil {
		c.JSON(http.StatusOK, types.UpdateMemberResponse{Code: types.UnknownError})
		return
	}

	c.JSON(http.StatusOK, types.UpdateMemberResponse{Code: types.OK})
}

func DeleteMember(c *gin.Context) {

	// 参数校验
	var json types.DeleteMemberRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.ParamInvalid})
		return
	}

	id, err := strconv.ParseInt(json.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.ParamInvalid})
		return
	}

	var member models.Member
	if code := repository.GetMemberById(id, &member); code != types.OK {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: code})
		return
	}

	db := mysql.GetDb()
	if err := db.Model(&member).Update("deleted", types.Deleted).Error; err != nil {
		c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.UnknownError})
		return
	}

	//if err := db.Where("student_id", member.ID).Delete(&models.StudentCourse{}).Error; err != nil {
	//	c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.UnknownError})
	//	return
	//}
	//

	cli := myRedis.GetClient()
	ctx := context.Background()
	pipe := cli.Pipeline()

	pipe.SRem(ctx, fmt.Sprintf(types.StudentKey), member.ID)
	pipe.Del(ctx, fmt.Sprintf(types.StudentHasCourseKey, member.ID))

	_, err = pipe.Exec(ctx)
	if err != nil {
		c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.UnknownError})
		return
	}
	c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.OK})
}

func GetMemberList(c *gin.Context) {
	offsetString := c.Query("Offset")
	limitString := c.Query("Limit")
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: types.ParamInvalid})
		return
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: types.ParamInvalid})
		return
	}

	var members []models.Member
	db := mysql.GetDb()
	if err := db.Where("deleted = ?", types.Default).Offset(offset).Limit(limit).Find(&members).Error; err != nil {
		c.JSON(http.StatusOK, types.GetMemberListResponse{Code: types.UnknownError})
		return
	}
	var memberList []types.TMember
	for _, member := range members {
		memberList = append(memberList, types.TMember{
			UserID:   strconv.FormatInt(member.ID, 10),
			Nickname: member.Nickname,
			Username: member.Username,
			UserType: member.UserType,
		})
	}

	c.JSON(http.StatusOK, types.GetMemberListResponse{
		Code: types.OK,
		Data: struct {
			MemberList []types.TMember
		}{memberList},
	})
}

func PasswordCheck(passwd string) error {
	num := `[0-9]{1}`
	a_z := `[a-z]{1}`
	A_Z := `[A-Z]{1}`
	if b, err := regexp.MatchString(num, passwd); !b || err != nil {
		return fmt.Errorf("password need num :%v", err)
	}
	if b, err := regexp.MatchString(a_z, passwd); !b || err != nil {
		return fmt.Errorf("password need a_z :%v", err)
	}
	if b, err := regexp.MatchString(A_Z, passwd); !b || err != nil {
		return fmt.Errorf("password need A_Z :%v", err)
	}
	return nil
}
