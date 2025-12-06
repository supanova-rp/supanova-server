package domain

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

//go:generate moq -out ../handlers/mocks/enrolment_mock.go -pkg mocks . EnrolmentRepository

type EnrolmentRepository interface {
	IsEnrolled(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error)
	EnrolInCourse(ctx context.Context, params sqlc.EnrolInCourseParams) error
	DisenrolInCourse(ctx context.Context, params sqlc.DisenrolInCourseParams) error
}
