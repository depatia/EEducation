package main

import (
	"AuthService/internal/app"
	"AuthService/internal/config"
	"AuthService/pkg/tools/jwt"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Errorf("failed to load config due to error: %w", err)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	jwt := jwt.JwtWrapper{
		SecretKey:       cfg.JWTSecretKey,
		Issuer:          "go-grpc-auth-svc",
		ExpirationHours: 24 * 180,
	}

	application := app.New(log, jwt, cfg.Port, cfg.DBUrl)

	go func() {
		application.GRPCServer.MustRun()
	}()

	log.Info("server is running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()
}
