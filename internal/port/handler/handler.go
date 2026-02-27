// Package handler ...
package handler

import (
	"context"
	"log/slog"
	"tgservice/internal/model"
	tgservicev1 "tgservice/proto/tgservice/v1"

	"google.golang.org/grpc"
)

// Session ...
type SeesionUsecase interface {
	CreateSession() (seesionID, QRCode string, err error)
	DeleteSession(sessionID string) (err error)
}

// Message ...
type MessageUsecase interface {
	SendMessage(sessionID, peer, test string) (messageID int, err error)
	SubscribeMessages(sessionID string) (messageDTO model.MessageDTO, err error)
}

type serverAPI struct {
	tgservicev1.UnimplementedTelegramServiceServer
	session SeesionUsecase
	message MessageUsecase
	logger  *slog.Logger
}

// Register ...
func Register(gRPCServer *grpc.Server, session SeesionUsecase, message MessageUsecase) {
	tgservicev1.RegisterTelegramServiceServer(gRPCServer, &serverAPI{session: session, message: message})
}

func (s *serverAPI) CreateSession(ctx context.Context, req *tgservicev1.CreateSessionRequest) (*tgservicev1.CreateSessionResponse, error) {
	return &tgservicev1.CreateSessionResponse{}, nil
}

func (s *serverAPI) DeleteSession(ctx context.Context, req *tgservicev1.DeleteSessionRequest) (*tgservicev1.DeleteSessionResponse, error) {
	return &tgservicev1.DeleteSessionResponse{}, nil
}

func (s *serverAPI) SendMessage(ctx context.Context, req *tgservicev1.SendMessageRequest) (*tgservicev1.SendMessageResponse, error) {
	return &tgservicev1.SendMessageResponse{}, nil
}

func (s *serverAPI) SubscribeMessages(req *tgservicev1.SubscribeMessagesRequest, stream grpc.ServerStreamingServer[tgservicev1.MessageUpdate]) error {
	return nil
}
