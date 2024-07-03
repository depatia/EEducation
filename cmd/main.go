package main

import (
	"ScheduleMS/internal/app"
	"ScheduleMS/internal/config"
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

	application := app.New(log, cfg.Port, cfg.DBPath, cfg.Clients)

	go func() {
		application.GRPCServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()
	log.Info("Gracefully stopped")
}
