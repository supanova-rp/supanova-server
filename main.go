package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JDGarner/go-template/internal/config"
	"github.com/JDGarner/go-template/internal/server"
	"github.com/JDGarner/go-template/internal/store"
)

func main() {
	err := run()
	if err != nil {
		slog.Error("run failed", slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("shutting down gracefully...")
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.ParseEnv()
	if err != nil {
		return fmt.Errorf("unable to parse env: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	db, err := store.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer db.Close()

	svr := server.New(db, cfg.Port)
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
