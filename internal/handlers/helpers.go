package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const userIDContextKey = "userID"

func pgUUID(id string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(id)
	return uuid, err
}

func pgText(text string) pgtype.Text {
	return pgtype.Text{
		String: text,
		Valid:  true,
	}
}

func userID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
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
