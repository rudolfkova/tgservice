// Package handler реализует gRPC-обработчики сервиса.
package handler

import (
	"context"
	"log/slog"
	"tgservice/internal/model"
	tgservicev1 "tgservice/proto/tgservice/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SessionUsecase ...
type SessionUsecase interface {
	CreateSession(ctx context.Context) (sessionID, qrCode string, err error)
	DeleteSession(ctx context.Context, sessionID string) error
}

// MessageUsecase ...
type MessageUsecase interface {
	SendMessage(ctx context.Context, sessionID, peer, text string) (messageID int64, err error)
	// SubscribeMessages возвращает канал входящих сообщений для сессии.
	// Канал закрывается, когда сессия удаляется или ctx отменяется.
	SubscribeMessages(ctx context.Context, sessionID string) (<-chan model.MessageDTO, error)
}

type serverAPI struct {
	tgservicev1.UnimplementedTelegramServiceServer
	session SessionUsecase
	message MessageUsecase
	logger  *slog.Logger
}

// Register регистрирует gRPC-сервер.
func Register(gRPCServer *grpc.Server, session SessionUsecase, message MessageUsecase, logger *slog.Logger) {
	tgservicev1.RegisterTelegramServiceServer(gRPCServer, &serverAPI{
		session: session,
		message: message,
		logger:  logger,
	})
}

// CreateSession запускает авторизацию через QR-код и возвращает session_id и данные QR.
func (s *serverAPI) CreateSession(ctx context.Context, req *tgservicev1.CreateSessionRequest) (*tgservicev1.CreateSessionResponse, error) {
	sessionID, qrCode, err := s.session.CreateSession(ctx)
	if err != nil {
		s.logger.Error("CreateSession failed", slog.String("err", err.Error()))
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	s.logger.Info("CreateSession success", slog.String("session_id", sessionID))

	return &tgservicev1.CreateSessionResponse{
		SessionId: sessionID,
		QrCode:    qrCode,
	}, nil
}

// DeleteSession останавливает клиент и освобождает ресурсы сессии.
func (s *serverAPI) DeleteSession(ctx context.Context, req *tgservicev1.DeleteSessionRequest) (*tgservicev1.DeleteSessionResponse, error) {
	if req.GetSessionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	if err := s.session.DeleteSession(ctx, req.GetSessionId()); err != nil {
		s.logger.Error("DeleteSession failed",
			slog.String("session_id", req.GetSessionId()),
			slog.String("err", err.Error()),
		)
		return nil, status.Errorf(codes.Internal, "failed to delete session: %v", err)
	}

	s.logger.Info("DeleteSession success", slog.String("session_id", req.GetSessionId()))

	return &tgservicev1.DeleteSessionResponse{}, nil
}

// SendMessage отправляет текстовое сообщение через указанную сессию.
func (s *serverAPI) SendMessage(ctx context.Context, req *tgservicev1.SendMessageRequest) (*tgservicev1.SendMessageResponse, error) {
	switch {
	case req.GetSessionId() == "":
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	case req.GetPeer() == "":
		return nil, status.Error(codes.InvalidArgument, "peer is required")
	case req.GetText() == "":
		return nil, status.Error(codes.InvalidArgument, "text is required")
	}

	msgID, err := s.message.SendMessage(ctx, req.GetSessionId(), req.GetPeer(), req.GetText())
	if err != nil {
		s.logger.Error("SendMessage failed",
			slog.String("session_id", req.GetSessionId()),
			slog.String("peer", req.GetPeer()),
			slog.String("err", err.Error()),
		)
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	s.logger.Info("SendMessage success",
		slog.String("session_id", req.GetSessionId()),
		slog.Int64("message_id", msgID),
	)

	return &tgservicev1.SendMessageResponse{MessageId: msgID}, nil
}

// SubscribeMessages стримит входящие сообщения для указанной сессии.
// Поток закрывается, когда клиент отключается или сессия удаляется.
func (s *serverAPI) SubscribeMessages(req *tgservicev1.SubscribeMessagesRequest, stream grpc.ServerStreamingServer[tgservicev1.MessageUpdate]) error {
	if req.GetSessionId() == "" {
		return status.Error(codes.InvalidArgument, "session_id is required")
	}

	ctx := stream.Context()

	msgCh, err := s.message.SubscribeMessages(ctx, req.GetSessionId())
	if err != nil {
		s.logger.Error("SubscribeMessages failed",
			slog.String("session_id", req.GetSessionId()),
			slog.String("err", err.Error()),
		)
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	s.logger.Info("SubscribeMessages started", slog.String("session_id", req.GetSessionId()))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("SubscribeMessages: client disconnected",
				slog.String("session_id", req.GetSessionId()),
			)
			return nil

		case msg, ok := <-msgCh:
			if !ok {
				s.logger.Info("SubscribeMessages: session closed",
					slog.String("session_id", req.GetSessionId()),
				)
				return status.Error(codes.Unavailable, "session was closed")
			}

			if err := stream.Send(&tgservicev1.MessageUpdate{
				MessageId: msg.ID,
				From:      msg.From,
				Text:      msg.Text,
				Timestamp: msg.Timestamp,
			}); err != nil {
				s.logger.Error("SubscribeMessages: send failed",
					slog.String("session_id", req.GetSessionId()),
					slog.String("err", err.Error()),
				)
				return status.Errorf(codes.Internal, "failed to send update: %v", err)
			}
		}
	}
}
