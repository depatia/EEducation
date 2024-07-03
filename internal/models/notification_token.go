package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationToken struct {
	ID        primitive.ObjectID `json:"id"  bson:"id"`
	UserID    string             `json:"user_id"  bson:"user_id"`
	DeviceID  string             `json:"device_id"  bson:"device_id"`
	Timestamp time.Time          `json:"timestamp"  bson:"timestamp"`
}
