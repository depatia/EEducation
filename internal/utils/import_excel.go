package utils

import (
	"ScheduleMS/internal/models"
	"fmt"
	"io"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func GetWeekFromExcelFile(r io.Reader, sheet string) ([]models.StudentsScheduleDay, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fmt.Println(sheet)
	cols, err := f.GetCols("Лист1")
	if err != nil {
		return nil, err
	}
	week := make([]models.StudentsScheduleDay, 0)

	var helpClassroom int64

	for i, col := range cols {
		day := new(models.StudentsScheduleDay)
		lesson := new(models.Lesson)
		day.ClassName = cols[0][0]
		if i%2 == 0 {
			for y, c := range col {
				if y >= 2 {
					classroom, _ := strconv.Atoi(c)
					helpClassroom = int64(classroom)
				}
			}
			continue
		} else {
			for y, c := range col {
				if y >= 2 {
					day.Date = col[1]
					lesson.Classroom = helpClassroom
					lesson.LessonName = c
					lesson.Date = day.Date
					day.Lessons = append(day.Lessons, lesson)
				}
			}
		}
		week = append(week, *day)
	}

	return week, nil
}
