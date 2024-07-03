package schedule

import (
	"ScheduleMS/internal/models"
	"ScheduleMS/internal/utils"
	"ScheduleMS/tools/logger/sl"
	"bytes"
	"context"
	"fmt"
	"log/slog"
)

type ScheduleStore struct {
	log              *slog.Logger
	scheduleSetter   ScheduleSetter
	scheduleProvider ScheduleProvider
}

func New(log *slog.Logger, scheduleSetter ScheduleSetter, scheduleProvider ScheduleProvider) *ScheduleStore {
	return &ScheduleStore{log: log, scheduleSetter: scheduleSetter, scheduleProvider: scheduleProvider}
}

type ScheduleSetter interface {
	SetDailySchedule(
		ctx context.Context,
		day models.StudentsScheduleDay,
	) error
}

type ScheduleProvider interface {
	GetWeekSchedule(
		ctx context.Context,
		className string,
	) ([]*models.StudentsScheduleDay, error)
	SetHomework(
		ctx context.Context,
		className string,
		date string,
		lessonName string,
		homework string,
	) error
}

func (s *ScheduleStore) SetSchedule(ctx context.Context, day models.StudentsScheduleDay) error {
	log := s.log.With(
		slog.String("Operation:", "Schedule - SetSchedule"),
	)

	log.Info("setting daily schedule")

	err := s.scheduleSetter.SetDailySchedule(ctx, day)
	if err != nil {
		log.Error("failed to set schedule", sl.Err(err))

		return fmt.Errorf("failed to set schedule due to error: %w", err)
	}

	log.Info("daily schedule successfully set")

	return nil
}

func (s *ScheduleStore) GetSchedule(ctx context.Context, className string) ([]*models.StudentsScheduleDay, error) {
	log := s.log.With(
		slog.String("Operation:", "Schedule - GetSchedule"),
		slog.String("Class name:", className),
	)

	log.Info("getting weekly schedule")

	weekSchedule, err := s.scheduleProvider.GetWeekSchedule(ctx, className)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return nil, fmt.Errorf("failed to get the list of lessons with grades due to error: %w", err)
	}

	log.Info("weekly schedule successfully given")

	return weekSchedule, nil
}

func (s *ScheduleStore) SetHomework(ctx context.Context, className string, date string, lessonName string, homework string) error {
	log := s.log.With(
		slog.String("Operation:", "Schedule - SetHomework"),
		slog.String("Lesson name:", lessonName),
		slog.String("Homework:", homework),
	)

	log.Info("setting homework")

	err := s.scheduleProvider.SetHomework(ctx, className, date, lessonName, homework)
	if err != nil {
		log.Error("failed to set homework", sl.Err(err))

		return fmt.Errorf("failed to set homework due to error: %w", err)
	}

	log.Info("homework successfully set")

	return nil
}

func (s *ScheduleStore) SetWeeklySchedule(data []byte, sheet string) ([]models.StudentsScheduleDay, error) {
	log := s.log.With(
		slog.String("Operation:", "Schedule - SetWeeklySchedule"),
	)
	log.Info("setting weekly schedule")

	reader := bytes.NewReader(data)
	week, err := utils.GetWeekFromExcelFile(reader, sheet)
	if err != nil {
		log.Error("failed to set weekly schedule", sl.Err(err))
		return nil, err
	}
	log.Info("weekly schedule successfully set")

	return week, nil
}
