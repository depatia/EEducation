package models

import "database/sql"

type ScheduleDay struct {
	ClassName  string
	LessonName string
	Classroom  int64
	Date       string
}

type Lesson struct {
	Date       string
	LessonName string
	Classroom  int64
	Grade      sql.NullInt64
	Homework   sql.NullString
}

type StudentsScheduleDay struct {
	ClassName string
	Lessons   []*Lesson
	Date      string
}
