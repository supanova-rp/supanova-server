package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (h *Handlers) isEnrolled(ctx context.Context, courseID pgtype.UUID) (bool, error) {
	role, ok := getUserRole(ctx)
	if !ok {
		return false, errors.Wrap(errors.NotFoundInCtx("role"))
	}

	// Admins are enrolled in every course
	if role == string(config.AdminRole) {
		return true, nil
	}

	userID, ok := getUserID(ctx)
	if !ok {
		return false, errors.Wrap(errors.NotFoundInCtx("user"))
	}

	return h.Enrollment.IsEnrolled(ctx, sqlc.IsUserEnrolledInCourseParams{
		UserID:   utils.PGTextFrom(userID),
		CourseID: courseID,
	})
}
