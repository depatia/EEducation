package utils

import (
	"ScheduleMS/internal/models"
)

func SortLessons(lessons []*models.Lesson) []*models.StudentsScheduleDay {
	week := []*models.StudentsScheduleDay{}

	var wow = make(map[string][]*models.Lesson)

	for _, lesson := range lessons {
		wow[lesson.Date] = append(wow[lesson.Date], lesson)
	}
	for date, lessons := range wow {
		day := new(models.StudentsScheduleDay)
		day.Date = date
		day.Lessons = lessons
		week = append(week, day)
	}

	return week
}
