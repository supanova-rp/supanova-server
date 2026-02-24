package domain

import (
	"context"

	"github.com/google/uuid"
)

//go:generate moq -out ../handlers/mocks/enrolment_mock.go -pkg mocks . EnrolmentRepository

type EnrolmentRepository interface {
	IsEnrolled(ctx context.Context, params IsEnrolledParams) (bool, error)
	EnrolInCourse(ctx context.Context, params EnrolInCourseParams) error
	DisenrolInCourse(ctx context.Context, params DisenrolInCourseParams) error
}

type IsEnrolledParams struct {
	UserID   string
	CourseID uuid.UUID
}

type EnrolInCourseParams struct {
	UserID   string
	CourseID uuid.UUID
}

type DisenrolInCourseParams struct {
	UserID   string
	CourseID uuid.UUID
}
