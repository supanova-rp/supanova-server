package domain

import (
	"context"

	"github.com/google/uuid"
)

type SystemRepository interface {
	pingDB(context.Context) error
}

type CourseRepository interface {
	getCourse(context.Context, uuid.UUID) (Course, error)
}
