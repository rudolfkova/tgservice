package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tgservice/internal/driver"
	"tgservice/internal/port"
	"tgservice/internal/port/handler"
	"tgservice/internal/usecase"
	"tgservice/pkg/logger"
)

func main() {
	logger := logger.NewLogger()
	cfg, err := parseConfig()
	if err != nil {
		logger.Error("config error", slog.String("err", err.Error()))
		os.Exit(1)
	}

	drv := driver.NewDriver(cfg.TGAppID, cfg.TGAppHash, logger)

	sessionUC := usecase.NewSessionUsecase(drv, logger)
	messageUC := usecase.NewMessageUsecase(drv, logger)

	srv := port.NewServer(logger)
	handler.Register(srv.GRPC(), sessionUC, messageUC, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Run(cfg.BindAddr); err != nil {
			logger.Error("server error", slog.String("err", err.Error()))
		}
	}()

	<-ctx.Done()

	logger.Info("Shutdown signal received")

	done := make(chan struct{})
	go func() {
		srv.GRPC().GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Server gracefully stopped")
	case <-time.After(10 * time.Second):
		logger.Warn("Graceful shutdown timeout, forcing stop")
		srv.GRPC().Stop()
	}
}
