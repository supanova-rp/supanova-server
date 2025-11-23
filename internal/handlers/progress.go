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
	var params GetProgressParams
	userID, ok := e.Request().Context().Value("userID").(string)
	if !ok || userID == "" {
		slog.Error(errors.ErrUserIDCtxNotFound)
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Getting(progressResource))
	}
	params.UserID = userID
	if err := e.Bind(&params); err != nil {
		return err
	}

	if err := e.Validate(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.ErrValidation)
	}

	courseUuid, err := toPGUUID(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.ErrInvalidUuid)
	}

	progress, err := h.Progress.GetProgress(e.Request().Context(), sqlc.GetProgressByIdParams{UserID: params.UserID, CourseID: courseUuid})
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(progressResource))
		}

		slog.Error(errors.Getting(progressResource), slog.Any("error", err), slog.String("id", params.CourseID))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Getting(progressResource))
	}

	return e.JSON(http.StatusOK, progress)
}
