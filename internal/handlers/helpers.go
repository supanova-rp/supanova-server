package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/middleware"
)

func getUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok || id == "" {
		slog.ErrorContext(ctx, errors.UserIDCtxNotFound)
		return "", false
	}

	return id, true
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

	slog.ErrorContext(ctx, message, logAttrs...)
	return echo.NewHTTPError(http.StatusInternalServerError, message)
}
