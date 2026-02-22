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
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidRequestBody)
	}

	if err := c.Validate(params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Validation)
	}

	return nil
}

func internalError(ctx context.Context, message string, err error, attrs ...slog.Attr) error {
	logAttrs := []any{slog.Any("error", err)}
	for _, attr := range attrs {
		logAttrs = append(logAttrs, attr)
	}

	if userID, ok := getUserID(ctx); ok {
		logAttrs = append(logAttrs, slog.String("user_id", userID))
	}

	slog.ErrorContext(ctx, message, logAttrs...)
	return echo.NewHTTPError(http.StatusInternalServerError, message)
}
