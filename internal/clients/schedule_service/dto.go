package schedule

type SetHomeworkReq struct {
	Date       string `json:"date"`
	Classname  string `json:"classname"`
	Homework   string `json:"homework"`
	LessonName string `json:"lesson_name"`
}

type SetWeeklyScheduleReq struct {
	Filename string `json:"filename"  bson:"filename"`
	Sheet    string `json:"sheet"  bson:"sheet"`
	FileData []byte `json:"file_data"  bson:"file_data"`
}
