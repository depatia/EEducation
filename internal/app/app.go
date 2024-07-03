package app

import (
	"NotificationMS/internal/app/firebase"
	"NotificationMS/internal/app/rabbitmq"
	"NotificationMS/internal/service/notification"
	notificationtoken "NotificationMS/internal/service/notification_token"
	"NotificationMS/internal/storage/mongo"
	"context"
	"log/slog"
)

type App struct {
	Server *rabbitmq.Server
}

func New(
	log *slog.Logger,
	storagePath string,
	storageName string,
	amqpPath string,
	ctx context.Context,
) *App {
	mngCli, err := mongo.New(ctx, storagePath)
	if err != nil {
		panic(err)
	}
	mngNotif := mongo.InitMongo(mngCli.Database(storageName).Collection("Notifications"))
	mngNotifTokens := mongo.InitMongo(mngCli.Database(storageName).Collection("NotificationTokens"))

	fb, err := firebase.Init(ctx)

	if err != nil {
		panic(err)
	}

	notifTokenService := notificationtoken.New(log, mngNotifTokens, mngNotifTokens)

	notifService := notification.New(log, mngNotif, mngNotif, notifTokenService, fb, ctx)

	amqpApp := rabbitmq.New(log, amqpPath, notifService)

	return &App{
		Server: amqpApp,
	}
}
