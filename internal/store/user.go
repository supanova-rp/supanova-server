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

func (s *Store) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := ExecQuery(ctx, func() (sqlc.User, error) {
		return s.Queries.GetUser(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:    user.ID,
		Name:  user.Name.String,
		Email: user.Email.String,
	}, nil
}

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
