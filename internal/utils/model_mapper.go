package utils

import (
	"ScheduleMS/internal/models"
	"ScheduleMS/internal/pb"
)

func DayMapper(day *pb.ScheduleDay, className string) models.StudentsScheduleDay {
	lessons := []*models.Lesson{}
	for _, lesson := range day.Lessons {
		les := new(models.Lesson)
		les.Classroom = lesson.Classroom
		les.LessonName = lesson.LessonName
		les.Grade.Int64 = lesson.Grade
		les.Homework.String = lesson.Homework
		lessons = append(lessons, les)
	}
	return models.StudentsScheduleDay{
		ClassName: className,
		Lessons:   lessons,
		Date:      day.Date,
	}
}

func WeekMapper(week []*models.StudentsScheduleDay) *pb.ScheduleWeek {
	var pbWeek pb.ScheduleWeek
	for _, day := range week {
		pbDay := new(pb.ScheduleDay)
		pbDay.Date = day.Date
		for _, lesson := range day.Lessons {
			pbDay.Lessons = append(pbDay.Lessons, &pb.Lesson{
				Classroom:  lesson.Classroom,
				LessonName: lesson.LessonName,
				Grade:      lesson.Grade.Int64,
				Homework:   lesson.Homework.String,
			})
		}
		pbWeek.ScheduleWeek = append(pbWeek.ScheduleWeek, pbDay)
	}
	return &pbWeek
}
