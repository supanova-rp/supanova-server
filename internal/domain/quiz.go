package domain

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

//go:generate moq -out ../handlers/mocks/quiz_mock.go -pkg mocks . QuizRepository

type QuizRepository interface {
	SaveQuizAttempt(context.Context, sqlc.SaveQuizAttemptParams) error
	UpsertQuizState(context.Context, sqlc.UpsertQuizStateParams) error
	GetQuizAttemptsByUserID(context.Context, string) ([]*QuizAttempts, error)
	GetAllQuizSections(context.Context) ([]*QuizSection, error)
	ResetQuizProgress(ctx context.Context, userID string, quizID pgtype.UUID) error
}

type QuizAttempts struct {
	QuizID         uuid.UUID        `json:"quizID"`
	TotalAttempts  int32            `json:"total"`
	Attempts       []*QuizAttempt   `json:"attempts"`
	CurrentAttempt *json.RawMessage `json:"currentAttempt"`
}

type QuizAttempt struct {
	Answers       json.RawMessage `json:"answers"`
	AttemptNumber int32           `json:"attemptNumber"`
}
