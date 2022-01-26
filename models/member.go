package models

import (
	"camp/types"
)

type Member struct {
	ID       int64
	Nickname string
	Username string
	Password string
	UserType types.UserType
	Deleted  types.DeleteType
}

func (Member) TableName() string {
	return "member"
}
