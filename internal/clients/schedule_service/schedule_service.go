package schedule

import (
	clienterrors "APIGateway/internal/client_errors"
	"APIGateway/pkg/tools/logger/sl"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/depatia/EEducation-Protos/gen/schedule"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	api schedule.ScheduleClient
	log *slog.Logger
}

type ScheduleService interface {
	GetWeekScheduleByClass(ctx context.Context, className string) (*schedule.ScheduleWeek, error)
	SetHomework(ctx context.Context, req *SetHomeworkReq) (int64, error)
	SetWeeklySchedule(ctx context.Context, req *SetWeeklyScheduleReq) (int64, error)
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

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	grpcClient := schedule.NewScheduleClient(cc)

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

func (c *Client) GetWeekScheduleByClass(ctx context.Context, className string) (*schedule.ScheduleWeek, error) {
	const op = "grpc.Schedule.GetWeekScheduleByClass"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("Classname", className),
	)

	log.Info("getting week schedule")

	if className == "" {
		log.Error("failed to get schedule", sl.Err(clienterrors.ErrAllFieldsRequired))

		return nil, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.GetWeekScheduleByClass(ctx, &schedule.GetWeekScheduleByClassRequest{
		ClassName: className,
	})
	if err != nil {
		log.Error("failed to get schedule", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("schedule successfully given")

	return resp.Schedule, nil
}

func (c *Client) SetHomework(ctx context.Context, req *SetHomeworkReq) (int64, error) {
	const op = "grpc.Schedule.SetHomework"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("Classname", req.Classname),
		slog.String("Homework", req.Homework),
	)

	log.Info("setting homework")

	if req.Classname == "" || req.LessonName == "" || req.Date == "" || req.Homework == "" {
		log.Error("failed to set homework", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.SetHomework(ctx, &schedule.SetHomeworkRequest{
		LessonName: req.LessonName,
		ClassName:  req.Classname,
		Date:       req.Date,
		Homework:   req.Homework,
	})
	if err != nil {
		log.Error("failed to set homework", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		} else if status.Code(err) == codes.InvalidArgument {
			return http.StatusBadRequest, fmt.Errorf("%s: %w", op, err)
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("homework successfully set")

	return resp.Status, nil
}

func (c *Client) SetWeeklySchedule(ctx context.Context, req *SetWeeklyScheduleReq) (int64, error) {
	const op = "grpc.Schedule.SetWeeklySchedule"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("File name", req.Filename),
	)

	log.Info("creating weekly schedule")

	if req.Filename == "" || req.Sheet == "" {
		log.Error("failed to create weekly schedule", sl.Err(clienterrors.ErrAllFieldsRequired))
		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	stream, err := c.api.SetWeeklySchedule(ctx)
	if err != nil {
		log.Error("failed to create weekly schedule", sl.Err(err))
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	if err := stream.Send(&schedule.SetWeeklyScheduleRequest{
		Filename: req.Filename,
		Sheet:    req.Sheet,
		FileData: req.FileData,
	}); err != nil {
		log.Error("failed to create weekly schedule", sl.Err(err))
		return http.StatusInternalServerError, err
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Error("failed to create weekly schedule", sl.Err(err))
		return http.StatusInternalServerError, err
	}

	log.Info("weekly schedule created")

	return reply.Status, err
}
