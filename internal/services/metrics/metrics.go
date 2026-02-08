package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supanova-rp/supanova-server/internal/config"
)

const (
	// How long the server will wait to read the entire request after the connection is accepted
	readTimeout = 10 * time.Second

	// How long the server has to write the response after reading the request
	writeTimeout = 10 * time.Second

	// How long to keep a keep-alive connection open waiting for the next request
	idleTimeout = 120 * time.Second

	shutdownTimeout = 10 * time.Second
)

type Server struct {
	echo *echo.Echo
	port string
}

func New(cfg *config.Metrics) *Server {
	e := echo.New()
	e.HideBanner = true // Prevents startup banner from being logged

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: cfg.ClientURLs,
	}))

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.Server.ReadTimeout = readTimeout
	e.Server.WriteTimeout = writeTimeout
	e.Server.IdleTimeout = idleTimeout

	return &Server{
		echo: e,
		port: cfg.Port,
	}
}

func (s *Server) Start() error {
	if err := s.echo.Start(":" + s.port); err != nil {
		slog.Error("metrics server error", slog.Any("error", err))
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return s.echo.Shutdown(shutdownCtx)
}

func Run(ctx context.Context, cfg *config.Metrics) (err error) {
	svr := New(cfg)
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

	shutdownErr := svr.Stop()
	if shutdownErr != nil {
		slog.Error("metrics server shutdown error", slog.Any("error", shutdownErr))
	}

	return err
}
