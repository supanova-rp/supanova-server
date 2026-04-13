package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

type sqlcVideoSection struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	StorageKey string `json:"storage_key"`
	Position   int    `json:"position"`
}

type sqlcQuizSectionLegacy struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

type sqlcQuizSection struct {
	ID        string             `json:"id"`
	Position  int                `json:"position"`
	Questions []SqlcQuizQuestion `json:"questions"`
}

type sqlcCourseMaterial struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	StorageKey string `json:"storage_key"`
	Position   int    `json:"position"`
}

func (s *Store) GetCourse(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
	course, err := ExecQuery(ctx, func() (sqlc.GetCourseRow, error) {
		return s.Queries.GetCourse(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return courseFrom(&course)
}

func courseFrom(row *sqlc.GetCourseRow) (*domain.Course, error) {
	var sqlcVideos []sqlcVideoSection
	if row.VideoSections != nil {
		if err := json.Unmarshal(row.VideoSections, &sqlcVideos); err != nil {
			return nil, fmt.Errorf("failed to unmarshal video sections: %w", err)
		}
	}

	var sqlcQuizzes []sqlcQuizSection
	if row.QuizSections != nil {
		if err := json.Unmarshal(row.QuizSections, &sqlcQuizzes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal quiz sections: %w", err)
		}
	}

	var sqlcMaterials []sqlcCourseMaterial
	if row.Materials != nil {
		if err := json.Unmarshal(row.Materials, &sqlcMaterials); err != nil {
			return nil, fmt.Errorf("failed to unmarshal materials: %w", err)
		}
	}

	sections := make([]domain.CourseSection, 0, len(sqlcVideos)+len(sqlcQuizzes))
	for _, v := range sqlcVideos {
		id, err := uuid.Parse(v.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse video section ID: %w", err)
		}
		storageKey, err := uuid.Parse(v.StorageKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse video section storage key: %w", err)
		}
		sections = append(sections, &domain.VideoSection{
			ID:         id,
			Title:      v.Title,
			Position:   v.Position,
			StorageKey: storageKey,
			Type:       domain.SectionTypeVideo,
		})
	}
	for _, q := range sqlcQuizzes {
		id, err := uuid.Parse(q.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quiz section ID: %w", err)
		}
		questions, err := utils.MapToWithError(q.Questions, quizQuestionFrom)
		if err != nil {
			return nil, fmt.Errorf("failed to map quiz questions: %w", err)
		}
		sections = append(sections, &domain.QuizSection{
			ID:        id,
			Title:     "Quiz",
			Position:  q.Position,
			Type:      domain.SectionTypeQuiz,
			Questions: questions,
		})
	}
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].GetPosition() < sections[j].GetPosition()
	})

	materials, err := utils.MapToWithError(sqlcMaterials, func(m sqlcCourseMaterial) (domain.CourseMaterial, error) {
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return domain.CourseMaterial{}, fmt.Errorf("failed to parse material ID: %w", err)
		}
		storageKey, err := uuid.Parse(m.StorageKey)
		if err != nil {
			return domain.CourseMaterial{}, fmt.Errorf("failed to parse material storage key: %w", err)
		}
		return domain.CourseMaterial{
			ID:         id,
			Name:       m.Name,
			Position:   m.Position,
			StorageKey: storageKey,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return &domain.Course{
		ID:                utils.UUIDFrom(row.ID),
		Title:             row.Title.String,
		Description:       row.Description.String,
		CompletionTitle:   row.CompletionTitle.String,
		CompletionMessage: row.CompletionMessage.String,
		Sections:          sections,
		Materials:         materials,
	}, nil
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

// TODO: Remove once edit course dashboard reuses /courses/overview endpoint
func (s *Store) GetAllCourses(ctx context.Context) ([]*domain.AllCourseLegacy, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetAllCoursesRow, error) {
		return s.Queries.GetAllCourses(ctx)
	})
	if err != nil {
		return nil, err
	}

	return utils.MapToWithError(rows, func(row sqlc.GetAllCoursesRow) (*domain.AllCourseLegacy, error) {
		return allCourseFrom(&row)
	})
}

// TODO: Remove once edit course dashboard reuses /courses/overview endpoint
func allCourseFrom(row *sqlc.GetAllCoursesRow) (*domain.AllCourseLegacy, error) {
	var sqlcVideos []sqlcVideoSection
	if row.VideoSections != nil {
		if err := json.Unmarshal(row.VideoSections, &sqlcVideos); err != nil {
			return nil, fmt.Errorf("failed to unmarshal video sections: %w", err)
		}
	}

	var sqlcQuizzes []sqlcQuizSectionLegacy
	if row.QuizSections != nil {
		if err := json.Unmarshal(row.QuizSections, &sqlcQuizzes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal quiz sections: %w", err)
		}
	}

	var sqlcMaterials []sqlcCourseMaterial
	if row.Materials != nil {
		if err := json.Unmarshal(row.Materials, &sqlcMaterials); err != nil {
			return nil, fmt.Errorf("failed to unmarshal materials: %w", err)
		}
	}

	sections := make([]domain.CourseSection, 0, len(sqlcVideos)+len(sqlcQuizzes))
	for _, v := range sqlcVideos {
		id, err := uuid.Parse(v.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse video section ID: %w", err)
		}
		storageKey, err := uuid.Parse(v.StorageKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse video section storage key: %w", err)
		}
		sections = append(sections, &domain.VideoSection{
			ID:         id,
			Title:      v.Title,
			Position:   v.Position,
			StorageKey: storageKey,
			Type:       domain.SectionTypeVideo,
		})
	}
	for _, q := range sqlcQuizzes {
		id, err := uuid.Parse(q.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quiz section ID: %w", err)
		}
		sections = append(sections, &domain.QuizSectionLegacy{
			ID:       id,
			Position: q.Position,
			Type:     domain.SectionTypeQuiz,
		})
	}
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].GetPosition() < sections[j].GetPosition()
	})

	materials, err := utils.MapToWithError(sqlcMaterials, func(m sqlcCourseMaterial) (domain.CourseMaterial, error) {
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return domain.CourseMaterial{}, fmt.Errorf("failed to parse material ID: %w", err)
		}
		storageKey, err := uuid.Parse(m.StorageKey)
		if err != nil {
			return domain.CourseMaterial{}, fmt.Errorf("failed to parse material storage key: %w", err)
		}
		return domain.CourseMaterial{
			ID:         id,
			Name:       m.Name,
			Position:   m.Position,
			StorageKey: storageKey,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return &domain.AllCourseLegacy{
		ID:                utils.UUIDFrom(row.ID),
		Title:             row.Title.String,
		Description:       row.Description.String,
		CompletionTitle:   row.CompletionTitle.String,
		CompletionMessage: row.CompletionMessage.String,
		Sections:          sections,
		Materials:         materials,
	}, nil
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

func (s *Store) GetAssignedCourseTitles(ctx context.Context, userID string) ([]domain.CourseOverview, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetAssignedCourseTitlesRow, error) {
		return s.Queries.GetAssignedCourseTitles(ctx, utils.PGTextFrom(userID))
	})
	if err != nil {
		return nil, err
	}

	return utils.Map(rows, func(row sqlc.GetAssignedCourseTitlesRow) domain.CourseOverview {
		return domain.CourseOverview{
			ID:          utils.UUIDFrom(row.ID),
			Title:       row.Title.String,
			Description: row.Description.String,
		}
	}), nil
}

