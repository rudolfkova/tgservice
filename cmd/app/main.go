package main

import (
	"log/slog"
	"os"

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

	if err := srv.Run(cfg.BindAddr); err != nil {
		logger.Error("server error", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
