package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

// The query GetCourseQuizSections returns a json byte array for quiz questions and answers
// which then has to be unmarshalled into this struct
type SqlcQuizQuestion struct {
	ID            string           `json:"id"`
	Question      string           `json:"question"`
	Position      int              `json:"position"`
	IsMultiAnswer bool             `json:"is_multi_answer"`
	Answers       []SqlcQuizAnswer `json:"answers"`
}

type SqlcQuizAnswer struct {
	ID            string `json:"id"`
	Answer        string `json:"answer"`
	CorrectAnswer bool   `json:"correct_answer"`
	Position      int    `json:"position"`
}

func (s *Store) GetQuizSections(ctx context.Context, courseID pgtype.UUID) ([]*domain.QuizSection, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetCourseQuizSectionsRow, error) {
		return s.Queries.GetCourseQuizSections(ctx, courseID)
	})
	if err != nil {
		return nil, err
	}

	sections, err := utils.MapToWithError(rows, quizSectionFrom)
	if err != nil {
		return nil, err
	}

	return sections, nil
}

func quizSectionFrom(q sqlc.GetCourseQuizSectionsRow) (*domain.QuizSection, error) {
	var sqlcQuestions []SqlcQuizQuestion
	err := json.Unmarshal(q.Questions, &sqlcQuestions)
	if err != nil {
		return nil, err
	}

	questions, err := utils.MapToWithError(sqlcQuestions, quizQuestionFrom)
	if err != nil {
		return nil, fmt.Errorf("failed to map questions: %w", err)
	}

	return &domain.QuizSection{
		ID:        utils.UUIDFrom(q.ID),
		Title:     "Quiz",
		Position:  int(q.Position.Int32),
		Type:      domain.SectionTypeQuiz,
		Questions: questions,
	}, nil
}

func quizQuestionFrom(q SqlcQuizQuestion) (domain.QuizQuestion, error) {
	id, err := uuid.Parse(q.ID)
	if err != nil {
		return domain.QuizQuestion{}, fmt.Errorf("failed to parse question ID: %w", err)
	}

	answers, err := utils.MapToWithError(q.Answers, quizAnswerFrom)
	if err != nil {
		return domain.QuizQuestion{}, fmt.Errorf("failed to map answers: %w", err)
	}

	return domain.QuizQuestion{
		ID:            id,
		Question:      q.Question,
		Position:      q.Position,
		IsMultiAnswer: q.IsMultiAnswer,
		Answers:       answers,
	}, nil
}

func quizAnswerFrom(q SqlcQuizAnswer) (domain.QuizAnswer, error) {
	id, err := uuid.Parse(q.ID)
	if err != nil {
		return domain.QuizAnswer{}, fmt.Errorf("failed to parse answer ID: %w", err)
	}

	return domain.QuizAnswer{
		ID:              id,
		Answer:          q.Answer,
		Position:        q.Position,
		IsCorrectAnswer: q.CorrectAnswer,
	}, nil
}

func (s *Store) SaveQuizAttempt(ctx context.Context, params domain.SaveQuizAttemptParams) error {
	quizID := utils.PGUUIDFromUUID(params.QuizID)
	sqlcParams := sqlc.SaveQuizAttemptParams{
		UserID:  params.UserID,
		QuizID:  quizID,
		Answers: params.Answers,
	}

	return ExecCommand(ctx, func() error {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx) //nolint:errcheck

		qtx := s.Queries.WithTx(tx)

		if err := qtx.SaveQuizAttempt(ctx, sqlcParams); err != nil {
			return fmt.Errorf("failed to save quiz attempt: %w", err)
		}

		if err := qtx.IncrementAttempts(ctx, sqlc.IncrementAttemptsParams{
			UserID: params.UserID,
			QuizID: quizID,
		}); err != nil {
			return fmt.Errorf("failed to increment attempts: %w", err)
		}

		return tx.Commit(ctx)
	})
}

