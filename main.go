package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/supanova-rp/supanova-server/internal/app"
	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/services/objectstorage"
	"github.com/supanova-rp/supanova-server/internal/services/secrets"
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

	cfg, err := config.ParseEnv()
	if err != nil {
		return fmt.Errorf("failed to parse env: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	awsCfg, err := newAWSConfig(ctx, cfg.AWS)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	secretsManager := secrets.New(ctx, awsCfg)

	cdnKey, err := secretsManager.Get(ctx, cfg.AWS.CDNKeyName)
	if err != nil {
		return fmt.Errorf("failed to fetch CDN key: %v", err)
	}

	objectStore, err := objectstorage.New(ctx, cfg.AWS, awsCfg, cdnKey)
	if err != nil {
		return fmt.Errorf("failed to create object store: %v", err)
	}

	return app.Run(ctx, cfg, app.Dependencies{
		ObjectStorage: objectStore,
	})
}

func newAWSConfig(ctx context.Context, cfg *config.AWS) (*aws.Config, error) {
	awsCfg, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(cfg.Region),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		))
	if err != nil {
		return nil, err
	}

	return &awsCfg, nil
}
