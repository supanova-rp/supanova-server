package domain

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

//go:generate moq -out ../handlers/mocks/quiz_mock.go -pkg mocks . QuizRepository

type QuizRepository interface {
	SaveQuizAttempt(context.Context, SaveQuizAttemptParams) error
	UpsertQuizState(context.Context, UpsertQuizStateParams) error
	SetQuizState(context.Context, SetQuizStateParams) error
	GetQuizAttemptsByUserID(context.Context, string) ([]*QuizAttempts, error)
	GetAllQuizSections(context.Context) ([]*QuizSection, error)
	ResetQuizProgress(ctx context.Context, userID string, quizID uuid.UUID) error
	GetQuizState(ctx context.Context, userID string, quizID uuid.UUID) (*QuizState, error)
	GetQuizQuestions(ctx context.Context, sectionIDs []uuid.UUID) ([]*QuizQuestionResult, error)
}

type SaveQuizAttemptParams struct {
	UserID  string
	QuizID  uuid.UUID
	Answers []byte
}

type UpsertQuizStateParams struct {
	UserID      string
	QuizID      uuid.UUID
	QuizAnswers []byte
}

type SetQuizStateParams struct {
	UserID    string
	QuizID    uuid.UUID
	QuizState []byte
}

type QuizAttempts struct {
	QuizID         uuid.UUID        `json:"quizID"`
	TotalAttempts  int32            `json:"total"`
	Attempts       []*QuizAttempt   `json:"attempts"`
	CurrentAnswers *json.RawMessage `json:"currentAnswers"`
}

type QuizAttempt struct {
	Answers       json.RawMessage `json:"answers"`
	AttemptNumber int32           `json:"attemptNumber"`
}

type QuizState struct {
	QuizID   uuid.UUID `json:"quizId"`
	State    [][]int   `json:"quizState"`
	Attempts int32     `json:"attempts"`
}

type QuizQuestionResult struct {
	ID            uuid.UUID    `json:"id"`
	Question      string       `json:"question"`
	Position      int          `json:"position"`
	IsMultiAnswer bool         `json:"isMultiAnswer"`
	QuizSectionID uuid.UUID    `json:"quizSectionId"`
	Answers       []QuizAnswer `json:"answers"`
}
