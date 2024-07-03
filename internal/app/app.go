package app

import (
	"ScheduleMS/internal/app/grpc"
	lessongrpc "ScheduleMS/internal/clients/lessons/grpc"
	"ScheduleMS/internal/config"
	"ScheduleMS/internal/services/schedule"
	"ScheduleMS/internal/storage/postgre"
	"context"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpc.GRPCApp
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	clientCfg config.ClientsConfig,
) *App {
	storage, err := postgre.New(storagePath)
	if err != nil {
		panic(err)
	}

	scheduleService := schedule.New(log, storage, storage)
	lessonService, err := lessongrpc.New(
		context.Background(),
		log,
		clientCfg.Lessons.Addr,
		time.Duration(clientCfg.Lessons.Timeout)*time.Second,
		clientCfg.Lessons.RetriesCount,
	)

	if err != nil {
		log.Error("failed to connect with lessons service")
	}

	grpcApp := grpc.New(log, scheduleService, lessonService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
