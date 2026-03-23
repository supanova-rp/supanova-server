package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const (
	usersWithAssignedCoursesResource = "users with assigned courses"
)

func (h *Handlers) GetUsersAndAssignedCourses(e echo.Context) error {
	ctx := e.Request().Context()

	users, err := h.User.GetUsersAndAssignedCourses(ctx)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(usersWithAssignedCoursesResource), err)
	}

	return e.JSON(http.StatusOK, users)
}
