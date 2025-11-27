package store

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func (s *Store) IsEnrolled(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error) {
	return s.Queries.IsUserEnrolledInCourse(ctx, params)
}
