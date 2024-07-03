package mongo

import (
	"NotificationMS/internal/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type StDb struct {
	col *mongo.Collection
}

func New(ctx context.Context, path string) (*mongo.Client, error) {
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(path))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db due to error: %w", err)
	}

	if err = cli.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to connect to db due to error: %w", err)
	}

	return cli, nil
}

func InitMongo(col *mongo.Collection) *StDb {
	return &StDb{
		col: col,
	}
}

func (s *StDb) CreateNotification(ctx context.Context, notif models.Notification) error {
	const op = "storage.mongo.CreateNotification"
	notif.ID = primitive.NewObjectID()
	_, err := s.col.InsertOne(ctx, notif)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *StDb) GetNotifications(ctx context.Context, userID string) ([]models.Notification, error) {
	const op = "storage.mongo.GetNotifications"
	filter := bson.D{{Key: "user_id", Value: userID}}
	notifCursor, err := s.col.Find(ctx, filter)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer notifCursor.Close(ctx)

	notifications := make([]models.Notification, 0)
	for notifCursor.Next(ctx) {
		var notification models.Notification
		err = notifCursor.Decode(&notification)
		notifications = append(notifications, notification)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return notifications, nil
}

func (s *StDb) GetNotificationTokens(ctx context.Context, userID string) ([]string, error) {
	const op = "storage.mongo.GetNotificationTokens"
	filter := bson.D{{Key: "user_id", Value: userID}}
	notifCursor, err := s.col.Find(ctx, filter)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer notifCursor.Close(ctx)

	tokens := make([]string, 0)
	for notifCursor.Next(ctx) {
		var token models.NotificationToken
		err = notifCursor.Decode(&token)
		tokens = append(tokens, token.DeviceID)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return tokens, nil
}

func (s *StDb) RegisterDevice(ctx context.Context, token *models.NotificationToken) error {
	const op = "storage.mongo.RegisterDevice"
	filter := bson.D{{Key: "device_id", Value: token.DeviceID}}
	res := s.col.FindOne(ctx, filter)
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			token.ID = primitive.NewObjectID()
			_, err := s.col.InsertOne(ctx, token)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		}
		return fmt.Errorf("%s: %w", op, res.Err())
	}

	_, err := s.col.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"timestamp": time.Now().UTC()}})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
