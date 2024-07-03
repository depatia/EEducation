package lesson_service

type SetLessonReq struct {
	TeacherID  int64  `json:"teacher_id"`
	LessonName string `json:"lesson_name"`
	ClassName  string `json:"class_name"`
}
