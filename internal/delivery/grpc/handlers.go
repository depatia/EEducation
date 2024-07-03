package grpc

import (
	"ScheduleMS/internal/models"
	"ScheduleMS/internal/pb"
	"ScheduleMS/internal/storage/sterrors"
	"ScheduleMS/internal/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScheduleRepo interface {
	SetSchedule(
		ctx context.Context,
		day models.StudentsScheduleDay,
	) error
	GetSchedule(
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
	SetWeeklySchedule(
		data []byte,
		sheet string,
	) ([]models.StudentsScheduleDay, error)
}

type LessonRepo interface {
	IsLessonExists(ctx context.Context, lessonName string, className string) (bool, error)
	IsClassExists(ctx context.Context, className string) (bool, error)
}

type api struct {
	pb.UnimplementedScheduleServer
	scheduleRepo ScheduleRepo
	lessonRepo   LessonRepo
}

func Register(gRPCServer *grpc.Server, scheduleRepo ScheduleRepo, lessonRepo LessonRepo) {
	pb.RegisterScheduleServer(gRPCServer, &api{scheduleRepo: scheduleRepo, lessonRepo: lessonRepo})
}

func (a *api) SetSchedule(ctx context.Context, req *pb.SetScheduleRequest) (*pb.SetScheduleResponse, error) {
	if req.ClassName == "" {
		return nil, status.Error(codes.InvalidArgument, "class name is required")
	}

	if err := utils.ValidatePbLessons(req.ScheduleDay.Lessons); err != nil {
		return nil, err
	}

	if req.ScheduleDay.Date == "" {
		req.ScheduleDay.Date = time.Now().Format(time.DateOnly)
	}

	for _, lesson := range req.ScheduleDay.Lessons {
		ok, _ := a.lessonRepo.IsLessonExists(ctx, lesson.LessonName, req.ClassName)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "no such class name - lesson combination")
		}
	}

	day := utils.DayMapper(req.ScheduleDay, req.ClassName)

	err := a.scheduleRepo.SetSchedule(ctx, day)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set the daily schedule")
	}
	return &pb.SetScheduleResponse{Status: http.StatusCreated}, nil
}

func (a *api) GetWeekScheduleByClass(ctx context.Context, req *pb.GetWeekScheduleByClassRequest) (*pb.GetWeekScheduleByClassResponse, error) {
	if req.ClassName == "" {
		return nil, status.Error(codes.InvalidArgument, "class name is required")
	}

	ok, _ := a.lessonRepo.IsClassExists(ctx, req.ClassName)
	if !ok {
		return nil, status.Error(codes.NotFound, "no such class name - lesson combination")
	}

	schedule, err := a.scheduleRepo.GetSchedule(ctx, req.ClassName)

	if err != nil {
		if errors.Is(err, sterrors.ErrClassNotFound) {
			return nil, status.Error(codes.NotFound, "class with this name not found")
		}

		return nil, status.Error(codes.Internal, "failed to get the schedule")
	}

	return &pb.GetWeekScheduleByClassResponse{Schedule: utils.WeekMapper(schedule)}, nil
}

func (a *api) SetHomework(ctx context.Context, req *pb.SetHomeworkRequest) (*pb.SetHomeworkResponse, error) {
	if req.ClassName == "" {
		return nil, status.Error(codes.InvalidArgument, "class name is required")
	}
	if req.Homework == "" {
		return nil, status.Error(codes.InvalidArgument, "homework is required")
	}
	if req.Date == "" {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}
	if req.LessonName == "" {
		return nil, status.Error(codes.InvalidArgument, "lesson name is required")
	}

	ok, err := a.lessonRepo.IsLessonExists(ctx, req.LessonName, req.ClassName)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check if lesson exists")
	}

	if !ok {
		return nil, status.Error(codes.NotFound, "no such class name - lesson combination")
	}

	err = a.scheduleRepo.SetHomework(ctx, req.ClassName, req.Date, req.LessonName, req.Homework)

	if err != nil {
		if errors.Is(err, sterrors.ErrScheduleNotFound) {
			return nil, status.Error(codes.NotFound, "lesson with this date or classroom not found")
		}

		return nil, status.Error(codes.Internal, "failed to set the homework")
	}
	return &pb.SetHomeworkResponse{Status: http.StatusOK}, nil
}

func (a *api) SetWeeklySchedule(stream pb.Schedule_SetWeeklyScheduleServer) error {
	ctx := context.Background()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stream.SendAndClose(&pb.SetWeeklyScheduleResponse{
				Status: http.StatusInternalServerError,
			})
		}
		week, err := a.scheduleRepo.SetWeeklySchedule(req.FileData, req.Sheet)
		if err != nil {
			return err
		}
		for _, day := range week {
			if day.ClassName == "" || day.Date == "" {
				return stream.SendAndClose(&pb.SetWeeklyScheduleResponse{
					Status: http.StatusBadRequest,
				})
			}
			if err := utils.ValidateLessons(day.Lessons); err != nil {
				fmt.Println(err)
				return stream.SendAndClose(&pb.SetWeeklyScheduleResponse{
					Status: http.StatusBadRequest,
				})
			}
			for _, lesson := range day.Lessons {
				ok, _ := a.lessonRepo.IsLessonExists(ctx, lesson.LessonName, day.ClassName)
				if !ok {
					return stream.SendAndClose(&pb.SetWeeklyScheduleResponse{
						Status: http.StatusBadRequest,
					})
				}
			}
			if err := a.scheduleRepo.SetSchedule(ctx, day); err != nil {
				return stream.SendAndClose(&pb.SetWeeklyScheduleResponse{
					Status: http.StatusInternalServerError,
				})
			}

		}
	}
	return stream.SendAndClose(&pb.SetWeeklyScheduleResponse{
		Status: http.StatusCreated,
	})
}
