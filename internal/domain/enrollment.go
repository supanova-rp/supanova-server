package domain

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type EnrollmentRepository interface {
	IsEnrolled(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error)
}
