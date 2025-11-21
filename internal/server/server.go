package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/supanova-rp/supanova-server/internal/handlers"
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

func New(h *handlers.Handlers, port string) *Server {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	e.HideBanner = true // Prevents startup banner from being logged

	// limits each unique IP to 60 requests per minute with a burst of 120.
	config := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      serverRateLimit,
		Burst:     serverBurstLimit,
		ExpiresIn: time.Minute,
	})
	e.Use(middleware.RateLimiter(config))

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
	e.GET(getRoute("v2", "health"), h.HealthCheck)
	e.GET(getRoute("v2", "course/:id"), h.GetCourse)
}

func getRoute(prefix, route string) string {
	return fmt.Sprintf("/%s/%s", prefix, route)
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}
