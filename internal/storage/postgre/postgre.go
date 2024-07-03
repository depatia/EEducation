package postgre

import (
	"ScheduleMS/internal/models"
	"ScheduleMS/internal/storage/sterrors"
	"ScheduleMS/internal/utils"
	"ScheduleMS/tools/converter"
	dateFormatter "ScheduleMS/tools/date_formatter"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type StDb struct {
	db *sql.DB
}

func New(path string) (*StDb, error) {
	db, err := sql.Open("pgx", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db due to error: %w", err)
	}
	return &StDb{db: db}, nil
}

func (s *StDb) GetWeekSchedule(ctx context.Context, className string) ([]*models.StudentsScheduleDay, error) {
	rows, err := s.db.QueryContext(
		ctx,
		"SELECT lesson_name, date, grade, classroom, homework  FROM service.schedule WHERE classname = $1 AND date >= $2 AND date <= $3",
		className,
		dateFormatter.FirstDayOfWeek(),
		dateFormatter.LastDayOfWeek(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query due to error: %w", err)
	}
	defer rows.Close()

	var lessons []*models.Lesson
	for rows.Next() {
		lesson := new(models.Lesson)
		err = rows.Scan(&lesson.LessonName, &lesson.Date, &lesson.Grade, &lesson.Classroom, &lesson.Homework)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("failed to get week schedule: %w", sterrors.ErrClassNotFound)
			}

			return nil, fmt.Errorf("failed to get week schedule due to error: %w", err)
		}

		lesson.Date = converter.Convert(lesson.Date)
		lessons = append(lessons, lesson)
	}
	week := utils.SortLessons(lessons)

	return week, nil
}

func (s *StDb) SetDailySchedule(ctx context.Context, day models.StudentsScheduleDay) error {
	for _, lesson := range day.Lessons {
		_, err := s.db.ExecContext(ctx,
			"INSERT INTO service.schedule(lesson_name, classroom, date, classname) VALUES ($1, $2, $3, $4)",
			lesson.LessonName,
			lesson.Classroom,
			day.Date,
			day.ClassName,
		)
		if err != nil {
			return fmt.Errorf("failed to set the daily schedule due to error: %w", err)
		}
	}

	return nil
}

func (s *StDb) SetHomework(ctx context.Context, className string, date string, lessonName string, homework string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE service.schedule SET homework = $1 WHERE classname = $2 AND lesson_name = $3 AND date = $4",
		homework,
		className,
		lessonName,
		date,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to set homework due to error: %w", sterrors.ErrScheduleNotFound)
		}

		return fmt.Errorf("failed to set homework due to error: %w", err)
	}

	return nil
}
