package store

import (
	"context"
	"encoding/json"
	"fmt"

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

type sqlcAssignedCourseTitle struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (s *Store) GetUsersAndAssignedCourses(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetUsersAndAssignedCoursesRow, error) {
		return s.Queries.GetUsersAndAssignedCourses(ctx)
	})
	if err != nil {
		return nil, err
	}

	return utils.MapToWithError(rows, func(row sqlc.GetUsersAndAssignedCoursesRow) (domain.UserWithAssignedCourses, error) {
		var sqlcCourses []sqlcAssignedCourseTitle
		if row.Courses != nil {
			if err := json.Unmarshal(row.Courses, &sqlcCourses); err != nil {
				return domain.UserWithAssignedCourses{}, fmt.Errorf("failed to unmarshal courses: %w", err)
			}
		}

		courses := utils.Map(sqlcCourses, func(c sqlcAssignedCourseTitle) domain.AssignedCourseTitle {
			return domain.AssignedCourseTitle{
				ID:    c.ID,
				Title: c.Title,
			}
		})

		return domain.UserWithAssignedCourses{
			ID:      row.ID,
			Name:    row.Name.String,
			Email:   row.Email.String,
			Courses: courses,
		}, nil
	})
}
