package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/services/auth"
)

type AuthParams struct {
	AccessToken string `json:"access_token" validate:"required"`
}

type Role string

const adminRole Role = "admin"
const userRole Role = "user"

var nonAdminPaths = []string{
	fmt.Sprintf("%s/course", config.APIVersion),
	fmt.Sprintf("%s/assigned-course-titles", config.APIVersion),
	fmt.Sprintf("%s/get-progress", config.APIVersion),
	fmt.Sprintf("%s/update-progress", config.APIVersion),
	fmt.Sprintf("%s/set-intro-completed", config.APIVersion),
	fmt.Sprintf("%s/set-course-completed", config.APIVersion),
	fmt.Sprintf("%s/video-url", config.APIVersion),
	fmt.Sprintf("%s/materials", config.APIVersion),
	fmt.Sprintf("%s/get-quiz-state", config.APIVersion),
	fmt.Sprintf("%s/set-quiz-state", config.APIVersion),
	fmt.Sprintf("%s/increment-attempts", config.APIVersion),
}

func AuthMiddleware(next echo.HandlerFunc, authProvider *auth.AuthProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var params AuthParams

		if err := handlers.BindAndValidate(c, params); err != nil {
			return err
		}

		user, err := authProvider.GetUserFromIDToken(ctx, params.AccessToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Getting("user"))
		}

		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, errors.Unauthorised)
		}

		isAdminPath := !slices.Contains(nonAdminPaths, c.Request().URL.Path)
		if isAdminPath && !user.IsAdmin {
			return echo.NewHTTPError(http.StatusUnauthorized, errors.Unauthorised)
		}

		role := getRole(user.IsAdmin)
		ctx = context.WithValue(ctx, config.RoleContextKey, role)
		ctx = context.WithValue(ctx, config.UserIDContextKey, user.ID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func TestAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Request().Header.Get("X-Test-User-ID")

		ctx := context.WithValue(c.Request().Context(), config.UserIDContextKey, userID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

func getRole(isAdmin bool) Role {
	if isAdmin {
		return adminRole
	}

	return userRole
}
