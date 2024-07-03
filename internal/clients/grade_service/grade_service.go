package grade

import (
	clienterrors "APIGateway/internal/client_errors"
	notification "APIGateway/internal/clients/notification_service"
	"APIGateway/pkg/tools/logger/sl"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/depatia/EEducation-Protos/gen/grades"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const notif = "Вы получили новую оценку: %d по %s."

type Client struct {
	api   grades.GradesClient
	notif *notification.Client
	log   *slog.Logger
}

type GradeService interface {
	GetAllLessonsGradesByStudentID(ctx context.Context, studentID int64) ([]*grades.Grade, error)
	GetLessonGrades(ctx context.Context, lessonName string, studentID int64) ([]*grades.Grade, error)
	SetGrade(ctx context.Context, req *SetGradeReq) (int64, error)
	DeleteGrade(ctx context.Context, req *DeleteGradeReq) (int, error)
	ChangeGrade(ctx context.Context, req *ChangeGradeReq) (int, error)
	SetTermGrade(ctx context.Context, req *SetTermGradeReq) (int, error)
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
	notifClient *notification.Client,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	grpcClient := grades.NewGradesClient(cc)

	return &Client{
		api:   grpcClient,
		notif: notifClient,
		log:   log,
	}, nil
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) GetAllLessonsGradesByStudentID(ctx context.Context, studentID int64) ([]*grades.Grade, error) {
	const op = "grpc.grade.GetAllLessonsGradesByStudentID"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", studentID),
	)

	log.Info("getting all grades")

	if studentID == 0 {
		log.Error("failed to get grades", sl.Err(clienterrors.ErrAllFieldsRequired))

		return nil, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.GetAllLessonsGradesByStudentID(ctx, &grades.GetAllLessonsGradesByStudentIDRequest{
		StudentId: studentID,
	})
	if err != nil {
		log.Error("failed to get grades", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grades successfully given")

	return resp.Grades, nil
}

func (c *Client) GetLessonGrades(ctx context.Context, lessonName string, studentID int64) ([]*grades.Grade, error) {
	const op = "grpc.grade.GetLessonGrades"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", studentID),
		slog.String("Lesson name", lessonName),
	)

	log.Info("getting lesson's grades")

	if studentID == 0 || lessonName == "" {
		log.Error("failed to get grades", sl.Err(clienterrors.ErrAllFieldsRequired))

		return nil, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.GetLessonGrades(ctx, &grades.GetLessonGradesRequest{
		StudentId:  studentID,
		LessonName: lessonName,
	})
	if err != nil {
		log.Error("failed to get grades", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grades successfully given")

	return resp.Grades, nil
}

func (c *Client) SetGrade(ctx context.Context, req *SetGradeReq) (int64, error) {
	const op = "grpc.grade.SetGrade"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.StudentID),
		slog.String("Lesson name", req.LessonName),
	)

	log.Info("creating grade")

	if req.StudentID == 0 || req.LessonName == "" || req.Grade == 0 || req.DeviceID == 0 {
		log.Error("failed to create grade", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.SetGrade(ctx, &grades.SetGradeRequest{
		StudentId:  req.StudentID,
		LessonName: req.LessonName,
		Grade:      req.Grade,
		IsTerm:     false,
	})
	if err != nil {
		log.Error("failed to create grade", sl.Err(err))

		if status.Code(err) == codes.AlreadyExists {
			return http.StatusConflict, fmt.Errorf("%s: %w", op, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.notif.SendNotification(ctx, strconv.Itoa(int(req.StudentID)), strconv.Itoa(req.DeviceID), fmt.Sprintf(notif, req.Grade, req.LessonName)); err != nil {
		log.Error("notification wasnt sent", sl.Err(err))

		return http.StatusInternalServerError, nil
	}

	log.Info("grade successfully created")

	return resp.Status, nil
}

func (c *Client) ChangeGrade(ctx context.Context, req *ChangeGradeReq) (int, error) {
	const op = "grpc.grade.ChangeGrade"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.StudentID),
		slog.String("Lesson name", req.LessonName),
	)

	log.Info("changing grade")

	if req.StudentID == 0 || req.LessonName == "" || req.Date == "" || req.Grade == 0 {
		log.Error("failed to change grade", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.ChangeGrade(ctx, &grades.ChangeGradeRequest{
		StudentId:  req.StudentID,
		LessonName: req.LessonName,
		Date:       req.Date,
		Grade:      req.Grade,
	})
	if err != nil {
		log.Error("failed to change grade", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grade successfully changed")

	return int(resp.Status), nil
}

func (c *Client) DeleteGrade(ctx context.Context, req *DeleteGradeReq) (int, error) {
	const op = "grpc.grade.DeleteGrade"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.StudentID),
		slog.String("Lesson name", req.LessonName),
	)

	log.Info("deleting grade")

	if req.StudentID == 0 || req.LessonName == "" || req.Date == "" {
		log.Error("failed to delete grade", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.DeleteGrade(ctx, &grades.DeleteGradeRequest{
		StudentId:  req.StudentID,
		LessonName: req.LessonName,
		Date:       req.Date,
	})
	if err != nil {
		log.Error("failed to delete grade", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grade successfully deleted")

	return int(resp.Status), nil
}

func (c *Client) SetTermGrade(ctx context.Context, req *SetTermGradeReq) (int, error) {
	const op = "grpc.grade.SetTermGrade"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.UserID),
	)

	log.Info("setting term grade")

	if req.UserID == 0 || req.LessonName == "" {
		log.Error("failed to set term grade", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.SetTermGrade(ctx, &grades.SetTermGradeRequest{
		StudentId:  req.UserID,
		LessonName: req.LessonName,
	})
	if err != nil {
		log.Error("failed to set term grade", sl.Err(err))

		if status.Code(err) == codes.Unavailable {
			return http.StatusNotAcceptable, fmt.Errorf("%s: %w", op, err)
		} else if status.Code(err) == codes.AlreadyExists {
			return http.StatusConflict, fmt.Errorf("%s: %w", op, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grade successfully set")

	return int(resp.Status), nil
}
