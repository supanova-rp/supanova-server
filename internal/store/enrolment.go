package store

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (s *Store) IsEnrolled(ctx context.Context, params domain.IsEnrolledParams) (bool, error) {
	sqlcParams := sqlc.IsUserEnrolledInCourseParams{
		UserID:   utils.PGTextFrom(params.UserID),
		CourseID: utils.PGUUIDFromUUID(params.CourseID),
	}

	return ExecQuery(ctx, func() (bool, error) {
		return s.Queries.IsUserEnrolledInCourse(ctx, sqlcParams)
	})
}

func (s *Store) EnrolInCourse(ctx context.Context, params domain.EnrolInCourseParams) error {
	sqlcParams := sqlc.EnrolInCourseParams{
		UserID:   utils.PGTextFrom(params.UserID),
		CourseID: utils.PGUUIDFromUUID(params.CourseID),
	}

	return ExecCommand(ctx, func() error {
		return s.Queries.EnrolInCourse(ctx, sqlcParams)
	})
}

func (s *Store) DisenrolInCourse(ctx context.Context, params domain.DisenrolInCourseParams) error {
	sqlcParams := sqlc.DisenrolInCourseParams{
		UserID:   utils.PGTextFrom(params.UserID),
		CourseID: utils.PGUUIDFromUUID(params.CourseID),
	}

	return ExecCommand(ctx, func() error {
		return s.Queries.DisenrolInCourse(ctx, sqlcParams)
	})
}