func (s *Store) GetCourseMaterials(ctx context.Context, courseID uuid.UUID) ([]domain.CourseMaterial, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetCourseMaterialsRow, error) {
		return s.Queries.GetCourseMaterials(ctx, utils.PGUUIDFromUUID(courseID))
	})
	if err != nil {
		return nil, err
	}

	return utils.Map(rows, courseMaterialFrom), nil
}

func (s *Store) EditCourse(ctx context.Context, params *domain.EditCourseParams) (*domain.Course, error) {
	courseID := utils.PGUUIDFromUUID(params.CourseID)

	err := ExecCommand(ctx, func() error {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx) //nolint:errcheck

		qtx := s.Queries.WithTx(tx)

		if err := qtx.UpdateCourse(ctx, sqlc.UpdateCourseParams{
			Title:             utils.PGTextFrom(params.Title),
			Description:       utils.PGTextFrom(params.Description),
			CompletionTitle:   utils.PGTextFrom(params.CompletionTitle),
			CompletionMessage: utils.PGTextFrom(params.CompletionMessage),
			ID:                courseID,
		}); err != nil {
			return fmt.Errorf("failed to update course: %w", err)
		}

		for _, m := range params.Materials {
			if err := qtx.UpsertCourseMaterial(ctx, sqlc.UpsertCourseMaterialParams{
				ID:         pgtype.UUID{Bytes: m.ID, Valid: true},
				Name:       m.Name,
				StorageKey: pgtype.UUID{Bytes: m.StorageKey, Valid: true},
				Position:   pgtype.Int4{Int32: int32(m.Position), Valid: true}, //nolint:gosec
				CourseID:   courseID,
			}); err != nil {
				return fmt.Errorf("failed to upsert material: %w", err)
			}
		}

		for _, v := range params.NewVideoSections {
			if err := qtx.InsertVideoSection(ctx, insertVideoSectionParamsFrom(&v, courseID)); err != nil {
				return fmt.Errorf("failed to insert video section: %w", err)
			}
		}

		for _, v := range params.ExistingVideoSections {
			if err := qtx.UpdateVideoSection(ctx, sqlc.UpdateVideoSectionParams{
				Title:      utils.PGTextFrom(v.Title),
				StorageKey: pgtype.UUID{Bytes: v.StorageKey, Valid: true},
				Position:   pgtype.Int4{Int32: int32(v.Position), Valid: true}, //nolint:gosec
				ID:         pgtype.UUID{Bytes: v.ID, Valid: true},
			}); err != nil {
				return fmt.Errorf("failed to update video section: %w", err)
			}
		}

		for _, q := range params.QuizSections {
			sectionID, err := upsertQuizSection(ctx, qtx, q, courseID)
			if err != nil {
				return err
			}

			for _, question := range q.Questions {
				if err := qtx.UpsertQuizQuestion(ctx, sqlc.UpsertQuizQuestionParams{
					ID:            pgtype.UUID{Bytes: question.ID, Valid: true},
					Question:      utils.PGTextFrom(question.Question),
					Position:      pgtype.Int4{Int32: int32(question.Position), Valid: true}, //nolint:gosec
					IsMultiAnswer: question.IsMultiAnswer,
					QuizSectionID: sectionID,
				}); err != nil {
					return fmt.Errorf("failed to upsert quiz question: %w", err)
				}

				for _, answer := range question.Answers {
					if err := qtx.UpsertQuizAnswer(ctx, sqlc.UpsertQuizAnswerParams{
						ID:             pgtype.UUID{Bytes: answer.ID, Valid: true},
						Answer:         utils.PGTextFrom(answer.Answer),
						CorrectAnswer:  pgtype.Bool{Bool: answer.IsCorrectAnswer, Valid: true},
						Position:       pgtype.Int4{Int32: int32(answer.Position), Valid: true}, //nolint:gosec
						QuizQuestionID: pgtype.UUID{Bytes: question.ID, Valid: true},
					}); err != nil {
						return fmt.Errorf("failed to upsert quiz answer: %w", err)
					}
				}
			}
		}

		if err := deleteCourseItems(ctx, qtx, params.DeletedSectionIDs, params.DeletedMaterialIDs); err != nil {
			return err
		}

		return tx.Commit(ctx)
	})
	if err != nil {
		return nil, err
	}

	return s.GetCourse(ctx, courseID)
}

