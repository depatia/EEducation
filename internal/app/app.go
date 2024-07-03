package app

import (
	_ "APIGateway/docs"
	gradeService "APIGateway/internal/clients/grade_service"
	lessonService "APIGateway/internal/clients/lesson_service"
	notification "APIGateway/internal/clients/notification_service"
	scheduleService "APIGateway/internal/clients/schedule_service"
	userService "APIGateway/internal/clients/user_service"
	"APIGateway/internal/config"
	"APIGateway/internal/delivery/http/auth"
	"APIGateway/internal/delivery/http/grade"
	"APIGateway/internal/delivery/http/lesson"
	"APIGateway/internal/delivery/http/schedule"
	"APIGateway/internal/delivery/http/user"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	server *http.Server
	Router *httprouter.Router
	log    *slog.Logger
	cfg    *config.Config
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.Config,
) *Server {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/swagger/:any", httpSwagger.WrapHandler)

	notifClient, err := notification.New(ctx, log, cfg.Clients.Notification.Addr, time.Duration(cfg.Clients.Notification.Timeout)*time.Second, cfg.Clients.Notification.RetriesCount)
	if err != nil {
		log.Error("failed to connect with notification service")
	}
	lessonClient, err := lessonService.New(ctx, log, cfg.Clients.Lesson.Addr, time.Duration(cfg.Clients.Lesson.Timeout)*time.Second, cfg.Clients.Lesson.RetriesCount)
	if err != nil {
		log.Error("failed to connect with lesson service")
	}
	lessonHandler := lesson.Handler{Service: lessonClient}
	lessonHandler.Register(router)

	gradeClient, err := gradeService.New(ctx, log, cfg.Clients.Grade.Addr, time.Duration(cfg.Clients.Grade.Timeout)*time.Second, cfg.Clients.Grade.RetriesCount, notifClient)
	if err != nil {
		log.Error("failed to connect with grade service")
	}
	gradeHandler := grade.Handler{Service: gradeClient}
	gradeHandler.Register(router)

	scheduleClient, err := scheduleService.New(ctx, log, cfg.Clients.Schedule.Addr, time.Duration(cfg.Clients.Schedule.Timeout)*time.Second, cfg.Clients.Schedule.RetriesCount)
	if err != nil {
		log.Error("failed to connect with schedule service")
	}
	scheduleHandler := schedule.Handler{Service: scheduleClient}
	scheduleHandler.Register(router)

	userClient, err := userService.New(ctx, log, cfg.Clients.User.Addr, time.Duration(cfg.Clients.User.Timeout)*time.Second, cfg.Clients.User.RetriesCount)
	if err != nil {
		log.Error("failed to connect with user service")
	}
	authHandler := auth.Handler{AuthService: userClient}
	userHandler := user.Handler{UserService: userClient}
	authHandler.Register(router)
	userHandler.Register(router)

	return &Server{Router: router, cfg: cfg, log: log}

}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.cfg.Port))
	if err != nil {
		s.log.Error(err.Error())
	}

	s.server = &http.Server{
		Handler:      s.Router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	s.log.Info("application started")

	if err := s.server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			s.log.Info("server shutdown")
		default:
			s.log.Error(err.Error())
		}
	}
}

func (s *Server) Stop() {
	const op = "grpcapp.Stop"

	s.log.With(slog.String("op", op)).
		Info("stopping http server", slog.Int("port", s.cfg.Port))

	s.server.Close()
}
