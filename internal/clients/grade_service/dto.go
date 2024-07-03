package grade

type Grade struct {
	StudentId  int64
	LessonName string
	Date       string
}

type SetGradeReq struct {
	StudentID  int64  `json:"student_id"`
	DeviceID   int    `json:"device_id"`
	LessonName string `json:"lesson_name"`
	Grade      int64  `json:"grade"`
}

type GetLessonGradesResp struct {
	StudentID  int64
	LessonName string
}

type ChangeGradeReq struct {
	StudentID  int64  `json:"student_id"  bson:"student_id"`
	LessonName string `json:"lesson_name"  bson:"lesson_name"`
	Grade      int64  `json:"grade"  bson:"grade"`
	Date       string `json:"date"  bson:"date"`
}

type DeleteGradeReq struct {
	StudentID  int64  `json:"student_id"  bson:"student_id"`
	LessonName string `json:"lesson_name"  bson:"lesson_name"`
	Date       string `json:"date"  bson:"date"`
}

type SetTermGradeReq struct {
	UserID     int64  `json:"user_id"  bson:"user_id"`
	LessonName string `json:"lesson_name"  bson:"lesson_name"`
}
