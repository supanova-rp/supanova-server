package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type GetCourseParams struct {
	ID string `param:"id" validate:"required"`
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
		return echo.NewHTTPError(http.StatusBadRequest, "validation failed")
	}

	var uuid pgtype.UUID
	err := uuid.Scan(params.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid uuid format")
	}

	course, err := h.Course.GetCourse(e.Request().Context(), uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "course not found")
		}

		slog.Error("failed to get course", slog.Any("error", err), slog.String("id", params.ID))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get course")
	}

	return e.JSON(http.StatusOK, course)
}

func (h *Handlers) AddCourse(e echo.Context) error {
	var params AddCourseParams

	if err := e.Bind(&params); err != nil {
		return err
	}

	if err := e.Validate(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "validation failed")
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
		slog.Error("failed to add course", slog.Any("error", err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add course")
	}

	return e.JSON(http.StatusCreated, map[string]any{
		"id": id,
	})
}
