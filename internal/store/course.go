package store

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (s *Store) GetCourse(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
	course, err := ExecQuery(ctx, func() (sqlc.Course, error) {
		return s.Queries.GetCourse(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	materials, err := ExecQuery(ctx, func() ([]sqlc.GetCourseMaterialsRow, error) {
		return s.Queries.GetCourseMaterials(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	sections, err := s.GetCourseSections(ctx, id)
	if err != nil {
		return nil, err
	}

	formattedCourse := &domain.Course{
		ID:                uuid.UUID(course.ID.Bytes),
		Title:             course.Title.String,
		Description:       course.Description.String,
		CompletionTitle:   course.CompletionTitle.String,
		CompletionMessage: course.CompletionMessage.String,
		Sections:          sections,
		Materials:         utils.Map(materials, courseMaterialFrom),
	}

	return formattedCourse, nil
}

func (s *Store) GetCourseSections(ctx context.Context, courseID pgtype.UUID) ([]domain.CourseSection, error) {
	videos, err := ExecQuery(ctx, func() ([]sqlc.GetCourseVideoSectionsRow, error) {
		return s.Queries.GetCourseVideoSections(ctx, courseID)
	})
	if err != nil {
		return nil, err
	}

	quizzes, err := s.GetQuizSections(ctx, courseID)
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

	return sections, nil
}

func (s *Store) AddCourse(ctx context.Context, params *domain.AddCourseParams) (*domain.Course, error) {
	var courseID pgtype.UUID

	err := ExecCommand(ctx, func() error {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx) //nolint:errcheck

		qtx := s.Queries.WithTx(tx)

		id, err := qtx.AddCourse(ctx, sqlc.AddCourseParams{
			Title:             utils.PGTextFrom(params.Title),
			Description:       utils.PGTextFrom(params.Description),
			CompletionTitle:   utils.PGTextFrom(params.CompletionTitle),
			CompletionMessage: utils.PGTextFrom(params.CompletionMessage),
		})
		if err != nil {
			return fmt.Errorf("failed to insert course: %w", err)
		}
		courseID = id

		for _, m := range params.Materials {
			if err := qtx.InsertCourseMaterial(ctx, insertCourseMaterialParamsFrom(m, id)); err != nil {
				return fmt.Errorf("failed to insert material: %w", err)
			}
		}

		for _, s := range params.Sections {
			switch sec := s.(type) {
			case *domain.AddVideoSectionParams:
				if err := qtx.InsertVideoSection(ctx, insertVideoSectionParamsFrom(sec, id)); err != nil {
					return fmt.Errorf("failed to insert video section: %w", err)
				}
			case *domain.AddQuizSectionParams:
				sectionID, err := qtx.InsertQuizSection(ctx, insertQuizSectionParamsFrom(sec, id))
				if err != nil {
					return fmt.Errorf("failed to insert quiz section: %w", err)
				}
				for _, q := range sec.Questions {
					questionID, err := qtx.InsertQuizQuestion(ctx, insertQuizQuestionParamsFrom(q, sectionID))
					if err != nil {
						return fmt.Errorf("failed to insert quiz question: %w", err)
					}
					for _, a := range q.Answers {
						if err := qtx.InsertQuizAnswer(ctx, insertQuizAnswerParamsFrom(a, questionID)); err != nil {
							return fmt.Errorf("failed to insert quiz answer: %w", err)
						}
					}
				}
			}
		}

		return tx.Commit(ctx)
	})
	if err != nil {
		return nil, err
	}

	return s.GetCourse(ctx, courseID)
}

func (s *Store) GetCoursesOverview(ctx context.Context) ([]domain.CourseOverview, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetCoursesOverviewRow, error) {
		return s.Queries.GetCoursesOverview(ctx)
	})
	if err != nil {
		return nil, err
	}

	return utils.Map(rows, courseOverviewFrom), nil
}

func (s *Store) EditCourse(ctx context.Context, id pgtype.UUID) error {
	return nil
}

func (s *Store) DeleteCourse(ctx context.Context, id uuid.UUID) error {
	return ExecCommand(ctx, func() error {
		return s.Queries.DeleteCourse(ctx, utils.PGUUIDFromUUID(id))
	})
}

func insertVideoSectionParamsFrom(sec *domain.AddVideoSectionParams, courseID pgtype.UUID) sqlc.InsertVideoSectionParams {
	return sqlc.InsertVideoSectionParams{
		Title:      utils.PGTextFrom(sec.Title),
		StorageKey: pgtype.UUID{Bytes: sec.StorageKey, Valid: true},
		Position:   pgtype.Int4{Int32: int32(sec.Position), Valid: true}, //nolint:gosec
		CourseID:   courseID,
	}
}

func insertQuizSectionParamsFrom(sec *domain.AddQuizSectionParams, courseID pgtype.UUID) sqlc.InsertQuizSectionParams {
	return sqlc.InsertQuizSectionParams{
		Position: pgtype.Int4{Int32: int32(sec.Position), Valid: true}, //nolint:gosec
		CourseID: courseID,
	}
}

func insertQuizQuestionParamsFrom(q domain.AddSectionQuestionParams, sectionID pgtype.UUID) sqlc.InsertQuizQuestionParams {
	return sqlc.InsertQuizQuestionParams{
		Question:      utils.PGTextFrom(q.Question),
		Position:      pgtype.Int4{Int32: int32(q.Position), Valid: true}, //nolint:gosec
		IsMultiAnswer: q.IsMultiAnswer,
		QuizSectionID: sectionID,
	}
}

func insertQuizAnswerParamsFrom(a domain.AddSectionQuestionAnswerParams, questionID pgtype.UUID) sqlc.InsertQuizAnswerParams {
	return sqlc.InsertQuizAnswerParams{
		Answer:         utils.PGTextFrom(a.Answer),
		CorrectAnswer:  pgtype.Bool{Bool: a.IsCorrectAnswer, Valid: true},
		Position:       pgtype.Int4{Int32: int32(a.Position), Valid: true}, //nolint:gosec
		QuizQuestionID: questionID,
	}
}

func insertCourseMaterialParamsFrom(m domain.AddMaterialParams, courseID pgtype.UUID) sqlc.InsertCourseMaterialParams {
	return sqlc.InsertCourseMaterialParams{
		ID:         pgtype.UUID{Bytes: m.ID, Valid: true},
		Name:       m.Name,
		StorageKey: pgtype.UUID{Bytes: m.StorageKey, Valid: true},
		Position:   pgtype.Int4{Int32: int32(m.Position), Valid: true}, //nolint:gosec
		CourseID:   courseID,
	}
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
