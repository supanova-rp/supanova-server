package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	customConfig "github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/server"
	"github.com/supanova-rp/supanova-server/internal/services/objectstorage"
	"github.com/supanova-rp/supanova-server/internal/services/secrets"
	"github.com/supanova-rp/supanova-server/internal/store"
)

func main() {
	err := run()
	if err != nil {
		slog.Error("run failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("shutting down gracefully...")
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := customConfig.ParseEnv()
	if err != nil {
		return fmt.Errorf("failed to parse env: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer st.Close()

	awsCfg, err := newAWSCfg(ctx, cfg.AWS)
	if err != nil {
		return fmt.Errorf("failed to load aws config: %v", err)
	}

	secretsManager := secrets.New(ctx, awsCfg)

	CDNKey, err := secretsManager.Get(ctx, cfg.AWS.CDNKeyName)
	if err != nil {
		return fmt.Errorf("failed to fetch cdn key: %v", err)
	}

	objectStore, err := objectstorage.New(ctx, cfg.AWS, awsCfg, CDNKey)
	if err != nil {
		return fmt.Errorf("failed to create object store: %v", err)
	}

	h := handlers.NewHandlers(
		st,
		st,
		st,
		st,
		objectStore,
	)

	svr := server.New(h, cfg.Port, cfg.Environment)
	serverErr := make(chan error, 1)

	go func() {
		serverErr <- svr.Start()
	}()

	select {
	case <-ctx.Done(): // Blocks until server error OR signal received (e.g. by ctrl-C or process killed)
		slog.Info("context cancelled")
	case svrErr := <-serverErr:
		err = svrErr
	}

	shutdownErr := svr.Stop()
	if shutdownErr != nil {
		slog.Error("server shutdown error", slog.Any("error", shutdownErr))
	}

	return err
}

func newAWSCfg(ctx context.Context, cfg *customConfig.AWS) (*aws.Config, error) {
	newCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		))
	if err != nil {
		return nil, err
	}

	return &newCfg, nil
}
