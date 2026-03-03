// Package usecase ...
package usecase

import (
	"context"
	"fmt"
	"log/slog"

	tgerror "tgservice/internal/error"
)

// TelegramDriver ...
type TelegramDriver interface {
	// Connect ...
	Connect(ctx context.Context) (clientID string, qrLink string, err error)

	// Disconnect ...
	Disconnect(ctx context.Context, clientID string) error
}

// SessionUsecase ...
type SessionUsecase struct {
	driver TelegramDriver
	logger *slog.Logger
}

// NewSessionUsecase ...
func NewSessionUsecase(driver TelegramDriver, logger *slog.Logger) *SessionUsecase {
	return &SessionUsecase{
		driver: driver,
		logger: logger,
	}
}

// CreateSession ...
func (uc *SessionUsecase) CreateSession(ctx context.Context) (sessionID, qrCode string, err error) {
	clientID, qrLink, err := uc.driver.Connect(ctx)
	if err != nil {
		return "", "", fmt.Errorf("driver.Connect: %w", err)
	}

	uc.logger.Info("session created", slog.String("session_id", clientID))

	return clientID, qrLink, nil
}

// DeleteSession ...
func (uc *SessionUsecase) DeleteSession(ctx context.Context, sessionID string) error {
	if err := uc.driver.Disconnect(ctx, sessionID); err != nil {
		return fmt.Errorf("driver.Disconnect: %w", tgerror.Wrap(err))
	}

	uc.logger.Info("session deleted", slog.String("session_id", sessionID))

	return nil
}
