package handlers

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

const courseResource = "course"

type GetCourseParams struct {
	ID string `json:"id" validate:"required"`
}

type AddCourseParams struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (h *Handlers) GetCourse(e echo.Context) error {
	var params GetCourseParams
	if err := e.Bind(&params); err != nil {
		return err
	}

	if err := e.Validate(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.ErrValidation)
	}

	uuid, err := toPGUUID(params.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.ErrInvalidUuid)
	}

	course, err := h.Course.GetCourse(e.Request().Context(), uuid)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(courseResource))
		}

		slog.Error(errors.Getting(courseResource), slog.Any("error", err), slog.String("id", params.ID))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Getting(courseResource))
	}

	return e.JSON(http.StatusOK, course)
}

func (h *Handlers) AddCourse(e echo.Context) error {
	var params AddCourseParams

	if err := e.Bind(&params); err != nil {
		return err
	}

	if err := e.Validate(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidFormat(courseResource))
	}

	sqlcParams := sqlc.AddCourseParams{
		Title: pgtype.Text{
			String: params.Title,
			Valid:  true,
		},
		Description: pgtype.Text{
			String: params.Description,
			Valid:  true,
		},
	}

	id, err := h.Course.AddCourse(e.Request().Context(), sqlcParams)
	if err != nil {
		slog.Error(errors.Getting(courseResource), slog.Any("error", err))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Getting(courseResource))
	}

	return e.JSON(http.StatusCreated, map[string]any{
		"id": id,
	})
}
