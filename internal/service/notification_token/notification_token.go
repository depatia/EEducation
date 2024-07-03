package notificationtoken

import (
	"NotificationMS/internal/models"
	"context"
	"log/slog"
	"time"
)

type NotifTokenService struct {
	log                *slog.Logger
	notifTokenSaver    NotificationTokenSaver
	notifTokenProvider NotificationTokenProvider
}

type NotificationTokenSaver interface {
	RegisterDevice(ctx context.Context, token *models.NotificationToken) error
}

type NotificationTokenProvider interface {
	GetNotificationTokens(ctx context.Context, userID string) ([]string, error)
}

func New(
	log *slog.Logger,
	notifTokenSaver NotificationTokenSaver,
	notifTokenProvider NotificationTokenProvider,
) *NotifTokenService {
	return &NotifTokenService{
		log:                log,
		notifTokenSaver:    notifTokenSaver,
		notifTokenProvider: notifTokenProvider,
	}
}

func (n *NotifTokenService) GetNotificationTokens(ctx context.Context, userID string) ([]string, error) {
	return n.notifTokenProvider.GetNotificationTokens(ctx, userID)
}

func (n *NotifTokenService) RegisterDevice(ctx context.Context, userID string, deviceID string) error {
	return n.notifTokenSaver.RegisterDevice(ctx, &models.NotificationToken{
		UserID:    userID,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
	})
}
