package handlers_test

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
)

func TestGetQuizQuestions_HappyPath(t *testing.T) {
	sectionID := uuid.New()
	questionID := uuid.New()
	answerID := uuid.New()

	t.Run("returns questions with answers successfully", func(t *testing.T) {
		expected := []*domain.QuizQuestionLegacy{
			{
				ID:            questionID,
				Question:      "What is 2+2?",
				Position:      0,
				IsMultiAnswer: false,
				QuizSectionID: sectionID,
				Answers: []domain.QuizAnswer{
					{
						ID:              answerID,
						Answer:          "4",
						Position:        0,
						IsCorrectAnswer: true,
					},
				},
			},
		}

		mockRepo := &mocks.QuizRepositoryMock{
			GetQuizQuestionsFunc: func(ctx context.Context, sectionIDs []uuid.UUID) ([]*domain.QuizQuestionLegacy, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Quiz: mockRepo}

		reqBody := handlers.GetQuizQuestionsParams{
			QuizSectionIDs: []string{sectionID.String()},
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "quiz-questions")

		err := h.GetQuizQuestions(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []*domain.QuizQuestionLegacy
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("quiz questions mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetQuizQuestionsCalls()), 1, testhelpers.GetQuizQuestionsHandlerName)
	})

	t.Run("returns empty slice when quizSectionIds is empty", func(t *testing.T) {
		mockRepo := &mocks.QuizRepositoryMock{}

		h := &handlers.Handlers{Quiz: mockRepo}

		reqBody := handlers.GetQuizQuestionsParams{
			QuizSectionIDs: []string{},
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "quiz-questions")

		err := h.GetQuizQuestions(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []*domain.QuizQuestionLegacy
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(actual) != 0 {
			t.Errorf("expected empty slice, got %v", actual)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetQuizQuestionsCalls()), 0, testhelpers.GetQuizQuestionsHandlerName)
	})
}

func TestGetQuizQuestions_UnhappyPath(t *testing.T) {
	sectionID := uuid.New()

	type testCase struct {
		name           string
		reqBody        handlers.GetQuizQuestionsParams
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "validation - missing quizSectionIds",
			reqBody:        handlers.GetQuizQuestionsParams{},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Quiz: &mocks.QuizRepositoryMock{}}
			},
		},
		{
			name: "validation - invalid uuid",
			reqBody: handlers.GetQuizQuestionsParams{
				QuizSectionIDs: []string{"invalid-uuid"},
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Quiz: &mocks.QuizRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.GetQuizQuestionsParams{
				QuizSectionIDs: []string{sectionID.String()},
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("quiz questions"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Quiz: &mocks.QuizRepositoryMock{
						GetQuizQuestionsFunc: func(ctx context.Context, sectionIDs []uuid.UUID) ([]*domain.QuizQuestionLegacy, error) {
							return nil, stdErrors.New("db error")
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "quiz-questions")
			err := h.GetQuizQuestions(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}
