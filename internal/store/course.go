package store

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
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

	quizzes, err := s.GetQuizSections(ctx, id)
	if err != nil {
		return nil, err
	}

	sections := make([]domain.CourseSection, 0, len(videos)+len(quizzes))
	for _, v := range videos {
		sections = append(sections, courseVideoSectionFrom(&v))
	}
	for _, q := range quizzes {
		sections = append(sections, q)
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].GetPosition() < sections[j].GetPosition()
	})

	return &domain.Course{
		ID:                uuid.UUID(course.ID.Bytes),
		Title:             course.Title.String,
		Description:       course.Description.String,
		CompletionTitle:   course.CompletionTitle.String,
		CompletionMessage: course.CompletionMessage.String,
		Sections:          sections,
		Materials:         utils.Map(materials, courseMaterialFrom),
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
		ID:         utils.UUIDFrom(m.ID),
		Name:       m.Name,
		Position:   int(m.Position.Int32),
		StorageKey: utils.UUIDFrom(m.StorageKey),
	}
}

func courseVideoSectionFrom(v *sqlc.GetCourseVideoSectionsRow) *domain.VideoSection {
	return &domain.VideoSection{
		ID:         utils.UUIDFrom(v.ID),
		Title:      v.Title.String,
		Position:   int(v.Position.Int32),
		StorageKey: utils.UUIDFrom(v.StorageKey),
		Type:       domain.SectionTypeVideo,
	}
}
