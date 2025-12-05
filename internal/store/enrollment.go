package store

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func (s *Store) IsEnrolled(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error) {
	return s.Queries.IsUserEnrolledInCourse(ctx, params)
}

func (s *Store) EnrollUserInCourse(ctx context.Context, params sqlc.EnrollUserInCourseParams) error {
	return s.Queries.EnrollUserInCourse(ctx, params)
}

func (s *Store) DisenrollUserInCourse(ctx context.Context, params sqlc.DisenrollUserInCourseParams) error {
	return s.Queries.DisenrollUserInCourse(ctx, params)
}
