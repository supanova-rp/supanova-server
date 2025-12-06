package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

const enrollmentResource = "enrollment"

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

	courseID, err := utils.PGUUIDFrom(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	if params.IsEnrolled {
		err = h.Enrolment.DisenrolInCourse(ctx, sqlc.DisenrolInCourseParams{
			UserID:   utils.PGTextFrom(userID),
			CourseID: courseID,
		})
		if err != nil {
			return internalError(ctx, errors.Deleting(enrollmentResource), err, slog.String("course_id", params.CourseID))
		}
	} else {
		err = h.Enrolment.EnrolInCourse(ctx, sqlc.EnrolInCourseParams{
			UserID:   utils.PGTextFrom(userID),
			CourseID: courseID,
		})
		if err != nil {
			return internalError(ctx, errors.Creating(enrollmentResource), err, slog.String("course_id", params.CourseID))
		}
	}

	return e.NoContent(http.StatusOK)
}

func (h *Handlers) isEnrolled(ctx context.Context, courseID pgtype.UUID) (bool, error) {
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

	return h.Enrolment.IsEnrolled(ctx, sqlc.IsUserEnrolledInCourseParams{
		UserID:   utils.PGTextFrom(userID),
		CourseID: courseID,
	})
}
