package auth

import (
	"context"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type AuthProvider struct {
	Client *firestore.Client
}

func New(ctx context.Context, credentials string) (*AuthProvider, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(credentials)))
	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close() //nolint:errcheck

	return &AuthProvider{
		Client: client,
	}, nil
}
