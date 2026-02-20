package app

import (
	"context"
	"log/slog"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/middleware"
	"github.com/supanova-rp/supanova-server/internal/server"
	"github.com/supanova-rp/supanova-server/internal/store"
)

type Dependencies struct {
	Store         *store.Store
	ObjectStorage handlers.ObjectStorage
	EmailService  handlers.EmailService
	AuthProvider  middleware.AuthProvider
}

func Run(ctx context.Context, cfg *config.App, deps Dependencies) (err error) {
	h := handlers.NewHandlers(
		deps.Store,
		deps.Store,
		deps.Store,
		deps.Store,
		deps.Store,
		deps.Store,
		deps.ObjectStorage,
		deps.EmailService,
	)

	svr := server.New(h, deps.AuthProvider, cfg)
	serverErr := make(chan error, 1)

	go func() {
		serverErr <- svr.Start()
	}()

	// blocks until signal received (e.g. by ctrl+C or process killed) OR server error
	select {
	case <-ctx.Done():
		slog.Info("context cancelled")
	case svrErr := <-serverErr:
		err = svrErr
	}

	deps.EmailService.StopRetry(ctx)

	shutdownErr := svr.Stop()
	if shutdownErr != nil {
		slog.Error("server shutdown error", slog.Any("error", shutdownErr))
	}

	return err
}
