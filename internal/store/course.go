package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
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

func (s *Store) AddCourse(ctx context.Context, course sqlc.AddCourseParams) (*uuid.UUID, error) {
	id, err := s.Queries.AddCourse(ctx, course)
	if err != nil {
		return nil, err
	}

	ID := uuid.UUID(id.Bytes)

	return &ID, nil
}
