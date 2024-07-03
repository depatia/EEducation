package grade

import (
	"GradesMS/internal/pb"
	"GradesMS/internal/storage/sterrors"
	"GradesMS/tools/logger/sl"
	"context"
	"fmt"
	"log/slog"
)

type Grade struct {
	log           *slog.Logger
	gradeSetter   GradeSetter
	gradeGetter   GradeGetter
	gradeProvider GradeProvider
}

func New(log *slog.Logger, gradeSetter GradeSetter, gradeProvider GradeProvider, gradeGetter GradeGetter) *Grade {
	return &Grade{log: log, gradeSetter: gradeSetter, gradeProvider: gradeProvider, gradeGetter: gradeGetter}
}

type GradeSetter interface {
	SetGrade(
		ctx context.Context,
		studentID int64,
		lessonName string,
		grade int64,
		isTerm bool,
	) (int64, error)
}

type GradeGetter interface {
	GetLessonGrades(
		ctx context.Context,
		studentID int64,
		lessonName string,
	) ([]*pb.Grade, error)
	GetAllGrades(
		ctx context.Context,
		studentID int64,
	) ([]*pb.Grade, error)
}

type GradeProvider interface {
	UpdateGrade(
		ctx context.Context,
		studentID int64,
		lessonName string,
		grade int64,
		date string,
	) error
	DelGrade(
		ctx context.Context,
		studentID int64,
		lessonName string,
		date string,
	) error
}

func (g *Grade) AddGrade(ctx context.Context, studentID int64, lessonName string, grade int64, isTerm bool) (int64, error) {
	log := g.log.With(
		slog.String("op", "grade.AddGrade"),
		slog.String("lesson Name", lessonName),
	)

	log.Info("creating new grade")

	id, err := g.gradeSetter.SetGrade(ctx, studentID, lessonName, grade, isTerm)
	if err != nil {
		log.Error("failed to set user grade", sl.Err(err))

		return 0, fmt.Errorf("failed to set grade due to error: %w", err)
	}

	log.Info("grade successfully created")

	return id, nil
}

func (g *Grade) GetLessonGradesByStudentID(ctx context.Context, studentID int64, lessonName string) ([]*pb.Grade, error) {
	log := g.log.With(
		slog.String("op", "grade.GetLessonGradesByStudentID"),
		slog.String("lesson Name", lessonName),
	)

	log.Info("getting lesson's grades by student id")

	grades, err := g.gradeGetter.GetLessonGrades(ctx, studentID, lessonName)
	if err != nil {
		log.Error("failed to get user's lesson grades", sl.Err(err))

		return nil, fmt.Errorf("failed to get the list of lessons with grades due to error: %w", err)
	}

	log.Info("grades successfully given")

	return grades, nil
}

func (g *Grade) GetAllGradesByStudentID(ctx context.Context, studentID int64) ([]*pb.Grade, error) {
	log := g.log.With(
		slog.String("op", "grade.GetAllGradesByStudentID"),
		slog.Int64("student id", studentID),
	)

	log.Info("getting all grades by student id")

	grades, err := g.gradeGetter.GetAllGrades(ctx, studentID)
	if err != nil {
		log.Error("failed to get user grades", sl.Err(err))

		return nil, fmt.Errorf("failed to get the list of all grades due to error: %w", err)
	}

	log.Info("grades successfully given")

	return grades, nil
}

func (g *Grade) ChangeGrade(ctx context.Context, studentID int64, lessonName string, grade int64, date string) error {
	log := g.log.With(
		slog.String("op", "grade.ChangeGrade"),
		slog.Int64("student id", studentID),
	)

	err := g.gradeProvider.UpdateGrade(ctx, studentID, lessonName, grade, date)
	if err != nil {
		log.Error("failed to update user grades", sl.Err(err))

		return err
	}

	return nil
}

func (g *Grade) DeleteGrade(ctx context.Context, studentID int64, lessonName string, date string) error {
	log := g.log.With(
		slog.String("op", "grade.DeleteGrade"),
		slog.Int64("student id", studentID),
	)
	log.Info("deleting grade")

	err := g.gradeProvider.DelGrade(ctx, studentID, lessonName, date)
	if err != nil {
		log.Error("failed to delete user's grade", sl.Err(err))

		return err
	}

	log.Info("grade successfully deleted")

	return nil
}
func (g *Grade) SetTermGrade(ctx context.Context, studentID int64, lessonName string) error {
	log := g.log.With(
		slog.String("op", "grade.SetTermGrade"),
		slog.Int64("student_id", studentID),
	)

	log.Info("setting term grade")

	grades, err := g.gradeGetter.GetLessonGrades(ctx, studentID, lessonName)
	if err != nil {
		log.Error("failed to get lesson grades", sl.Err(err))

		return err
	}

	sumGrades := 0

	for _, grade := range grades {
		sumGrades += int(grade.Grade)
	}
	if sumGrades == 0 {
		log.Warn("sum of the grades must not be 0", sl.Err(sterrors.ErrNotEnoughGrades))
		return sterrors.ErrNotEnoughGrades
	}

	termGrade := int64(sumGrades / len(grades))

	_, err = g.gradeSetter.SetGrade(ctx, studentID, lessonName, termGrade, true)
	if err != nil {
		log.Error("failed to set term grade", sl.Err(err))

		return err
	}

	log.Info("term grade successfully created")

	return nil
}
