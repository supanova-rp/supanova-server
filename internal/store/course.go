package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
)

func (s *Store) GetCourse(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
	course, err := s.Queries.GetCourseById(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.Course{
		ID:          uuid.UUID(course.ID.Bytes),
		Title:       course.Title.String,
		Description: course.Description.String,
	}, nil
}
