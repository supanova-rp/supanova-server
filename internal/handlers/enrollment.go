package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (h *Handlers) isEnrolled(ctx context.Context, courseID pgtype.UUID) (bool, error) {
	// TODO: if isAdmin => return true (admins are enrolled in every course)

	userID, ok := getUserID(ctx)
	if !ok {
		return false, errors.Wrap(errors.UserIDCtxNotFound)
	}

	return h.Enrollment.IsEnrolled(ctx, sqlc.IsUserEnrolledInCourseParams{
		UserID:   utils.PGTextFrom(userID),
		CourseID: courseID,
	})
}
