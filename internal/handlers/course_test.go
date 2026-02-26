package handlers_test

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
)

func TestGetCourse_HappyPath(t *testing.T) {
	t.Run("returns course successfully - admin user", func(t *testing.T) {
		expected := testhelpers.Course

		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{
			Course: mockCourseRepo,
		}

		reqBody := handlers.GetCourseParams{
			ID: testhelpers.Course.ID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.GetCourse(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual domain.Course
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("course mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseCalls()), 1, testhelpers.GetCourseHandlerName)
	})

	t.Run("returns course successfully - non admin user", func(t *testing.T) {
		expected := testhelpers.Course

		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return expected, nil
			},
		}

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			IsEnrolledFunc: func(ctx context.Context, params domain.IsEnrolledParams) (bool, error) {
				return true, nil
			},
		}

		h := &handlers.Handlers{
			Course:    mockCourseRepo,
			Enrolment: mockEnrolmentRepo,
		}

		reqBody := handlers.GetCourseParams{
			ID: testhelpers.Course.ID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "course", testhelpers.WithRole(config.UserRole))

		err := h.GetCourse(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual domain.Course
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("course mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseCalls()), 1, testhelpers.GetCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.IsEnrolledCalls()), 1, testhelpers.IsEnrolledHandlerName)
	})
}

func TestGetCourse_UnhappyPath(t *testing.T) {
	courseID := testhelpers.Course.ID
	userRole := config.UserRole

	type testCase struct {
		name           string
		reqBody        handlers.GetCourseParams
		userRole       *config.Role
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "validation - missing id",
			reqBody:        handlers.GetCourseParams{},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				courseRepo := &mocks.CourseRepositoryMock{}
				h := &handlers.Handlers{Course: courseRepo}
				return h
			},
		},
		{
			name: "validation - invalid uuid",
			reqBody: handlers.GetCourseParams{
				ID: "invalid-uuid",
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				courseRepo := &mocks.CourseRepositoryMock{}
				h := &handlers.Handlers{Course: courseRepo}
				return h
			},
		},
		{
			name: "course not found",
			reqBody: handlers.GetCourseParams{
				ID: courseID.String(),
			},
			wantStatus:     http.StatusNotFound,
			expectedErrMsg: errors.NotFound("course"),
			setup: func() *handlers.Handlers {
				courseRepo := &mocks.CourseRepositoryMock{
					GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
						return nil, pgx.ErrNoRows
					},
				}
				h := &handlers.Handlers{Course: courseRepo}
				return h
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.GetCourseParams{
				ID: courseID.String(),
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("course"),
			setup: func() *handlers.Handlers {
				courseRepo := &mocks.CourseRepositoryMock{
					GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
						return nil, stdErrors.New("db error")
					},
				}
				h := &handlers.Handlers{Course: courseRepo}
				return h
			},
		},
		{
			name: "forbidden - user not enrolled",
			reqBody: handlers.GetCourseParams{
				ID: courseID.String(),
			},
			userRole:       &userRole,
			wantStatus:     http.StatusForbidden,
			expectedErrMsg: errors.Forbidden("course"),
			setup: func() *handlers.Handlers {
				courseRepo := &mocks.CourseRepositoryMock{
					GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
						return testhelpers.Course, nil
					},
				}
				enrolRepo := &mocks.EnrolmentRepositoryMock{
					IsEnrolledFunc: func(ctx context.Context, params domain.IsEnrolledParams) (bool, error) {
						return false, nil
					},
				}
				h := &handlers.Handlers{
					Course:    courseRepo,
					Enrolment: enrolRepo,
				}
				return h
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()

			var opts []testhelpers.EchoTestOption
			if tt.userRole != nil {
				opts = append(opts, testhelpers.WithRole(*tt.userRole))
			}

			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "course", opts...)

			err := h.GetCourse(ctx)

			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}

func TestAddCourse_HappyPath(t *testing.T) {
	t.Run("adds course successfully", func(t *testing.T) {
		expected := testhelpers.Course

		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params *domain.AddCourseParams) (*domain.Course, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		reqBody := handlers.AddCourseParams{
			Title:             "New Course",
			Description:       "New Description",
			CompletionTitle:   "Completion Title",
			CompletionMessage: "Completion Message",
			Sections: []handlers.AddSectionParams{
				{Video: &handlers.AddVideoSectionParams{
					Title:      "Intro Video",
					StorageKey: uuid.New().String(),
					Position:   0,
					Type:       domain.SectionTypeVideo,
				}},
				{Quiz: &handlers.AddQuizSectionParams{
					Position: 0,
					Type:     domain.SectionTypeQuiz,
					Questions: []handlers.AddQuizQuestionParams{
						{
							Question: "What is 2+2?",
							Position: 0,
							Answers: []handlers.AddQuizAnswerParams{
								{Answer: "4", IsCorrectAnswer: true, Position: 0},
							},
						},
					},
				}},
			},
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.AddCourse(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 1, testhelpers.AddCourseHandlerName)
	})
}

func TestAddCourse_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		reqBody        any
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "validation error - missing title",
			reqBody:        handlers.AddCourseParams{Description: testhelpers.Course.Description},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name:           "validation error - missing description",
			reqBody:        handlers.AddCourseParams{Title: testhelpers.Course.Title},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.AddCourseParams{
				Title:             testhelpers.Course.Title,
				Description:       testhelpers.Course.Description,
				CompletionTitle:   testhelpers.Course.CompletionTitle,
				CompletionMessage: testhelpers.Course.CompletionMessage,
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Creating("course"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Course: &mocks.CourseRepositoryMock{
						AddCourseFunc: func(ctx context.Context, params *domain.AddCourseParams) (*domain.Course, error) {
							return nil, stdErrors.New("database connection failed")
						},
					},
				}
			},
		},
		{
			name: "validation error - video section missing title",
			reqBody: handlers.AddCourseParams{
				Title:             "New Course",
				Description:       "New Description",
				CompletionTitle:   "Completion Title",
				CompletionMessage: "Completion Message",
				Sections: []handlers.AddSectionParams{
					{Video: &handlers.AddVideoSectionParams{
						StorageKey: uuid.New().String(),
						Position:   0,
						Type:       domain.SectionTypeVideo,
					}},
				},
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "validation error - quiz section no questions",
			reqBody: handlers.AddCourseParams{
				Title:             "New Course",
				Description:       "New Description",
				CompletionTitle:   "Completion Title",
				CompletionMessage: "Completion Message",
				Sections: []handlers.AddSectionParams{
					{Quiz: &handlers.AddQuizSectionParams{
						Position:  0,
						Questions: []handlers.AddQuizQuestionParams{},
						Type:      domain.SectionTypeQuiz,
					}},
				},
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "unknown section type",
			reqBody: struct {
				Title             string           `json:"title"`
				Description       string           `json:"description"`
				CompletionTitle   string           `json:"completionTitle"`
				CompletionMessage string           `json:"completionMessage"`
				Sections          []map[string]any `json:"sections"`
			}{
				Title:             "New Course",
				Description:       "New Description",
				CompletionTitle:   "Completion Title",
				CompletionMessage: "Completion Message",
				Sections:          []map[string]any{{"type": "unknown"}},
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidRequestBody,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "course")
			err := h.AddCourse(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}

func TestGetCoursesOverview_HappyPath(t *testing.T) {
	t.Run("returns course overviews successfully", func(t *testing.T) {
		expected := []domain.CourseOverview{
			{
				ID:          testhelpers.Course.ID,
				Title:       testhelpers.Course.Title,
				Description: testhelpers.Course.Description,
			},
			{
				ID:          uuid.New(),
				Title:       "Course 2",
				Description: "Description 2",
			},
		}

		mockRepo := &mocks.CourseRepositoryMock{
			GetCoursesOverviewFunc: func(ctx context.Context) ([]domain.CourseOverview, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, struct{}{}, "course-titles")

		err := h.GetCoursesOverview(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []domain.CourseOverview
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("course overviews mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetCoursesOverviewCalls()), 1, testhelpers.GetCoursesOverviewHandlerName)
	})
}

func TestGetCoursesOverview_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "internal server error",
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("course overview"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Course: &mocks.CourseRepositoryMock{
						GetCoursesOverviewFunc: func(ctx context.Context) ([]domain.CourseOverview, error) {
							return nil, stdErrors.New("database connection failed")
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			ctx, _ := testhelpers.SetupEchoContext(t, struct{}{}, "course-titles")
			err := h.GetCoursesOverview(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}
