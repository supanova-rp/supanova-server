package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/middleware"
	"github.com/supanova-rp/supanova-server/internal/services/auth"
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

func New(h *handlers.Handlers, authProvider *auth.AuthProvider, cfg *config.App) *Server {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	e.HideBanner = true // Prevents startup banner from being logged

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: cfg.ClientURLs,
	}))

	// limits each unique IP to 60 requests per minute with a burst of 120.
	e.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStoreWithConfig(
		echoMiddleware.RateLimiterMemoryStoreConfig{
			Rate:      serverRateLimit,
			Burst:     serverBurstLimit,
			ExpiresIn: time.Minute,
		},
	)))

	if cfg.Environment == config.EnvironmentTest {
		e.Use(middleware.TestAuthMiddleware)
	} else {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return middleware.AuthMiddleware(next, authProvider)
		})
	}

	registerRoutes(e, h)

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

func registerRoutes(e *echo.Echo, h *handlers.Handlers) {
	e.GET(getRoute(config.APIVersion, "health"), h.HealthCheck)
	RegisterCourseRoutes(e, h, config.APIVersion)
	RegisterProgressRoutes(e, h, config.APIVersion)
	RegisterMediaRoutes(e, h, config.APIVersion)
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
