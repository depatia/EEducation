package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateNotifReq struct {
	UserID       string `json:"userID"`
	Notification string `json:"notification"`
	DeviceID     string `json:"deviceID"  bson:"device_id"`
}

type Notification struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	UserID       string             `json:"user_id" bson:"user_id"`
	Notification string             `json:"notification" bson:"notification"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
}
