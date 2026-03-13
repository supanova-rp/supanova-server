package auth

import (
	"context"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"github.com/supanova-rp/supanova-server/internal/config"
)

type Token string

type AuthProvider struct {
	client *auth.Client
}

type User struct {
	ID      string
	IsAdmin bool
}

func New(ctx context.Context, credentials string) (*AuthProvider, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(credentials)))
	if err != nil {
		return nil, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return &AuthProvider{
		client: client,
	}, nil
}

func (a *AuthProvider) GetUserFromIDToken(ctx context.Context, accessToken string) (*User, error) {
	token, err := a.client.VerifyIDToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	userRecord, err := a.client.GetUser(ctx, token.UID)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:      userRecord.UID,
		IsAdmin: isAdmin(token),
	}, nil
}

func (a *AuthProvider) CreateUser(ctx context.Context, email, password, name string) (string, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password).
		DisplayName(name)

	userRecord, err := a.client.CreateUser(ctx, params)
	if err != nil {
		return "", err
	}

	return userRecord.UID, nil
}

func isAdmin(token *auth.Token) bool {
	adminValue, ok := token.Claims[string(config.AdminRole)]
	if !ok {
		return false
	}

	isAdmin, ok := adminValue.(bool)
	return ok && isAdmin
}
