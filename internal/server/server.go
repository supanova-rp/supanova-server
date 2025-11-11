package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/JDGarner/go-template/internal/handlers"
	"github.com/JDGarner/go-template/internal/store"
)

const (
	serverTimeout    = 10 * time.Second
	serverRateLimit  = 60
	serverBurstLimit = 120
)

type Server struct {
	echo *echo.Echo
	port string
}

func New(s *store.Store, port string) *Server {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}

	// limits each unique IP to 60 requests per minute with a burst of 120.
	config := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      serverRateLimit,
		Burst:     serverBurstLimit,
		ExpiresIn: time.Minute,
	})
	e.Use(middleware.RateLimiter(config))

	h := &handlers.Handlers{
		Store: s,
	}
	registerRoutes(e, h)

	return &Server{
		echo: e,
		port: port,
	}
}

func (s *Server) Start() error {
	err := s.echo.Start(":" + s.port)
	if err != nil && err != http.ErrServerClosed {
		slog.Error("server error", slog.Any("error", err))
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverTimeout)
	defer cancel()

	return s.echo.Shutdown(shutdownCtx)
}

func registerRoutes(e *echo.Echo, h *handlers.Handlers) {
	e.GET("/health", h.HealthCheck)
	e.GET("/item/:id", h.GetItem)
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}
