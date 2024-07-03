package app

import (
	"GradesMS/internal/app/grpc"
	"GradesMS/internal/services/grade"
	"GradesMS/internal/storage/postgre"
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

	gradesService := grade.New(log, storage, storage, storage)

	if err != nil {
		log.Error("failed to connect with lessons service")
	}

	grpcApp := grpc.New(log, gradesService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
