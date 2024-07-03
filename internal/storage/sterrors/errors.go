package sterrors

import "errors"

var (
	ErrClassNotFound    = errors.New("class with this ID not found")
	ErrScheduleNotFound = errors.New("lesson with class or date not found")
)
