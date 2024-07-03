package postgre

import (
	"GradesMS/internal/pb"
	"GradesMS/internal/storage/sterrors"
	"GradesMS/tools/converter"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (s *StDb) GetAllGrades(ctx context.Context, studentID int64) ([]*pb.Grade, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT student_id, lesson_name, date, grade FROM service.grades WHERE student_id = $1", studentID)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query due to error: %w", err)
	}
	defer rows.Close()

	var grades []*pb.Grade
	for rows.Next() {
		grade := new(pb.Grade)
		err = rows.Scan(&grade.StudentId, &grade.LessonName, &grade.Date, &grade.Grade)

		if err != nil {
			return nil, fmt.Errorf("failed to scanning rows due to error: %w", err)
		}

		grade.Date = converter.Convert(grade.Date)

		grades = append(grades, grade)
	}

	return grades, nil
}

func (s *StDb) GetLessonGrades(ctx context.Context, studentID int64, lessonName string) ([]*pb.Grade, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT student_id, lesson_name, date, grade FROM service.grades WHERE student_id = $1 AND lesson_name = $2", studentID, lessonName)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement due to error: %w", err)
	}

	defer rows.Close()

	var grades []*pb.Grade
	for rows.Next() {
		grade := new(pb.Grade)
		err = rows.Scan(&grade.StudentId, &grade.LessonName, &grade.Date, &grade.Grade)
		if err != nil {
			return nil, fmt.Errorf("failed to scanning rows due to error: %w", err)
		}

		grade.Date = converter.Convert(grade.Date)

		grades = append(grades, grade)
	}

	return grades, nil
}

func (s *StDb) SetGrade(ctx context.Context, studentID int64, lessonName string, grade int64, isTerm bool) (int64, error) {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO service.grades(student_id, lesson_name, date, grade, is_term) VALUES($1, $2, $3, $4, $5)",
		studentID,
		lessonName,
		time.Now().Format(time.DateOnly),
		grade,
		isTerm,
	)

	fmt.Println(studentID, lessonName, grade, isTerm)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("failed to set the grade due to error: %w", sterrors.ErrGradeAlreadyExists)
			}
		}

		return 0, fmt.Errorf("failed to set the grade due to error: %w", err)
	}

	return studentID, nil
}

func (s *StDb) UpdateGrade(ctx context.Context, studentID int64, lessonName string, grade int64, date string) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE service.grades SET grade = $1 WHERE student_id = $2 AND lesson_name = $3 AND date = $4",
		grade,
		studentID,
		lessonName,
		date,
	)

	if err != nil {
		return fmt.Errorf("failed to update the grade due to error: %w", err)
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return sterrors.ErrGradeNotFound
	}

	return nil
}

func (s *StDb) DelGrade(ctx context.Context, studentID int64, lessonName string, date string) error {
	res, err := s.db.ExecContext(ctx,
		"DELETE FROM service.grades WHERE student_id = $1 AND lesson_name = $2 AND date = $3",
		studentID,
		lessonName,
		date,
	)

	if err != nil {
		return fmt.Errorf("failed to set the grade due to error: %w", err)
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return sterrors.ErrGradeNotFound
	}

	return nil
}
