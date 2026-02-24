package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const enrolmentResource = "enrolment"

type UpdateCourseEnrolmentParams struct {
	CourseID   string `json:"course_id" validate:"required"`
	IsEnrolled bool   `json:"isAssigned"`
}

func (h *Handlers) UpdateCourseEnrolment(e echo.Context) error {
	ctx := e.Request().Context()

	var params UpdateCourseEnrolmentParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	if params.IsEnrolled {
		err := h.Enrolment.DisenrolInCourse(ctx, domain.DisenrolInCourseParams{
			UserID:   userID,
			CourseID: courseID,
		})
		if err != nil {
			return internalError(ctx, errors.Deleting(enrolmentResource), err, slog.String("course_id", params.CourseID))
		}
	} else {
		err := h.Enrolment.EnrolInCourse(ctx, domain.EnrolInCourseParams{
			UserID:   userID,
			CourseID: courseID,
		})
		if err != nil {
			return internalError(ctx, errors.Creating(enrolmentResource), err, slog.String("course_id", params.CourseID))
		}
	}

	return e.NoContent(http.StatusOK)
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
