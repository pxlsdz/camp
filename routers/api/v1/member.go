package v1

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"camp/types"
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
	find := db.Limit(1).Where("username = ?", json.Username).Find(&member)
	if find.RowsAffected == 1 {
		c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.UserHasExisted})
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
		c.JSON(http.StatusOK, types.CreateMemberResponse{Code: types.UnknownError})
		return
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

	db := mysql.GetDb()
	var member models.Member
	result := db.Take(&member, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: types.UserNotExisted})
		return
	}
	// 判断用户是否已经删除
	if member.Deleted == types.Deleted {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: types.UserHasDeleted})
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
	db := mysql.GetDb()
	result := db.Take(&member, id)
	// 判断用户是否存在
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, types.UpdateMemberResponse{Code: types.UserNotExisted})
		return
	}
	// 判断用户是否已经删除
	if member.Deleted == types.Deleted {
		c.JSON(http.StatusOK, types.UpdateMemberResponse{Code: types.UserHasDeleted})
		return
	}

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
	db := mysql.GetDb()
	result := db.Take(&member, id)
	// 判断用户是否存在
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.UserNotExisted})
		return
	}

	// 判断用户是否已经删除
	if member.Deleted == types.Deleted {
		c.JSON(http.StatusOK, types.GetMemberResponse{Code: types.UserHasDeleted})
		return
	}

	if err := db.Model(&member).Update("deleted", types.Deleted).Error; err != nil {
		c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.UnknownError})
		return
	}

	c.JSON(http.StatusOK, types.DeleteMemberResponse{Code: types.OK})
}

func GetMemberList(c *gin.Context) {

	// 参数校验
	var json types.GetMemberListRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusOK, types.GetMemberListResponse{Code: types.ParamInvalid})
		return
	}

	var members []models.Member
	db := mysql.GetDb()
	if err := db.Where("deleted = ?", types.Default).Offset(json.Offset).Limit(json.Limit).Find(&members).Error; err != nil {
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
