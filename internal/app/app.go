package app

import (
	"context"
	"log/slog"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/middleware"
	"github.com/supanova-rp/supanova-server/internal/server"
	"github.com/supanova-rp/supanova-server/internal/services/cron"
	"github.com/supanova-rp/supanova-server/internal/store"
)

type Dependencies struct {
	Store            *store.Store
	ObjectStorage    handlers.ObjectStorage
	EmailService     handlers.EmailService
	AuthProvider     middleware.AuthProvider
	EmailFailureCron *cron.Cron
}

func Run(ctx context.Context, cfg *config.App, deps Dependencies) (err error) {
	h := handlers.NewHandlers(
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

	// TODO: move this into email service?
	cancelEmailFailureCron, err := deps.EmailService.SetupRetry()
	defer cancelEmailFailureCron()
	if err != nil {
		return err
	}

	// blocks until signal received (e.g. by ctrl+C or process killed) OR server error
	select {
	case <-ctx.Done():
		slog.Info("context cancelled")
	case svrErr := <-serverErr:
		err = svrErr
	}

	cancelEmailFailureCron() // cancel cron contexts to prevent new jobs from starting

	stopEmailFailureCtx := deps.EmailFailureCron.Stop() // returns a context that waits until existing cron jobs finish
	<-stopEmailFailureCtx.Done()
	slog.Info("email failure cron jobs completed") // TODO: Take this out?

	shutdownErr := svr.Stop()
	if shutdownErr != nil {
		slog.Error("server shutdown error", slog.Any("error", shutdownErr))
	}

	return err
}
