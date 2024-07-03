package postgre

import (
	"LessonsMS/internal/models"
	"LessonsMS/internal/storage/sterrors"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (s *StDb) IsCombinationExists(ctx context.Context, lessonName string, className string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT exists (SELECT service.lessons.id FROM service.lessons where service.lessons.lesson_name = $1 AND service.lessons.classname = $2)",
		lessonName,
		className,
	).Scan(&exists)

	if err != nil {
		return exists, fmt.Errorf("failed to execute a query due to error: %w", err)
	}

	return exists, nil
}

func (s *StDb) IsExistsByClassname(ctx context.Context, className string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT exists (SELECT service.lessons.id FROM service.lessons where service.lessons.classname = $1)",
		className,
	).Scan(&exists)

	if err != nil {
		return exists, fmt.Errorf("failed to execute a query due to error: %w", err)
	}

	return exists, nil
}

func (s *StDb) CreateLesson(ctx context.Context, teacherID int64, className string, lessonName string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO service.lessons(teacher_id, classname, lesson_name) VALUES ($1, $2, $3)",
		teacherID,
		className,
		lessonName,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("failed to set a new lesson: %w", sterrors.ErrLessonAlreadyExists)
			}
		}
		return fmt.Errorf("failed to set a new lesson: %w", err)
	}

	return nil
}

func (s *StDb) GetLessonsByTeacherID(ctx context.Context, teacherID int64) ([]*models.Lesson, error) {
	rows, err := s.db.QueryContext(
		ctx,
		"SELECT classname, lesson_name FROM service.lessons WHERE teacher_id = $1",
		teacherID,
	)

	defer rows.Close()

	var lessons []*models.Lesson
	for rows.Next() {
		lesson := new(models.Lesson)
		err = rows.Scan(&lesson.ClassName, &lesson.LessonName)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("failed to get lessons: %w", sterrors.ErrTeacherNotFound)
			}

			return nil, fmt.Errorf("failed to scanning rows due to error: %w", err)
		}

		lessons = append(lessons, lesson)
	}

	return lessons, nil
}
