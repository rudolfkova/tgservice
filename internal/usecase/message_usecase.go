// Package usecase ...
package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"tgservice/internal/model"
)

// TelegramMessenger ...
type TelegramMessenger interface {
	// Send ...
	Send(ctx context.Context, clientID, peer, text string) (messageID int64, err error)

	// Messages ...
	Messages(ctx context.Context, clientID string) (<-chan model.MessageDTO, error)
}

// MessageUsecase ...
type MessageUsecase struct {
	messenger TelegramMessenger
	logger    *slog.Logger
}

// NewMessageUsecase ...
func NewMessageUsecase(messenger TelegramMessenger, logger *slog.Logger) *MessageUsecase {
	return &MessageUsecase{
		messenger: messenger,
		logger:    logger,
	}
}

// SendMessage ...
func (uc *MessageUsecase) SendMessage(ctx context.Context, sessionID, peer, text string) (int64, error) {
	msgID, err := uc.messenger.Send(ctx, sessionID, peer, text)
	if err != nil {
		return 0, fmt.Errorf("messenger.Send: %w", err)
	}

	uc.logger.Info("message sent",
		slog.String("session_id", sessionID),
		slog.String("peer", peer),
		slog.Int64("message_id", msgID),
	)

	return msgID, nil
}

// SubscribeMessages ...
func (uc *MessageUsecase) SubscribeMessages(ctx context.Context, sessionID string) (<-chan model.MessageDTO, error) {
	ch, err := uc.messenger.Messages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("messenger.Messages: %w", err)
	}

	uc.logger.Info("subscribed to messages", slog.String("session_id", sessionID))

	return ch, nil
}
