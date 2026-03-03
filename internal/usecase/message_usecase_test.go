package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"tgservice/internal/mocks"
	"tgservice/internal/model"
	"tgservice/internal/usecase"
)

func TestSendMessage_Success(t *testing.T) {
	messengerMock := mocks.NewTelegramMessenger(t)
	messengerMock.
		On("Send", mock.Anything, "session-123", "@durov", "hello").
		Return(int64(42), nil)

	uc := usecase.NewMessageUsecase(messengerMock, logger)

	msgID, err := uc.SendMessage(context.Background(), "session-123", "@durov", "hello")
	require.NoError(t, err)
	assert.Equal(t, int64(42), msgID)
}

func TestSendMessage_MessengerError(t *testing.T) {
	messengerMock := mocks.NewTelegramMessenger(t)
	messengerMock.
		On("Send", mock.Anything, "session-123", "@durov", "hello").
		Return(int64(0), errors.New("not authorized"))

	uc := usecase.NewMessageUsecase(messengerMock, logger)

	_, err := uc.SendMessage(context.Background(), "session-123", "@durov", "hello")
	require.Error(t, err)
	assert.ErrorContains(t, err, "messenger.Send")
}

func TestSubscribeMessages_Success(t *testing.T) {
	messengerMock := mocks.NewTelegramMessenger(t)

	ch := make(chan model.MessageDTO, 1)
	messengerMock.
		On("Messages", mock.Anything, "session-123").
		Return((<-chan model.MessageDTO)(ch), nil)

	uc := usecase.NewMessageUsecase(messengerMock, logger)

	result, err := uc.SubscribeMessages(context.Background(), "session-123")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSubscribeMessages_MessengerError(t *testing.T) {
	messengerMock := mocks.NewTelegramMessenger(t)
	messengerMock.
		On("Messages", mock.Anything, "session-123").
		Return(nil, errors.New("session not found"))

	uc := usecase.NewMessageUsecase(messengerMock, logger)

	_, err := uc.SubscribeMessages(context.Background(), "session-123")
	require.Error(t, err)
	assert.ErrorContains(t, err, "messenger.Messages")
}
