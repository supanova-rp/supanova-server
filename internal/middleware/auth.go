package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/services/auth"
)

//go:generate moq -out ./mocks/authprovider_mock.go -pkg mocks . AuthProvider

type AuthProvider interface {
	GetUserFromIDToken(ctx context.Context, token string) (*auth.User, error)
}

type ContextKey string

const (
	UserIDContextKey ContextKey = "userID"
	RoleContextKey   ContextKey = "role"
)

type AuthParams struct {
	AccessToken string `json:"access_token" validate:"required"`
}

var nonAdminPaths = []string{
	fmt.Sprintf("/%s/course", config.APIVersion),
	fmt.Sprintf("/%s/assigned-course-titles", config.APIVersion),
	fmt.Sprintf("/%s/get-progress", config.APIVersion),
	fmt.Sprintf("/%s/update-progress", config.APIVersion),
	fmt.Sprintf("/%s/set-intro-completed", config.APIVersion),
	fmt.Sprintf("/%s/set-course-completed", config.APIVersion),
	fmt.Sprintf("/%s/video-url", config.APIVersion),
	fmt.Sprintf("/%s/materials", config.APIVersion),
	fmt.Sprintf("/%s/get-quiz-state", config.APIVersion),
	fmt.Sprintf("/%s/set-quiz-state", config.APIVersion),
	fmt.Sprintf("/%s/increment-attempts", config.APIVersion),
}

func AuthMiddleware(next echo.HandlerFunc, authProvider *auth.AuthProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		// Restore body so handler can use c.Bind()
		c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var params AuthParams
		if err := json.Unmarshal(bodyBytes, &params); err != nil {
			slog.Error("failed to unmarshal auth middleware request body", slog.Any("error", err))
			return echo.NewHTTPError(http.StatusUnauthorized, errors.Unauthorised)
		}

		user, err := authProvider.GetUserFromIDToken(ctx, params.AccessToken)
		if err != nil {
			slog.Error(err.Error())
			return echo.NewHTTPError(http.StatusUnauthorized, errors.Unauthorised)
		}

		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, errors.Unauthorised)
		}

		isAdminPath := !slices.Contains(nonAdminPaths, c.Request().URL.Path)
		if isAdminPath && !user.IsAdmin {
			slog.Warn("unauthorised access", slog.String("user id", user.ID))
			return echo.NewHTTPError(http.StatusUnauthorized, errors.Unauthorised)
		}

		role := getUserRole(user.IsAdmin)
		ctx = context.WithValue(ctx, RoleContextKey, role)
		ctx = context.WithValue(ctx, UserIDContextKey, user.ID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func TestAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Request().Header.Get("X-Test-User-ID")
		role := c.Request().Header.Get("X-Test-User-Role")

		ctx := c.Request().Context()
		ctx = context.WithValue(ctx, UserIDContextKey, userID)
		ctx = context.WithValue(ctx, RoleContextKey, role)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func getUserRole(isAdmin bool) config.Role {
	if isAdmin {
		return config.AdminRole
	}

	return config.UserRole
}
