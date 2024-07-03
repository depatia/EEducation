package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

func Init(ctx context.Context) (*messaging.Client, error) {
	opt := option.WithCredentialsFile("../firebase-notif.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		fmt.Println("Unable to Connect To Firebase", err)
		return nil, err
	}

	fcmClient, err := app.Messaging(ctx)
	if err != nil {
		fmt.Println("Unable to Connect To Firebase", err)
		return nil, err
	}

	return fcmClient, nil
}
