package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/middleware"
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

func New(h *handlers.Handlers, authProvider middleware.AuthProvider, cfg *config.App) *Server {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	e.HideBanner = true // Prevents startup banner from being logged

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: cfg.ClientURLs,
	}))

	e.Use(middleware.LoggingMiddleware)

	// limits each unique IP to 60 requests per minute with a burst of 120.
	e.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStoreWithConfig(
		echoMiddleware.RateLimiterMemoryStoreConfig{
			Rate:      serverRateLimit,
			Burst:     serverBurstLimit,
			ExpiresIn: time.Minute,
		},
	)))

	// Public route (no auth middleware)
	public := e.Group("/" + config.APIVersion)
	// Private group with auth middleware
	private := e.Group("/" + config.APIVersion)

	if cfg.Environment == config.EnvironmentTest {
		private.Use(middleware.TestAuthMiddleware)
	} else {
		private.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return middleware.AuthMiddleware(next, authProvider)
		})
	}

	registerRoutes(private, public, h)

	return &Server{
		echo: e,
		port: cfg.Port,
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

func registerRoutes(private, public *echo.Group, h *handlers.Handlers) {
	public.GET("/health", h.HealthCheck)

	RegisterCourseRoutes(private, h)
	RegisterProgressRoutes(private, h)
	RegisterMediaRoutes(private, h)
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}
