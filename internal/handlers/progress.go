package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

const progressResource = "user progress"

type GetProgressParams struct {
	CourseID string `json:"courseId" validate:"required"`
}

func (h *Handlers) GetProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params GetProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := utils.PGUUIDFrom(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sqlcParams := sqlc.GetProgressParams{
		UserID:   userID,
		CourseID: courseID,
	}

	progress, err := h.Progress.GetProgress(e.Request().Context(), sqlcParams)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(progressResource))
		}

		return internalError(ctx, errors.Getting(progressResource), err, slog.String("id", params.CourseID))
	}

	return e.JSON(http.StatusOK, progress)
}
