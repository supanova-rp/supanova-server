package store

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func (s *Store) IsEnrolled(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error) {
	return s.Queries.IsUserEnrolledInCourse(ctx, params)
}

func (s *Store) EnrolInCourse(ctx context.Context, params sqlc.EnrolInCourseParams) error {
	return s.Queries.EnrolInCourse(ctx, params)
}

func (s *Store) DisenrolInCourse(ctx context.Context, params sqlc.DisenrolInCourseParams) error {
	return s.Queries.DisenrolInCourse(ctx, params)
}
