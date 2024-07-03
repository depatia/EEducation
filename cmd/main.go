package main

import (
	"NotificationMS/internal/app"
	"NotificationMS/internal/config"
	"context"
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

	ctx := context.Background()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	application := app.New(log, cfg.DBPath, cfg.DBName, cfg.AMQPUrl, ctx)

	err = application.Server.Start()
	if err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Server.Stop()
	log.Info("Gracefully stopped")
}
