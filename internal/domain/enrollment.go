package domain

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

//go:generate moq -out ../handlers/mocks/enrollment_mock.go -pkg mocks . EnrollmentRepository

type EnrollmentRepository interface {
	IsEnrolled(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error)
	EnrollUserInCourse(ctx context.Context, params sqlc.EnrollUserInCourseParams) error
	DisenrollUserInCourse(ctx context.Context, params sqlc.DisenrollUserInCourseParams) error
}
