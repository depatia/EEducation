package grpc

import (
	usergrpc "AuthService/internal/grpc"
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCApp struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func NewGRPCApp(log *slog.Logger, authService usergrpc.AuthRepo, userService usergrpc.UserRepo, port int) *GRPCApp {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			//logging.StartCall, logging.FinishCall,
			logging.PayloadReceived, logging.PayloadSent,
		),
		// Add any other option (check functions starting with logging.With).
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
	))

	usergrpc.Register(gRPCServer, authService, userService)

	return &GRPCApp{gRPCServer: gRPCServer, port: port, log: log}
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (a *GRPCApp) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("failed to listen tcp due to error: %w", err)
	}

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to serve listener. due to error: %w", err)
	}

	return nil
}

func (a *GRPCApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *GRPCApp) Stop() {
	a.gRPCServer.GracefulStop()
}
