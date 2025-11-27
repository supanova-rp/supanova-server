package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

//go:generate moq -out ../handlers/mocks/course_mock.go -pkg mocks . CourseRepository

type CourseRepository interface {
	GetCourse(context.Context, pgtype.UUID) (*Course, error)
	AddCourse(context.Context, sqlc.AddCourseParams) (*Course, error)
}

type Course struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}
