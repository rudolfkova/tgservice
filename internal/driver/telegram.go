// Package driver ...
package driver

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"tgservice/internal/model"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
)

type client struct {
	tg         *telegram.Client
	cancel     context.CancelFunc
	msgCh      chan model.MessageDTO
	authorized bool
}

// Driver ...
type Driver struct {
	appID   int
	appHash string
	logger  *slog.Logger

	mu      sync.RWMutex
	clients map[string]*client
}

// NewDriver ...
func NewDriver(appID int, appHash string, logger *slog.Logger) *Driver {
	return &Driver{
		appID:   appID,
		appHash: appHash,
		logger:  logger,
		clients: make(map[string]*client),
	}
}

// Connect ...
func (d *Driver) Connect(ctx context.Context) (sessionID string, qrLink string, err error) {
	sessionID = uuid.New().String()

	dispatcher := tg.NewUpdateDispatcher()

	tgClient := telegram.NewClient(d.appID, d.appHash, telegram.Options{
		UpdateHandler: dispatcher,
	})
	qrCh := make(chan string, 1)
	errCh := make(chan error, 1)
	msgCh := make(chan model.MessageDTO, 100)

	sessionCtx, cancel := context.WithCancel(context.Background())

	loggedIn := qrlogin.OnLoginToken(dispatcher)

	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok || msg.Out {
			return nil
		}

		select {
		case msgCh <- model.MessageDTO{
			ID:        int64(msg.ID),
			From:      resolveFrom(msg, e),
			Text:      msg.Message,
			Timestamp: int64(msg.Date),
		}:
		default:
			d.logger.Warn("msgCh full, dropping message",
				slog.String("session_id", sessionID),
				slog.Int("msg_id", msg.ID),
			)
		}

		return nil
	})

	go func() {
		runErr := tgClient.Run(sessionCtx, func(ctx context.Context) error {
			qr := tgClient.QR()

			authorization, authErr := qr.Auth(ctx, loggedIn, func(ctx context.Context, token qrlogin.Token) error {
				select {
				case qrCh <- token.URL():
				default:
					d.logger.Info("qr token refreshed", slog.String("session_id", sessionID))
				}
				return nil
			})
			if authErr != nil {
				return fmt.Errorf("qr auth: %w", authErr)
			}

			user, ok := authorization.User.AsNotEmpty()
			if !ok {
				return fmt.Errorf("empty user after authorization")
			}

			d.mu.Lock()
			if c, ok := d.clients[sessionID]; ok {
				c.authorized = true
			}
			d.mu.Unlock()

			d.logger.Info("session authorized",
				slog.String("session_id", sessionID),
				slog.Int64("user_id", user.ID),
				slog.String("username", user.Username),
			)

			<-ctx.Done()
			return nil
		})

		if runErr != nil {
			select {
			case errCh <- runErr:
			default:
			}
		}

		close(msgCh)
	}()

	// Ждём первую QR-ссылку, ошибку запуска или отмену запроса.
	select {
	case link := <-qrCh:
		qrLink = link
	case runErr := <-errCh:
		cancel()
		return "", "", fmt.Errorf("client run: %w", runErr)
	case <-ctx.Done():
		cancel()
		return "", "", ctx.Err()
	}

	d.mu.Lock()
	d.clients[sessionID] = &client{
		tg:     tgClient,
		cancel: cancel,
		msgCh:  msgCh,
	}
	d.mu.Unlock()

	d.logger.Info("client connected", slog.String("session_id", sessionID))

	return sessionID, qrLink, nil
}

// resolveFrom извлекает строковый идентификатор отправителя из сообщения.
func resolveFrom(msg *tg.Message, e tg.Entities) string {
	peer, ok := msg.GetFromID()
	if !ok {
		return "unknown"
	}

	switch p := peer.(type) {
	case *tg.PeerUser:
		if u, ok := e.Users[p.UserID]; ok && u.Username != "" {
			return "@" + u.Username
		}
		return fmt.Sprintf("%d", p.UserID)
	case *tg.PeerChannel:
		if ch, ok := e.Channels[p.ChannelID]; ok && ch.Username != "" {
			return "@" + ch.Username
		}
		return fmt.Sprintf("channel:%d", p.ChannelID)
	default:
		return "unknown"
	}
}
