package grpc

import (
	"LessonsMS/internal/models"
	"LessonsMS/internal/pb"
	"LessonsMS/internal/services/lesson"
	"LessonsMS/internal/storage/sterrors"
	"LessonsMS/internal/utils"
	"LessonsMS/tools"
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Lessoner interface {
	CreateNewLesson(
		ctx context.Context,
		teacherID int64,
		className string,
		lessonName string,
	) error
	GetLessons(
		ctx context.Context,
		teacherID int64,
	) ([]*models.Lesson, error)
	IsLessonWithClassExists(
		ctx context.Context,
		lessonName string,
		className string,
	) (bool, error)
	IsExistsByClassname(
		ctx context.Context,
		className string,
	) (bool, error)
}

type api struct {
	pb.UnimplementedLessonServer
	lessoner Lessoner
}

func Register(gRPCServer *grpc.Server, lessoner Lessoner) {
	pb.RegisterLessonServer(gRPCServer, &api{lessoner: lessoner})
}

func (a *api) SetLesson(ctx context.Context, req *pb.SetLessonRequest) (*pb.SetLessonResponse, error) {
	if req.TeacherId == 0 {
		return &pb.SetLessonResponse{Status: http.StatusBadRequest}, status.Error(codes.InvalidArgument, "teacher ID is required")
	}

	if req.LessonName == "" {
		return &pb.SetLessonResponse{Status: http.StatusBadRequest}, status.Error(codes.InvalidArgument, "lesson name is required")
	}

	if req.ClassName == "" {
		return &pb.SetLessonResponse{Status: http.StatusBadRequest}, status.Error(codes.InvalidArgument, "classname is required")
	}

	if ok := tools.CheckClass(req.ClassName); !ok {
		return &pb.SetLessonResponse{Status: http.StatusBadRequest}, status.Error(codes.InvalidArgument, "classname is incorrect")
	}

	err := a.lessoner.CreateNewLesson(ctx, req.TeacherId, req.ClassName, req.LessonName)

	if err != nil {
		if errors.Is(err, sterrors.ErrLessonAlreadyExists) {
			return &pb.SetLessonResponse{Status: http.StatusConflict}, status.Error(codes.AlreadyExists, "lesson already exists")
		}
		return &pb.SetLessonResponse{Status: http.StatusInternalServerError}, status.Error(codes.Internal, "failed to set the new lesson")
	}

	return &pb.SetLessonResponse{Status: http.StatusCreated}, nil
}

func (a *api) GetAllTeacherLessons(ctx context.Context, req *pb.GetAllTeacherLessonsRequest) (*pb.GetAllTeacherLessonsResponse, error) {
	if req.TeacherId == 0 {
		return nil, status.Error(codes.InvalidArgument, "teacher ID is required")
	}

	lessons, err := a.lessoner.GetLessons(ctx, req.TeacherId)

	if err != nil {
		if errors.Is(err, lesson.ErrIncorrectTeacherID) {
			return nil, status.Error(codes.NotFound, "incorrect teacher ID")
		}

		return nil, status.Error(codes.Internal, "failed to get the lessons")
	}

	return &pb.GetAllTeacherLessonsResponse{Lessons: utils.LessonsMapper(lessons)}, nil
}

func (a *api) IsLessonClassCombinationExists(ctx context.Context, req *pb.IsLessonClassCombinationExistsRequest) (*pb.IsLessonClassCombinationExistsResponse, error) {
	if req.LessonName == "" {
		return nil, status.Error(codes.InvalidArgument, "lesson name is required")
	}

	if req.ClassName == "" {
		return nil, status.Error(codes.InvalidArgument, "classname is required")
	}

	exists, err := a.lessoner.IsLessonWithClassExists(ctx, req.LessonName, req.ClassName)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check if the lesson exists")
	}
	return &pb.IsLessonClassCombinationExistsResponse{Exists: exists}, nil
}

func (a *api) IsClassExists(ctx context.Context, req *pb.IsClassExistsRequest) (*pb.IsClassExistsResponse, error) {
	if req.ClassName == "" {
		return nil, status.Error(codes.InvalidArgument, "classname is required")
	}

	exists, err := a.lessoner.IsExistsByClassname(ctx, req.ClassName)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check if the class exists")
	}
	return &pb.IsClassExistsResponse{Exists: exists}, nil
}
