package sterrors

import "errors"

var (
	ErrGradeAlreadyExists = errors.New("student already have the grade on this date")
	ErrStudentNotFound    = errors.New("student not found")
	ErrGradeNotFound      = errors.New("grade not found")
	ErrNotEnoughGrades    = errors.New("not enough grades")
)
