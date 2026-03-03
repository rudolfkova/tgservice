package usecase_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	tgerror "tgservice/internal/error"
	"tgservice/internal/mocks"
	"tgservice/internal/usecase"
)

var logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

func TestCreateSession_Success(t *testing.T) {
	driverMock := mocks.NewTelegramDriver(t)
	driverMock.
		On("Connect", mock.Anything).
		Return("session-123", "tg://login?token=abc", nil)

	uc := usecase.NewSessionUsecase(driverMock, logger)

	sessionID, qrCode, err := uc.CreateSession(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "session-123", sessionID)
	assert.Equal(t, "tg://login?token=abc", qrCode)
}

func TestCreateSession_DriverError(t *testing.T) {
	driverMock := mocks.NewTelegramDriver(t)
	driverMock.
		On("Connect", mock.Anything).
		Return("", "", errors.New("telegram unavailable"))

	uc := usecase.NewSessionUsecase(driverMock, logger)

	_, _, err := uc.CreateSession(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "driver.Connect")
}

func TestDeleteSession_Success(t *testing.T) {
	driverMock := mocks.NewTelegramDriver(t)
	driverMock.
		On("Disconnect", mock.Anything, "session-123").
		Return(nil)

	uc := usecase.NewSessionUsecase(driverMock, logger)

	err := uc.DeleteSession(context.Background(), "session-123")
	require.NoError(t, err)
}

func TestDeleteSession_NotFound(t *testing.T) {
	driverMock := mocks.NewTelegramDriver(t)
	driverMock.
		On("Disconnect", mock.Anything, "session-123").
		Return(tgerror.ErrSessionNotFound)

	uc := usecase.NewSessionUsecase(driverMock, logger)

	err := uc.DeleteSession(context.Background(), "session-123")
	require.Error(t, err)
	assert.ErrorIs(t, err, tgerror.ErrSessionNotFound)
}

func TestDeleteSession_DriverError(t *testing.T) {
	driverMock := mocks.NewTelegramDriver(t)
	driverMock.
		On("Disconnect", mock.Anything, "session-123").
		Return(errors.New("unexpected error"))

	uc := usecase.NewSessionUsecase(driverMock, logger)

	err := uc.DeleteSession(context.Background(), "session-123")
	require.Error(t, err)
	assert.ErrorContains(t, err, "driver.Disconnect")
}
