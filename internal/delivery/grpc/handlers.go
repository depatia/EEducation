package grpc

import (
	"GradesMS/internal/pb"
	"GradesMS/internal/storage/sterrors"
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Grader interface {
	AddGrade(
		ctx context.Context,
		studentID int64,
		lessonName string,
		grade int64,
		isTerm bool,
	) (int64, error)
	GetLessonGradesByStudentID(
		ctx context.Context,
		studentID int64,
		lessonName string,
	) ([]*pb.Grade, error)
	GetAllGradesByStudentID(
		ctx context.Context,
		studentID int64,
	) ([]*pb.Grade, error)
	SetTermGrade(
		ctx context.Context,
		studentID int64,
		lessonName string,
	) error
	ChangeGrade(
		ctx context.Context,
		studentID int64,
		lessonName string,
		grade int64,
		date string,
	) error
	DeleteGrade(
		ctx context.Context,
		studentID int64,
		lessonName string,
		date string,
	) error
}

type api struct {
	pb.UnimplementedGradesServer
	grader Grader
}

func Register(gRPCServer *grpc.Server, grader Grader) {
	pb.RegisterGradesServer(gRPCServer, &api{grader: grader})
}

func (a *api) SetGrade(ctx context.Context, req *pb.SetGradeRequest) (*pb.SetGradeResponse, error) {
	if req.StudentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "studentID is required")
	}

	if req.LessonName == "" {
		return nil, status.Error(codes.InvalidArgument, "lesson name is required")
	}

	if req.Grade < 0 {
		return nil, status.Error(codes.InvalidArgument, "grade must be greater than or equal to 0")
	}

	_, err := a.grader.AddGrade(ctx, req.StudentId, req.LessonName, req.Grade, req.IsTerm)

	if err != nil {
		if errors.Is(err, sterrors.ErrGradeAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "grade already exists, try to edit the grade")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.SetGradeResponse{Status: http.StatusCreated}, nil
}

func (a *api) GetAllLessonsGradesByStudentID(ctx context.Context, req *pb.GetAllLessonsGradesByStudentIDRequest) (*pb.GetAllLessonsGradesByStudentIDResponse, error) {
	if req.StudentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "studentID is required")
	}

	grades, err := a.grader.GetAllGradesByStudentID(ctx, req.StudentId)

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.GetAllLessonsGradesByStudentIDResponse{Grades: grades}, nil
}

func (a *api) GetLessonGrades(ctx context.Context, req *pb.GetLessonGradesRequest) (*pb.GetLessonGradesResponse, error) {
	if req.StudentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "studentID is required")
	}

	grades, err := a.grader.GetLessonGradesByStudentID(ctx, req.StudentId, req.LessonName)

	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.GetLessonGradesResponse{Grades: grades}, nil
}

func (a *api) ChangeGrade(ctx context.Context, req *pb.ChangeGradeRequest) (*pb.ChangeGradeResponse, error) {
	if req.StudentId == 0 {
		return nil, status.Error(codes.Internal, "student id is required")
	}

	if req.LessonName == "" {
		return nil, status.Error(codes.Internal, "lesson name is required")
	}

	if req.Date == "" {
		return nil, status.Error(codes.Internal, "date is required")
	}

	if req.Grade == 0 {
		return nil, status.Error(codes.Internal, "grade is required")
	}

	err := a.grader.ChangeGrade(ctx, req.StudentId, req.LessonName, req.Grade, req.Date)

	if err != nil {
		if errors.Is(err, sterrors.ErrGradeNotFound) {
			return nil, status.Error(codes.NotFound, "grade not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.ChangeGradeResponse{Status: http.StatusOK}, nil
}

func (a *api) DeleteGrade(ctx context.Context, req *pb.DeleteGradeRequest) (*pb.DeleteGradeResponse, error) {
	if req.StudentId == 0 {
		return nil, status.Error(codes.Internal, "student id is required")
	}

	if req.LessonName == "" {
		return nil, status.Error(codes.Internal, "lesson name is required")
	}

	if req.Date == "" {
		return nil, status.Error(codes.Internal, "date is required")
	}

	err := a.grader.DeleteGrade(ctx, req.StudentId, req.LessonName, req.Date)

	if err != nil {
		if errors.Is(err, sterrors.ErrGradeNotFound) {
			return nil, status.Error(codes.NotFound, "grade not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.DeleteGradeResponse{Status: http.StatusOK}, nil
}

func (a *api) SetTermGrade(ctx context.Context, req *pb.SetTermGradeRequest) (*pb.SetTermGradeResponse, error) {
	if req.StudentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "student id is required")
	}

	if req.LessonName == "" {
		return nil, status.Error(codes.InvalidArgument, "lesson name is required")
	}

	err := a.grader.SetTermGrade(ctx, req.StudentId, req.LessonName)

	if err != nil {
		if errors.Is(err, sterrors.ErrNotEnoughGrades) {
			return nil, status.Error(codes.Unavailable, "not enough grades")
		} else if errors.Is(err, sterrors.ErrGradeAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "grade already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")

	}
	return &pb.SetTermGradeResponse{Status: http.StatusCreated}, status.Error(codes.OK, "created")
}
