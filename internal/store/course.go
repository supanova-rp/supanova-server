package store

import (
	"context"

	"github.com/IBM/fp-go/array"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func (s *Store) GetCourse(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
	course, err := s.Queries.GetCourse(ctx, id)
	if err != nil {
		return nil, err
	}

	materials, err := s.Queries.GetCourseMaterials(ctx, id)
	if err != nil {
		return nil, err
	}

	videos, err := s.Queries.GetCourseVideoSections(ctx, id)
	if err != nil {
		return nil, err
	}

	// quizzes, err := s.Queries.GetCourseQuizSections(ctx, id)

	// GetQuizSections for that course

	// TODO: add type: 'quiz' field
	// TODO: add title: `Quiz ${position}`, field

	var sections []domain.CourseSection // TODO: could pre-allocate based on length of quizzes + videos
	for _, v := range videos {
		sections = append(sections, courseVideoSectionFrom(v))
	}

	// TODO: sort sections by position

	return &domain.Course{
		ID:                uuid.UUID(course.ID.Bytes),
		Title:             course.Title.String,
		Description:       course.Description.String,
		CompletionTitle:   course.CompletionTitle.String,
		CompletionMessage: course.CompletionMessage.String,
		Sections:          []domain.CourseSection{},
		Materials:         array.Map(courseMaterialFrom)(materials),
	}, nil
}

func (s *Store) AddCourse(ctx context.Context, course sqlc.AddCourseParams) (*domain.Course, error) {
	id, err := s.Queries.AddCourse(ctx, course)
	if err != nil {
		return nil, err
	}

	created, err := s.GetCourse(ctx, id)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func courseMaterialFrom(m sqlc.GetCourseMaterialsRow) domain.CourseMaterial {
	return domain.CourseMaterial{
		ID:         uuidFrom(m.ID),
		Name:       m.Name,
		Position:   int(m.Position.Int32),
		StorageKey: uuidFrom(m.StorageKey),
	}
}

func courseVideoSectionFrom(v sqlc.GetCourseVideoSectionsRow) domain.VideoSection {
	return domain.VideoSection{
		ID:         uuidFrom(v.ID),
		Title:      v.Title.String,
		Position:   int(v.Position.Int32),
		StorageKey: uuidFrom(v.StorageKey),
		Type:       domain.SectionTypeVideo,
	}
}

// TODO: where to put this?
func uuidFrom(pgUUID pgtype.UUID) uuid.UUID {
	return uuid.UUID(pgUUID.Bytes)
}
