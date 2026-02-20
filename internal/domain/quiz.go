package domain

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

//go:generate moq -out ../handlers/mocks/quiz_mock.go -pkg mocks . QuizRepository

type QuizRepository interface {
	SaveQuizAttempt(context.Context, sqlc.SaveQuizAttemptParams) error
	UpsertQuizState(context.Context, sqlc.UpsertQuizStateParams) error
	GetQuizAttemptsByUserID(context.Context, string) ([]*QuizAttemptHistory, error)
}

type QuizAttemptHistory struct {
	QuizID        uuid.UUID      `json:"quizID"`
	TotalAttempts int32          `json:"total"`
	Attempts      []*QuizAttempt `json:"attempts"`
}

type QuizAttempt struct {
	AttemptData   json.RawMessage `json:"attemptData"`
	AttemptNumber int32           `json:"attemptNumber"`
}
