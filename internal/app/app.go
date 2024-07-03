package app

import (
	"LessonsMS/internal/app/grpc"
	"LessonsMS/internal/services/lesson"
	"LessonsMS/internal/storage/postgre"
	"log/slog"
)

type App struct {
	GRPCServer *grpc.GRPCApp
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
) *App {
	storage, err := postgre.New(storagePath)
	if err != nil {
		panic(err)
	}

	lessonService := lesson.New(log, storage, storage)

	grpcApp := grpc.New(log, lessonService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
