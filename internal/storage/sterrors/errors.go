package sterrors

import "errors"

var (
	ErrLessonAlreadyExists = errors.New("lesson with this teacher id, classname and lesson name already exists")
	ErrTeacherNotFound     = errors.New("teacher with this id not found")
)
