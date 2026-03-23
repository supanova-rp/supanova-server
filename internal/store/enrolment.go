package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (s *Store) GetUsersAndAssignedCourses(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetUsersAndAssignedCoursesRow, error) {
		return s.Queries.GetUsersAndAssignedCourses(ctx)
	})
	if err != nil {
		return nil, err
	}

	return utils.MapToWithError(rows, func(row sqlc.GetUsersAndAssignedCoursesRow) (domain.UserWithAssignedCourses, error) {
		var courseIDs []uuid.UUID
		if row.CourseIds != nil {
			if err := json.Unmarshal(row.CourseIds, &courseIDs); err != nil {
				return domain.UserWithAssignedCourses{}, fmt.Errorf("failed to unmarshal course IDs: %w", err)
			}
		}

		return domain.UserWithAssignedCourses{
			ID:        row.ID,
			Name:      row.Name.String,
			Email:     row.Email.String,
			CourseIDs: courseIDs,
		}, nil
	})
}

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
