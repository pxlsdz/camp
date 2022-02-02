package models

type StudentCourse struct {
	ID        int64
	StudentID int64
	CourseID  int64
}

func (StudentCourse) TableName() string {
	return "student_course"
}
