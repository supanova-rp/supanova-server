package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("failed to load .env file", slog.Any("error", err))
	}
	creds := os.Getenv("FIREBASE_CREDENTIALS")

	var makeAdmin bool
	flag.BoolVar(&makeAdmin, "make-admin", true, "Make admin if true/remove admin status if false")
	flag.Parse()

	if err := run(creds, makeAdmin); err != nil {
		slog.Error("run failed", slog.Any("error", err))
	}
}

func run(creds string, makeAdmin bool) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the email of the user:\n")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email input %v", err)
	}
	email = strings.TrimSpace(email)

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(creds)))
	if err != nil {
		return fmt.Errorf("failed to initialise firebase app %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth client %v", err)
	}

	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to fetch user %v", err)
	}
	slog.Info("successfully fetched user data", slog.String("email", email), slog.String("id", user.UID))

	err = client.SetCustomUserClaims(ctx, user.UID, map[string]any{
		"admin": makeAdmin,
	})
	if err != nil {
		return fmt.Errorf("failed to update user admin status %v", err)
	}

	slog.Info("successfully updated user admin status", slog.String("email", email), slog.Bool("make_admin", makeAdmin))
	return nil
}
