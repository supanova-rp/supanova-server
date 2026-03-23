package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const (
	enrolmentResource                = "enrolment"
	usersWithAssignedCoursesResource = "users with assigned courses"
)

func (h *Handlers) GetUsersAndAssignedCourses(e echo.Context) error {
	ctx := e.Request().Context()

	usersToCourses, err := h.Enrolment.GetUsersAndAssignedCourses(ctx)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Getting(usersWithAssignedCoursesResource), err)
	}

	return e.JSON(http.StatusOK, usersToCourses)
}

type UpdateCourseEnrolmentParams struct {
	UserID     string `json:"user_id" validate:"required"`
	CourseID   string `json:"course_id" validate:"required"`
	IsEnrolled bool   `json:"isAssigned"`
}

func (h *Handlers) UpdateCourseEnrolment(e echo.Context) error {
	ctx := e.Request().Context()

	var params UpdateCourseEnrolmentParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	if params.IsEnrolled {
		err := h.Enrolment.DisenrolInCourse(ctx, domain.DisenrolInCourseParams{
			UserID:   params.UserID,
			CourseID: courseID,
		})
		if err != nil {
			return httpError(http.StatusInternalServerError, errors.Deleting(enrolmentResource), err)
		}
	} else {
		err := h.Enrolment.EnrolInCourse(ctx, domain.EnrolInCourseParams{
			UserID:   params.UserID,
			CourseID: courseID,
		})
		if err != nil {
			return httpError(http.StatusInternalServerError, errors.Creating(enrolmentResource), err)
		}
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) isEnrolled(ctx context.Context, courseID uuid.UUID) (bool, error) {
	role, ok := getUserRole(ctx)
	if !ok {
		return false, errors.Wrap(errors.NotFoundInCtx("role"))
	}

	// Admins are enrolled in every course
	if role == config.AdminRole {
		return true, nil
	}

	userID, ok := getUserID(ctx)
	if !ok {
		return false, errors.Wrap(errors.NotFoundInCtx("user"))
	}

	return h.Enrolment.IsEnrolled(ctx, domain.IsEnrolledParams{
		UserID:   userID,
		CourseID: courseID,
	})
}
