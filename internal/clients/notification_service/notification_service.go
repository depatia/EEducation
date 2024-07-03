package notification

import (
	"APIGateway/pkg/tools/logger/sl"
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	log *slog.Logger
	ch  *amqp091.Channel
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "grpc.New"

	connectRabbitMQ, err := amqp091.Dial("amqp://guest:guest@127.0.0.53:5672/")
	if err != nil {
		panic(err)
	}

	// Let's start by opening a channel to our RabbitMQ
	// instance over the connection we have already
	// established.
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}

	// With the instance and declare Queues that we can
	// publish and subscribe to.
	_, err = channelRabbitMQ.QueueDeclare(
		"NotificationService", // queue name
		true,                  // durable
		false,                 // auto delete
		false,                 // exclusive
		false,                 // no wait
		nil,                   // arguments
	)
	if err != nil {
		panic(err)
	}

	return &Client{
		ch:  channelRabbitMQ,
		log: log,
	}, nil
}

func (c *Client) SendNotification(ctx context.Context, userID, deviceID, notification string) error {
	const op = "notification.SendNotification"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("UserID", userID),
	)

	log.Info("sending notification")

	message, err := json.Marshal(&NotifRequest{
		UserID:       userID,
		DeviceID:     deviceID,
		Notification: notification,
	})

	if err != nil {
		log.Error("failed to send notification", sl.Err(err))

		return err
	}

	if err := c.ch.Publish(
		"",                    // exchange
		"NotificationService", // queue name
		false,                 // mandatory
		false,                 // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	); err != nil {
		log.Error("failed to send notification", sl.Err(err))

		return err
	}

	log.Info("notification has been sent")

	return nil
}
