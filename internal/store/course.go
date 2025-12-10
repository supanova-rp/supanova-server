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
	cachedCourse, ok := s.courseCache.Get(id.String())
	if ok {
		return &cachedCourse, nil
	}

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

	formattedCourse := &domain.Course{
		ID:                uuid.UUID(course.ID.Bytes),
		Title:             course.Title.String,
		Description:       course.Description.String,
		CompletionTitle:   course.CompletionTitle.String,
		CompletionMessage: course.CompletionMessage.String,
		Sections:          sections,
		Materials:         utils.Map(materials, courseMaterialFrom),
	}

	s.courseCache.Set(id.String(), *formattedCourse)

	return formattedCourse, nil
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

func (s *Store) GetCoursesOverview(ctx context.Context) ([]domain.CourseOverview, error) {
	rows, err := s.Queries.GetCoursesOverview(ctx)
	if err != nil {
		return nil, err
	}

	return utils.Map(rows, courseOverviewFrom), nil
}

func (s *Store) EditCourse(ctx context.Context, id pgtype.UUID) error {
	// TODO: update course in cache when course is edited
	return nil
}

func (s *Store) DeleteCourse(ctx context.Context, id pgtype.UUID) error {
	// TODO: remove course from cache when course is deleted
	return nil
}

func courseOverviewFrom(row sqlc.GetCoursesOverviewRow) domain.CourseOverview {
	return domain.CourseOverview{
		ID:          utils.UUIDFrom(row.ID),
		Title:       row.Title.String,
		Description: row.Description.String,
	}
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
