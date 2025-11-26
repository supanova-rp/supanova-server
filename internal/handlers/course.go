package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

const courseResource = "course"

type GetCourseParams struct {
	ID string `json:"id" validate:"required"`
}

func (h *Handlers) GetCourse(e echo.Context) error {
	ctx := e.Request().Context()

	var params GetCourseParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := pgUUID(params.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	enrolled, err := h.isEnrolled(ctx, courseID)
	if err != nil {
		return internalError(ctx, errors.Getting(courseResource), err, slog.String("id", params.ID))
	}
	if !enrolled {
		return echo.NewHTTPError(http.StatusForbidden, errors.Forbidden(courseResource))
	}

	course, err := h.Course.GetCourse(ctx, courseID)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(courseResource))
		}

		return internalError(ctx, errors.Getting(courseResource), err, slog.String("id", params.ID))
	}

	return e.JSON(http.StatusOK, course)
}

type AddCourseParams struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (h *Handlers) AddCourse(e echo.Context) error {
	ctx := e.Request().Context()

	var params AddCourseParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	sqlcParams := sqlc.AddCourseParams{
		Title:       pgText(params.Title),
		Description: pgText(params.Description),
	}

	course, err := h.Course.AddCourse(ctx, sqlcParams)
	if err != nil {
		return internalError(ctx, errors.Creating(courseResource), err)
	}

	return e.JSON(http.StatusCreated, course)
}
