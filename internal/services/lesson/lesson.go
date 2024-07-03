package lesson

import (
	"LessonsMS/internal/models"
	"LessonsMS/internal/storage/sterrors"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var ErrIncorrectTeacherID = errors.New("incorrect teacher id")

type LessonStore struct {
	log            *slog.Logger
	lessonCreater  LessonCreater
	lessonProvider LessonProvider
}

func New(log *slog.Logger, lessonCreater LessonCreater, lessonProvider LessonProvider) *LessonStore {
	return &LessonStore{log: log, lessonCreater: lessonCreater, lessonProvider: lessonProvider}
}

type LessonCreater interface {
	CreateLesson(
		ctx context.Context,
		teacherID int64,
		className string,
		lessonName string,
	) error
}

type LessonProvider interface {
	IsCombinationExists(
		ctx context.Context,
		lessonName string,
		className string,
	) (bool, error)
	GetLessonsByTeacherID(
		ctx context.Context,
		teacherID int64,
	) ([]*models.Lesson, error)
	IsExistsByClassname(
		ctx context.Context,
		className string,
	) (bool, error)
}

func (s *LessonStore) CreateNewLesson(ctx context.Context, teacherID int64, className string, lessonName string) error {
	log := s.log.With(
		slog.String("Operation:", "Lessons - CreateNewLesson"),
		slog.Int64("teacherID: ", teacherID),
		slog.String("lesson name: ", lessonName),
	)

	log.Info("creating new lesson")

	err := s.lessonCreater.CreateLesson(ctx, teacherID, className, lessonName)
	if err != nil {
		log.Error("failed to create lesson", err)

		return fmt.Errorf("failed to create lesson due to error: %w", err)
	}

	log.Info("lesson is created")

	return nil
}

func (s *LessonStore) GetLessons(ctx context.Context, teacherID int64) ([]*models.Lesson, error) {
	log := s.log.With(
		slog.String("Operation: ", "Lessons - GetLessons"),
		slog.Int64("teacher ID: ", teacherID),
	)
	log.Info("getting lessons by teacher id")

	lessons, err := s.lessonProvider.GetLessonsByTeacherID(ctx, teacherID)
	if err != nil {
		log.Error("failed to get teacher lessons:", err)
		if errors.Is(err, sterrors.ErrTeacherNotFound) {
			return nil, fmt.Errorf("failed to get teacher lessons: %w", ErrIncorrectTeacherID)
		}
		return nil, fmt.Errorf("failed to get teacher lessons due to error: %w", err)
	}

	log.Info("lessons given")

	return lessons, nil
}

func (s *LessonStore) IsLessonWithClassExists(ctx context.Context, lessonName string, className string) (bool, error) {
	log := s.log.With(
		slog.String("Operation:", "Lessons - IsLessonWithClassExists"),
	)

	exists, err := s.lessonProvider.IsCombinationExists(ctx, lessonName, className)
	if err != nil {
		log.Error("failed to check if the lesson exists due to error: ", err)

		return exists, fmt.Errorf("failed to check if the lesson exists due to error: %w", err)
	}

	return exists, nil
}

func (s *LessonStore) IsExistsByClassname(ctx context.Context, className string) (bool, error) {
	log := s.log.With(
		slog.String("Operation:", "Lessons - IsExistsByClassname"),
		slog.String("ClassName:", className),
	)

	exists, err := s.lessonProvider.IsExistsByClassname(ctx, className)
	if err != nil {
		log.Error("failed to check if the class exists due to error: ", err)

		return exists, fmt.Errorf("failed to check if the class exists due to error: %w", err)
	}

	return exists, nil
}
