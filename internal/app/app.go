package app

import (
	"AuthService/internal/app/grpc"
	"AuthService/internal/services/auth"
	"AuthService/internal/services/user"
	"AuthService/internal/storage/mysql"
	"AuthService/pkg/tools/jwt"
	"log/slog"
)

type App struct {
	GRPCServer *grpc.GRPCApp
}

func New(log *slog.Logger, wrapper jwt.JwtWrapper, port int, storagePath string) *App {
	storage, err := mysql.New(storagePath)

	if err != nil {
		panic(err)
	}

	authService := auth.New(wrapper, storage, storage, storage, storage, log)
	userService := user.New(log, storage, storage, storage)

	grpcApp := grpc.NewGRPCApp(log, authService, userService, port)

	return &App{GRPCServer: grpcApp}
}