func upsertQuizSection(ctx context.Context, qtx *sqlc.Queries, section domain.EditQuizSectionParams, courseID pgtype.UUID) (pgtype.UUID, error) {
	if section.IsNewSection {
		id, err := qtx.InsertQuizSection(ctx, insertQuizSectionParamsFrom(&domain.AddQuizSectionParams{
			Position: section.Position,
		}, courseID))
		if err != nil {
			return pgtype.UUID{}, fmt.Errorf("failed to insert quiz section: %w", err)
		}
		return id, nil
	}

	if err := qtx.UpdateQuizSectionPosition(ctx, sqlc.UpdateQuizSectionPositionParams{
		Position: pgtype.Int4{Int32: int32(section.Position), Valid: true}, //nolint:gosec
		ID:       pgtype.UUID{Bytes: section.ID, Valid: true},
	}); err != nil {
		return pgtype.UUID{}, fmt.Errorf("failed to update quiz section: %w", err)
	}

	return pgtype.UUID{Bytes: section.ID, Valid: true}, nil
}

func deleteCourseItems(ctx context.Context, qtx *sqlc.Queries, deletedSectionIDs domain.DeletedSectionIDs, deletedMaterialIDs []uuid.UUID) error {
	if len(deletedSectionIDs.AnswerIDs) > 0 {
		pgIDs := utils.Map(deletedSectionIDs.AnswerIDs, utils.PGUUIDFromUUID)
		if err := qtx.DeleteQuizAnswers(ctx, pgIDs); err != nil {
			return fmt.Errorf("failed to delete quiz answers: %w", err)
		}
	}

	if len(deletedSectionIDs.QuestionIDs) > 0 {
		pgIDs := utils.Map(deletedSectionIDs.QuestionIDs, utils.PGUUIDFromUUID)
		if err := qtx.DeleteQuizQuestions(ctx, pgIDs); err != nil {
			return fmt.Errorf("failed to delete quiz questions: %w", err)
		}
	}

	// Deleting a quiz section cascades to its questions and answers
	if len(deletedSectionIDs.QuizSectionIDs) > 0 {
		pgIDs := utils.Map(deletedSectionIDs.QuizSectionIDs, utils.PGUUIDFromUUID)
		if err := qtx.DeleteQuizSections(ctx, pgIDs); err != nil {
			return fmt.Errorf("failed to delete quiz sections: %w", err)
		}
	}

	if len(deletedSectionIDs.VideoSectionIDs) > 0 {
		pgIDs := utils.Map(deletedSectionIDs.VideoSectionIDs, utils.PGUUIDFromUUID)
		if err := qtx.DeleteVideoSections(ctx, pgIDs); err != nil {
			return fmt.Errorf("failed to delete video sections: %w", err)
		}
	}

	allDeletedSectionIDs := append(deletedSectionIDs.VideoSectionIDs, deletedSectionIDs.QuizSectionIDs...)
	if len(allDeletedSectionIDs) > 0 {
		pgIDs := utils.Map(allDeletedSectionIDs, utils.PGUUIDFromUUID)
		if err := qtx.RemoveDeletedSectionsFromProgress(ctx, pgIDs); err != nil {
			return fmt.Errorf("failed to remove deleted sections from progress: %w", err)
		}
	}

	if len(deletedMaterialIDs) > 0 {
		pgIDs := utils.Map(deletedMaterialIDs, utils.PGUUIDFromUUID)
		if err := qtx.DeleteCourseMaterials(ctx, pgIDs); err != nil {
			return fmt.Errorf("failed to delete course materials: %w", err)
		}
	}

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
