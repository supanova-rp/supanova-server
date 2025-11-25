package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type RequestParams struct {
	ID string `param:"id" validate:"required"`
}

func (h *Handlers) GetCourse(e echo.Context) error {
	var params RequestParams
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

	course, err := h.Store.Queries.GetCourseById(e.Request().Context(), uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "course not found")
		}

		slog.Error("failed to get course", slog.Any("err", err), slog.String("id", params.ID))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get course")
	}

	return e.JSON(http.StatusOK, course)
}
