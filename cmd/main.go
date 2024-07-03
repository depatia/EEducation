package main

import (
	"APIGateway/internal/app"
	"APIGateway/internal/config"
	"APIGateway/pkg/tools/logger/sl"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// @title EEducation API service
// @version 1.0

// @securityDefinitions.apikey TokenAuth
// @in header
// @name Authorization

// @host localhost:1234
// @BasePath /
func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("failed to load config due to error: %w", sl.Err(err))
	}

	app := app.New(context.Background(), log, cfg)
	go func() {
		app.Start()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	app.Stop()
	log.Info("Gracefully stopped")
}
