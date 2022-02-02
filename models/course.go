package models

import "camp/types"

type Course struct {
	ID        int64
	Name      string
	Cap       int
	TeacherID int64
	Deleted   types.DeleteType
}

func (Course) TableName() string {
	return "course"
}
