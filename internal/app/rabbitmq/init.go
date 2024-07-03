package rabbitmq

import (
	"NotificationMS/internal/service/notification"
	"fmt"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
)

type Server struct {
	log          *slog.Logger
	conn         *amqp091.Connection
	channel      *amqp091.Channel
	notifService *notification.NotificationService
}

func New(
	log *slog.Logger,
	amqpPath string,
	notifService *notification.NotificationService,
) *Server {
	connectRabbitMQ, err := amqp091.Dial(amqpPath)
	if err != nil {
		panic(err)
	}

	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}

	return &Server{
		log:          log,
		conn:         connectRabbitMQ,
		channel:      channelRabbitMQ,
		notifService: notifService,
	}
}

func (s *Server) Start() error {
	messages, err := s.channel.Consume(
		"NotificationService", // queue name
		"GradeService",        // consumer
		true,                  // auto-ack
		false,                 // exclusive
		false,                 // no local
		false,                 // no wait
		nil,                   // arguments
	)
	if err != nil {
		return err
	}

	go func() {
		for message := range messages {
			if err := s.notifService.ConsumeNotification(message); err != nil {
				s.log.Error(err.Error())
			}
			fmt.Printf(" > Received message: %s\n", message.Body)
		}
	}()

	// Build a welcome message.
	s.log.Info("Successfully connected to RabbitMQ")
	s.log.Info("Waiting for messages")

	return nil
}

func (s *Server) Stop() {
	op := "rabbitmq.Stop"
	s.log.With(slog.String("op", op)).
		Info("stopping amqp")
	s.conn.Close()
	s.channel.Close()
}
