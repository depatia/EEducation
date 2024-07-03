package utils

import (
	"LessonsMS/internal/models"
	"LessonsMS/internal/pb"
)

func LessonsMapper(lessons []*models.Lesson) []*pb.TeacherLesson {
	var pbLessons []*pb.TeacherLesson
	for _, lesson := range lessons {
		pbLessons = append(pbLessons, &pb.TeacherLesson{
			LessonName: lesson.LessonName,
			ClassName:  lesson.ClassName,
		})
	}
	return pbLessons
}
