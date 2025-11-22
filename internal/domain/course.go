package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CourseRepository interface {
	GetCourse(context.Context, pgtype.UUID) (*Course, error)
}

type Course struct {
	ID          uuid.UUID
	Title       string
	Description string
}
