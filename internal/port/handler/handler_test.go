package handler_test

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"tgservice/internal/mocks"
	"tgservice/internal/model"
	"tgservice/internal/port/handler"
	tgservicev1 "tgservice/proto/tgservice/v1"
)

const testAddr = ":50052"

func newTestServer(t *testing.T, session handler.SessionUsecase, message handler.MessageUsecase) tgservicev1.TelegramServiceClient {
	t.Helper()

	lis, err := net.Listen("tcp", testAddr)
	require.NoError(t, err)

	srv := grpc.NewServer()
	handler.Register(srv, session, message, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	conn, err := grpc.NewClient(testAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return tgservicev1.NewTelegramServiceClient(conn)
}

func TestCreateSession_Success(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	sessionMock.
		On("CreateSession", mock.Anything).
		Return("session-123", "tg://login?token=abc", nil)

	client := newTestServer(t, sessionMock, messageMock)

	resp, err := client.CreateSession(context.Background(), &tgservicev1.CreateSessionRequest{})
	require.NoError(t, err)
	assert.Equal(t, "session-123", resp.SessionId)
	assert.Equal(t, "tg://login?token=abc", resp.QrCode)
}

func TestCreateSession_UsecaseError(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	sessionMock.
		On("CreateSession", mock.Anything).
		Return("", "", errors.New("telegram unavailable"))

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.CreateSession(context.Background(), &tgservicev1.CreateSessionRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestDeleteSession_Success(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	sessionMock.
		On("DeleteSession", mock.Anything, "session-123").
		Return(nil)

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.DeleteSession(context.Background(), &tgservicev1.DeleteSessionRequest{
		SessionId: "session-123",
	})
	require.NoError(t, err)
}

func TestDeleteSession_EmptySessionID(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.DeleteSession(context.Background(), &tgservicev1.DeleteSessionRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestDeleteSession_UsecaseError(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	sessionMock.
		On("DeleteSession", mock.Anything, "session-123").
		Return(errors.New("session not found"))

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.DeleteSession(context.Background(), &tgservicev1.DeleteSessionRequest{
		SessionId: "session-123",
	})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestSendMessage_Success(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	messageMock.
		On("SendMessage", mock.Anything, "session-123", "@durov", "hello").
		Return(int64(42), nil)

	client := newTestServer(t, sessionMock, messageMock)

	resp, err := client.SendMessage(context.Background(), &tgservicev1.SendMessageRequest{
		SessionId: "session-123",
		Peer:      "@durov",
		Text:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(42), resp.MessageId)
}

func TestSendMessage_EmptySessionID(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.SendMessage(context.Background(), &tgservicev1.SendMessageRequest{
		Peer: "@durov",
		Text: "hello",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSendMessage_EmptyPeer(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.SendMessage(context.Background(), &tgservicev1.SendMessageRequest{
		SessionId: "session-123",
		Text:      "hello",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSendMessage_EmptyText(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.SendMessage(context.Background(), &tgservicev1.SendMessageRequest{
		SessionId: "session-123",
		Peer:      "@durov",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSendMessage_UsecaseError(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	messageMock.
		On("SendMessage", mock.Anything, "session-123", "@durov", "hello").
		Return(int64(0), errors.New("not authorized"))

	client := newTestServer(t, sessionMock, messageMock)

	_, err := client.SendMessage(context.Background(), &tgservicev1.SendMessageRequest{
		SessionId: "session-123",
		Peer:      "@durov",
		Text:      "hello",
	})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestSubscribeMessages_EmptySessionID(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	client := newTestServer(t, sessionMock, messageMock)

	stream, err := client.SubscribeMessages(context.Background(), &tgservicev1.SubscribeMessagesRequest{})
	require.NoError(t, err)

	_, err = stream.Recv()
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSubscribeMessages_UsecaseError(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	messageMock.
		On("SubscribeMessages", mock.Anything, "session-123").
		Return(nil, errors.New("session not found"))

	client := newTestServer(t, sessionMock, messageMock)

	stream, err := client.SubscribeMessages(context.Background(), &tgservicev1.SubscribeMessagesRequest{
		SessionId: "session-123",
	})
	require.NoError(t, err)

	_, err = stream.Recv()
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestSubscribeMessages_ReceivesMessages(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	ch := make(chan model.MessageDTO, 2)
	ch <- model.MessageDTO{ID: 1, From: "@alice", Text: "hi", Timestamp: 1000}
	ch <- model.MessageDTO{ID: 2, From: "@bob", Text: "hey", Timestamp: 2000}

	messageMock.
		On("SubscribeMessages", mock.Anything, "session-123").
		Return((<-chan model.MessageDTO)(ch), nil)

	client := newTestServer(t, sessionMock, messageMock)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stream, err := client.SubscribeMessages(ctx, &tgservicev1.SubscribeMessagesRequest{
		SessionId: "session-123",
	})
	require.NoError(t, err)

	msg1, err := stream.Recv()
	require.NoError(t, err)
	assert.Equal(t, int64(1), msg1.MessageId)
	assert.Equal(t, "@alice", msg1.From)
	assert.Equal(t, "hi", msg1.Text)

	msg2, err := stream.Recv()
	require.NoError(t, err)
	assert.Equal(t, int64(2), msg2.MessageId)
	assert.Equal(t, "@bob", msg2.From)
	assert.Equal(t, "hey", msg2.Text)
}

func TestSubscribeMessages_ChannelClosed(t *testing.T) {
	sessionMock := mocks.NewSessionUsecase(t)
	messageMock := mocks.NewMessageUsecase(t)

	ch := make(chan model.MessageDTO)
	close(ch)

	messageMock.
		On("SubscribeMessages", mock.Anything, "session-123").
		Return((<-chan model.MessageDTO)(ch), nil)

	client := newTestServer(t, sessionMock, messageMock)

	stream, err := client.SubscribeMessages(context.Background(), &tgservicev1.SubscribeMessagesRequest{
		SessionId: "session-123",
	})
	require.NoError(t, err)

	_, err = stream.Recv()
	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err))
}
