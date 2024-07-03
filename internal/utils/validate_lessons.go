package utils

import (
	"ScheduleMS/internal/models"
	"ScheduleMS/internal/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidatePbLessons(lessons []*pb.Lesson) error {
	for _, lesson := range lessons {
		if lesson.Classroom == 0 {
			return status.Error(codes.InvalidArgument, "classname is required")
		}
		if lesson.LessonName == "" {
			return status.Error(codes.InvalidArgument, "lesson name is required")
		}
	}

	return nil
}

func ValidateLessons(lessons []*models.Lesson) error {
	for _, lesson := range lessons {
		if lesson.Classroom == 0 {
			return status.Error(codes.InvalidArgument, "classname is required")
		}
		if lesson.LessonName == "" {
			return status.Error(codes.InvalidArgument, "lesson name is required")
		}
	}

	return nil
}
