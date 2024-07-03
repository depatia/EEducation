package lesson_service

import (
	clienterrors "APIGateway/internal/client_errors"
	"APIGateway/pkg/tools/logger/sl"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/depatia/EEducation-Protos/gen/lesson"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	api lesson.LessonClient
	log *slog.Logger
}

type LessonService interface {
	SetLesson(ctx context.Context, req *SetLessonReq) (int64, error)
	GetAllTeacherLessons(ctx context.Context, teacherID int64) ([]*lesson.TeacherLesson, error)
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
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

	cc, err := grpc.Dial("127.0.0.1:8082",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	grpcClient := lesson.NewLessonClient(cc)

	return &Client{
		api: grpcClient,
		log: log,
	}, nil
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) IsClassExists(ctx context.Context, className string) (bool, error) {
	const op = "grpc.lesson.IsClassExists"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("TeacherID", className),
	)

	log.Info("checking is class exists")

	if className == "" {
		log.Error("failed to check is class exists", sl.Err(clienterrors.ErrAllFieldsRequired))

		return false, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.IsClassExists(ctx, &lesson.IsClassExistsRequest{
		ClassName: className,
	})
	if err != nil {
		log.Error("failed to check is class exists", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked")

	return resp.Exists, nil
}

func (c *Client) GetAllTeacherLessons(ctx context.Context, teacherID int64) ([]*lesson.TeacherLesson, error) {
	const op = "grpc.lesson.GetAllTeacherLessons"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("TeacherID", teacherID),
	)

	log.Info("getting all teacher lessons")

	if teacherID == 0 {
		log.Error("failed to get lessons", sl.Err(clienterrors.ErrAllFieldsRequired))

		return nil, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.GetAllTeacherLessons(ctx, &lesson.GetAllTeacherLessonsRequest{
		TeacherId: teacherID,
	})
	if err != nil {
		log.Error("failed to get lessons", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("lessons successfully given")

	return resp.Lessons, nil
}

func (c *Client) SetLesson(ctx context.Context, req *SetLessonReq) (int64, error) {
	const op = "grpc.lesson.SetLesson"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("TeacherID", req.TeacherID),
	)

	log.Info("creating lesson")

	if req.ClassName == "" || req.LessonName == "" || req.TeacherID == 0 {
		log.Error("failed to set lesson", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.SetLesson(ctx, &lesson.SetLessonRequest{
		TeacherId:  req.TeacherID,
		LessonName: req.LessonName,
		ClassName:  req.ClassName,
	})

	if err != nil {
		log.Error("failed to set lesson", sl.Err(err))

		if status.Code(err) == codes.AlreadyExists {
			return http.StatusConflict, fmt.Errorf("%s: %w", op, err)
		} else if status.Code(err) == codes.InvalidArgument {
			return http.StatusBadRequest, fmt.Errorf("%s: %w", op, err)
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("lesson created")

	return resp.Status, nil
}
