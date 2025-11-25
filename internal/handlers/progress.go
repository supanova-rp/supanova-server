package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

const progressResource = "user progress"

type GetProgressParams struct {
	CourseID string `json:"courseId" validate:"required"`
	UserID   string `validate:"required"` // comes from context
}

func (h *Handlers) GetProgress(e echo.Context) error {
	ctx := e.Request().Context()

	id, ok := userID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Getting(progressResource))
	}

	var params GetProgressParams
	params.UserID = id
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseUUID, err := pgUUID(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sqlcParams := sqlc.GetProgressByIDParams{
		UserID:   params.UserID,
		CourseID: courseUUID,
	}

	progress, err := h.Progress.GetProgress(e.Request().Context(), sqlcParams)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(progressResource))
		}

		slog.ErrorContext(ctx, errors.Getting(progressResource), slog.Any("error", err), slog.String("id", params.CourseID))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Getting(progressResource))
	}

	return e.JSON(http.StatusOK, progress)
}
