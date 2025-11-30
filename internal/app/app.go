package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/server"
	"github.com/supanova-rp/supanova-server/internal/services/auth"
	"github.com/supanova-rp/supanova-server/internal/store"
)

type Dependencies struct {
	ObjectStorage handlers.ObjectStorage
	AuthProvider  *auth.AuthProvider
}

func Run(ctx context.Context, cfg *config.App, deps Dependencies) error {
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer st.Close()

	h := handlers.NewHandlers(
		st,
		st,
		st,
		st,
		deps.ObjectStorage,
	)

	svr := server.New(h, cfg.Port, cfg.Environment, deps.AuthProvider)
	serverErr := make(chan error, 1)

	go func() {
		serverErr <- svr.Start()
	}()

	select {
	case <-ctx.Done():
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