func (s *Store) GetQuizAttemptsByUserID(ctx context.Context, userID string) ([]*domain.QuizAttempts, error) {
	attemptRows, err := ExecQuery(ctx, func() ([]sqlc.GetQuizAttemptsByUserIDRow, error) {
		return s.Queries.GetQuizAttemptsByUserID(ctx, userID)
	})
	if err != nil {
		return nil, err
	}

	currentAnswersByQuizID, err := s.GetCurrentQuizAnswersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	attemptsByQuizID := map[uuid.UUID]*domain.QuizAttempts{}

	for _, row := range attemptRows {
		quizID := utils.UUIDFrom(row.QuizID)
		if _, ok := attemptsByQuizID[quizID]; !ok {
			attemptsByQuizID[quizID] = &domain.QuizAttempts{
				QuizID:        quizID,
				TotalAttempts: row.TotalAttempts,
			}
		}

		// If user has a current attempt/state for this quiz, add it to the attempt data
		currentAnswers, ok := currentAnswersByQuizID[quizID]
		if ok {
			attemptsByQuizID[quizID].CurrentAnswers = &currentAnswers
		}

		if row.ID.Valid {
			attemptsByQuizID[quizID].Attempts = append(attemptsByQuizID[quizID].Attempts, quizAttemptFrom(&row))
		}
	}

	result := make([]*domain.QuizAttempts, 0, len(attemptsByQuizID))
	for _, quizAttempt := range attemptsByQuizID {
		result = append(result, quizAttempt)
	}

	return result, nil
}

func (s *Store) GetAllQuizSections(ctx context.Context) ([]*domain.QuizSection, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetAllQuizSectionsRow, error) {
		return s.Queries.GetAllQuizSections(ctx)
	})
	if err != nil {
		return nil, err
	}

	sections, err := utils.MapToWithError(rows, func(row sqlc.GetAllQuizSectionsRow) (*domain.QuizSection, error) {
		return quizSectionFrom(sqlc.GetCourseQuizSectionsRow(row))
	})
	if err != nil {
		return nil, err
	}

	return sections, nil
}

func (s *Store) ResetQuizProgress(ctx context.Context, userID string, quizID uuid.UUID) error {
	pgQuizID := utils.PGUUIDFromUUID(quizID)
	return ExecCommand(ctx, func() error {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx) //nolint:errcheck

		qtx := s.Queries.WithTx(tx)

		if err := qtx.DeleteUserQuizState(ctx, sqlc.DeleteUserQuizStateParams{
			UserID: userID,
			QuizID: pgQuizID,
		}); err != nil {
			return fmt.Errorf("failed to delete user quiz state: %w", err)
		}

		if err := qtx.DeleteQuizAttempts(ctx, sqlc.DeleteQuizAttemptsParams{
			UserID: userID,
			QuizID: pgQuizID,
		}); err != nil {
			return fmt.Errorf("failed to delete quiz attempts: %w", err)
		}

		return tx.Commit(ctx)
	})
}

func (s *Store) UpsertQuizState(ctx context.Context, params domain.UpsertQuizStateParams) error {
	sqlcParams := sqlc.UpsertQuizStateParams{
		UserID:      params.UserID,
		QuizID:      utils.PGUUIDFromUUID(params.QuizID),
		QuizAnswers: params.QuizAnswers,
	}

	return ExecCommand(ctx, func() error {
		return s.Queries.UpsertQuizState(ctx, sqlcParams)
	})
}

func quizAttemptFrom(row *sqlc.GetQuizAttemptsByUserIDRow) *domain.QuizAttempt {
	return &domain.QuizAttempt{
		Answers:       row.Answers,
		AttemptNumber: row.AttemptNumber.Int32,
	}
}

func (s *Store) GetCurrentQuizAnswersByUserID(ctx context.Context, userID string) (map[uuid.UUID]json.RawMessage, error) {
	completedSectionIDs, err := s.getCompletedSectionIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	rows, err := ExecQuery(ctx, func() ([]sqlc.GetCurrentQuizAnswersByUserIDRow, error) {
		return s.Queries.GetCurrentQuizAnswersByUserID(ctx, userID)
	})
	if err != nil {
		return nil, err
	}

	completedSet := make(map[uuid.UUID]struct{}, len(completedSectionIDs))
	for _, id := range completedSectionIDs {
		completedSet[id] = struct{}{}
	}

	result := map[uuid.UUID]json.RawMessage{}
	for _, row := range rows {
		quizID := uuid.UUID(row.QuizID.Bytes)
		if _, ok := completedSet[quizID]; ok {
			continue
		}
		result[quizID] = row.QuizAnswers
	}

	return result, nil
}

func (s *Store) getCompletedSectionIDs(ctx context.Context, userID string) ([]uuid.UUID, error) {
	rows, err := ExecQuery(ctx, func() ([][]pgtype.UUID, error) {
		return s.Queries.GetCompletedSectionIDsByUserID(ctx, userID)
	})
	if err != nil {
		return nil, err
	}

	var result []uuid.UUID
	for _, sectionIDs := range rows {
		for _, id := range sectionIDs {
			result = append(result, uuid.UUID(id.Bytes))
		}
	}

	return result, nil
}
