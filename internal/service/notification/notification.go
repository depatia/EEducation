package notification

import (
	"NotificationMS/internal/models"
	notificationtoken "NotificationMS/internal/service/notification_token"
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/rabbitmq/amqp091-go"
)

type NotificationService struct {
	log               *slog.Logger
	notifSaver        NotificationSaver
	notifProvider     NotificationProvider
	notifTokenService *notificationtoken.NotifTokenService
	fcm               *messaging.Client
	ctx               context.Context
}

func New(
	log *slog.Logger,
	notifSaver NotificationSaver,
	notifProvider NotificationProvider,
	notifTokenService *notificationtoken.NotifTokenService,
	fcm *messaging.Client,
	ctx context.Context,
) *NotificationService {
	return &NotificationService{
		log:               log,
		notifSaver:        notifSaver,
		notifProvider:     notifProvider,
		notifTokenService: notifTokenService,
		fcm:               fcm,
		ctx:               ctx,
	}
}

type NotificationSaver interface {
	CreateNotification(ctx context.Context, motif models.Notification) error
}

type NotificationProvider interface {
	GetNotifications(ctx context.Context, userID string) ([]models.Notification, error)
}

func (n *NotificationService) CreateNotification(userID, notification string) error {
	return n.notifSaver.CreateNotification(n.ctx, models.Notification{
		UserID:       userID,
		Notification: notification,
		CreatedAt:    time.Now().UTC(),
	})
}

func (n *NotificationService) GetNotifications(userID string) ([]models.Notification, error) {
	return n.notifProvider.GetNotifications(n.ctx, userID)
}

func (n *NotificationService) SendNotification(token, message string) error {
	_, err := n.fcm.Send(n.ctx, &messaging.Message{
		Data:  map[string]string{message: message},
		Token: token,
	})

	return err
}

func (n *NotificationService) SendBatchNotifications(tokens []string, message string) error {
	_, err := n.fcm.SendMulticast(n.ctx, &messaging.MulticastMessage{
		Data:   map[string]string{message: message},
		Tokens: tokens,
	})

	return err
}

func (n *NotificationService) ConsumeNotification(nmsg amqp091.Delivery) error {
	var tkns []string
	var req models.CreateNotifReq

	err := json.Unmarshal(nmsg.Body, &req)
	if err != nil {
		return err
	}

	err = n.notifTokenService.RegisterDevice(n.ctx, req.UserID, req.DeviceID)
	if err != nil {
		return err
	}

	err = n.CreateNotification(req.UserID, req.Notification)
	if err != nil {
		return err
	}

	tkns, err = n.notifTokenService.GetNotificationTokens(n.ctx, req.UserID)
	if err != nil {
		return err
	}

	err = n.SendBatchNotifications(tkns, req.Notification)
	if err != nil {
		return err
	}

	return nil
}
