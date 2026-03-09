package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/middleware"
)

func getUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok || id == "" {
		slog.ErrorContext(ctx, errors.NotFoundInCtx("user"))
		return "", false
	}

	return id, true
}

func getUserRole(ctx context.Context) (config.Role, bool) {
	role, ok := ctx.Value(middleware.RoleContextKey).(config.Role)
	if !ok || role == "" {
		slog.ErrorContext(ctx, errors.NotFoundInCtx("role"))
		return "", false
	}

	return role, true
}

func bindAndValidate(c echo.Context, params any) error {
	if err := c.Bind(params); err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidRequestBody, err)
	}

	if err := c.Validate(params); err != nil {
		return httpError(http.StatusBadRequest, errors.Validation, err)
	}

	return nil
}

func httpError(code int, message string, err error) *echo.HTTPError {
	// Call SetInternal to ensure the internal error doesn't get lost so it can be logged
	return echo.NewHTTPError(code, message).SetInternal(err)
}
