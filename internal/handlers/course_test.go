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

func TestGetAssignedCourseTitles_HappyPath(t *testing.T) {
	t.Run("returns assigned course titles successfully", func(t *testing.T) {
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
			GetAssignedCourseTitlesFunc: func(ctx context.Context, userID string) ([]domain.CourseOverview, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, struct{}{}, "assigned-course-titles")

		err := h.GetAssignedCourseTitles(ctx)
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

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAssignedCourseTitlesCalls()), 1, testhelpers.GetAssignedCourseTitlesHandlerName)
	})

	t.Run("returns empty slice when no courses assigned", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			GetAssignedCourseTitlesFunc: func(ctx context.Context, userID string) ([]domain.CourseOverview, error) {
				return []domain.CourseOverview{}, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, struct{}{}, "assigned-course-titles")

		err := h.GetAssignedCourseTitles(ctx)
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

		if len(actual) != 0 {
			t.Errorf("expected empty slice, got %v", actual)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAssignedCourseTitlesCalls()), 1, testhelpers.GetAssignedCourseTitlesHandlerName)
	})
}

func TestGetAssignedCourseTitles_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		opts           []testhelpers.EchoTestOption
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "user not in context",
			opts:           []testhelpers.EchoTestOption{testhelpers.WithUserID("")},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.NotFoundInCtx("user"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name:           "internal server error",
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("assigned course titles"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Course: &mocks.CourseRepositoryMock{
						GetAssignedCourseTitlesFunc: func(ctx context.Context, userID string) ([]domain.CourseOverview, error) {
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
			ctx, _ := testhelpers.SetupEchoContext(t, struct{}{}, "assigned-course-titles", tt.opts...)
			err := h.GetAssignedCourseTitles(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
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

// TODO: Remove all GetCourses tests once edit course dashboard reuses /courses/overview endpoint
func TestGetCourses_HappyPath(t *testing.T) {
	t.Run("returns courses with sections and materials successfully", func(t *testing.T) {
		videoSectionID := uuid.New()
		quizSectionID := uuid.New()
		materialID := uuid.New()
		storageKey := uuid.New()
		materialStorageKey := uuid.New()

		expected := []*domain.AllCourseLegacy{
			{
				ID:                testhelpers.Course.ID,
				Title:             testhelpers.Course.Title,
				Description:       testhelpers.Course.Description,
				CompletionTitle:   testhelpers.Course.CompletionTitle,
				CompletionMessage: testhelpers.Course.CompletionMessage,
				Sections: []domain.CourseSection{
					&domain.VideoSection{
						ID:         videoSectionID,
						Title:      "Intro Video",
						Position:   0,
						StorageKey: storageKey,
						Type:       domain.SectionTypeVideo,
					},
					&domain.QuizSectionLegacy{
						ID:       quizSectionID,
						Position: 1,
						Type:     domain.SectionTypeQuiz,
					},
				},
				Materials: []domain.CourseMaterial{
					{
						ID:         materialID,
						Name:       "Study Guide",
						Position:   0,
						StorageKey: materialStorageKey,
					},
				},
			},
			{
				ID:                uuid.New(),
				Title:             "Course 2",
				Description:       "Description 2",
				CompletionTitle:   "Completion Title 2",
				CompletionMessage: "Completion Message 2",
				Sections:          []domain.CourseSection{},
				Materials:         []domain.CourseMaterial{},
			},
		}

		mockRepo := &mocks.CourseRepositoryMock{
			GetAllCoursesFunc: func(ctx context.Context) ([]*domain.AllCourseLegacy, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, nil, "courses")

		err := h.GetCourses(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []*domain.AllCourseLegacy
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("courses mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAllCoursesCalls()), 1, testhelpers.GetCoursesHandlerName)
	})

	t.Run("returns empty slice when no courses exist", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			GetAllCoursesFunc: func(ctx context.Context) ([]*domain.AllCourseLegacy, error) {
				return []*domain.AllCourseLegacy{}, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, nil, "courses")

		err := h.GetCourses(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []*domain.AllCourseLegacy
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(actual) != 0 {
			t.Errorf("expected empty slice, got %v", actual)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAllCoursesCalls()), 1, testhelpers.GetCoursesHandlerName)
	})
}

func TestGetCourses_UnhappyPath(t *testing.T) {
	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			GetAllCoursesFunc: func(ctx context.Context) ([]*domain.AllCourseLegacy, error) {
				return nil, stdErrors.New("db error")
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, nil, "courses")

		err := h.GetCourses(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("courses"))
	})
}

func TestEditCourse_HappyPath(t *testing.T) {
	t.Run("edits course successfully", func(t *testing.T) {
		expected := testhelpers.Course

		mockRepo := &mocks.CourseRepositoryMock{
			EditCourseFunc: func(_ context.Context, _ *domain.EditCourseParams) (*domain.Course, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, validEditCourseRequest(), "edit-course")

		err := h.EditCourse(ctx)
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

		testhelpers.AssertRepoCalls(t, len(mockRepo.EditCourseCalls()), 1, testhelpers.EditCourseHandlerName)
	})
}

func TestEditCourse_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		reqBody        any
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	existingVideoSection := handlers.EditVideoSectionParams{
		Type:         domain.SectionTypeVideo,
		ID:           uuid.New().String(),
		IsNewSection: false,
		Title:        "Video Title",
		StorageKey:   uuid.New().String(),
		Position:     0,
	}

	tests := []testCase{
		{
			name:           "validation - missing course ID",
			reqBody:        withCourseID(validEditCourseRequest(), ""),
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name:           "validation - missing title",
			reqBody:        withTitle(validEditCourseRequest(), ""),
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name:           "validation - missing description",
			reqBody:        withDescription(validEditCourseRequest(), ""),
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name:           "invalid course ID",
			reqBody:        withCourseID(validEditCourseRequest(), "not-a-uuid"),
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "invalid video section storage key",
			reqBody: withSections(validEditCourseRequest(), []handlers.EditSectionParams{
				{Video: &handlers.EditVideoSectionParams{
					Type:         domain.SectionTypeVideo,
					IsNewSection: true,
					Title:        "Video",
					StorageKey:   "not-a-uuid",
					Position:     0,
				}},
			}),
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "existing video section missing ID",
			reqBody: withSections(validEditCourseRequest(), []handlers.EditSectionParams{
				{Video: &handlers.EditVideoSectionParams{
					Type:         domain.SectionTypeVideo,
					IsNewSection: false,
					Title:        "Video",
					StorageKey:   uuid.New().String(),
					Position:     0,
					// ID intentionally omitted — will fail UUID parse
				}},
			}),
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "quiz section - no questions",
			reqBody: withSections(validEditCourseRequest(), []handlers.EditSectionParams{
				{Quiz: &handlers.EditQuizSectionParams{
					Type:         domain.SectionTypeQuiz,
					IsNewSection: true,
					Position:     0,
					Questions:    []handlers.EditQuizQuestionParams{},
				}},
			}),
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "unknown section type",
			reqBody: struct {
				CourseID     string           `json:"edited_course_id"`
				EditedCourse map[string]any   `json:"edited_course"`
				DeletedIDs   map[string][]any `json:"deleted_section_ids_map"`
				DeletedMats  []any            `json:"deleted_materials_ids"`
			}{
				CourseID: uuid.New().String(),
				EditedCourse: map[string]any{
					"title":             "T",
					"description":       "D",
					"completionTitle":   "CT",
					"completionMessage": "CM",
					"sections":          []map[string]any{{"type": "unknown"}},
					"materials":         []any{},
				},
				DeletedIDs:  map[string][]any{"videoSectionIds": {}, "quizSectionIds": {}, "questionIds": {}, "answerIds": {}},
				DeletedMats: []any{},
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidRequestBody,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: withSections(validEditCourseRequest(), []handlers.EditSectionParams{
				{Video: &existingVideoSection},
			}),
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Updating("course"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Course: &mocks.CourseRepositoryMock{
						EditCourseFunc: func(_ context.Context, _ *domain.EditCourseParams) (*domain.Course, error) {
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
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "edit-course")
			err := h.EditCourse(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}

func validEditCourseRequest() handlers.EditCourseRequest {
	existingVideoID := uuid.New().String()
	existingVideoStorageKey := uuid.New().String()
	existingQuizID := uuid.New().String()
	existingQuestionID := uuid.New().String()
	existingAnswerID := uuid.New().String()
	existingMaterialID := uuid.New().String()
	existingMaterialStorageKey := uuid.New().String()
	deletedVideoSectionID := uuid.New().String()
	deletedQuizSectionID := uuid.New().String()
	deletedQuestionID := uuid.New().String()
	deletedAnswerID := uuid.New().String()
	deletedMaterialID := uuid.New().String()
	newVideoStorageKey := uuid.New().String()

	return handlers.EditCourseRequest{
		CourseID: uuid.New().String(),
		EditedCourse: handlers.EditedCourseFields{
			Title:             "Updated Title",
			Description:       "Updated Description",
			CompletionTitle:   "Updated Completion Title",
			CompletionMessage: "Updated Completion Message",
			Materials: []handlers.EditMaterialParams{
				{
					ID:         existingMaterialID,
					Name:       "Updated Study Guide",
					StorageKey: existingMaterialStorageKey,
					Position:   0,
				},
			},
			Sections: []handlers.EditSectionParams{
				{Video: &handlers.EditVideoSectionParams{
					Type:         domain.SectionTypeVideo,
					ID:           existingVideoID,
					IsNewSection: false,
					Title:        "Updated Video Title",
					StorageKey:   existingVideoStorageKey,
					Position:     0,
				}},
				{Quiz: &handlers.EditQuizSectionParams{
					Type:         domain.SectionTypeQuiz,
					ID:           existingQuizID,
					IsNewSection: false,
					Position:     1,
					Questions: []handlers.EditQuizQuestionParams{
						{
							ID:            existingQuestionID,
							Question:      "Updated question?",
							Position:      0,
							IsMultiAnswer: false,
							Answers: []handlers.EditQuizAnswerParams{
								{
									ID:              existingAnswerID,
									Answer:          "Updated answer",
									IsCorrectAnswer: true,
									Position:        0,
								},
							},
						},
					},
				}},
				{Video: &handlers.EditVideoSectionParams{
					Type:         domain.SectionTypeVideo,
					IsNewSection: true,
					Title:        "New Video Section",
					StorageKey:   newVideoStorageKey,
					Position:     2,
				}},
			},
		},
		DeletedSectionIDs: handlers.DeletedSectionIDs{
			VideoSectionIDs: []string{deletedVideoSectionID},
			QuizSectionIDs:  []string{deletedQuizSectionID},
			QuestionIDs:     []string{deletedQuestionID},
			AnswerIDs:       []string{deletedAnswerID},
		},
		DeletedMaterialIDs: []string{deletedMaterialID},
	}
}

func withCourseID(r handlers.EditCourseRequest, id string) handlers.EditCourseRequest {
	r.CourseID = id
	return r
}

func withTitle(r handlers.EditCourseRequest, title string) handlers.EditCourseRequest {
	r.EditedCourse.Title = title
	return r
}

func withDescription(r handlers.EditCourseRequest, description string) handlers.EditCourseRequest {
	r.EditedCourse.Description = description
	return r
}

func withSections(r handlers.EditCourseRequest, sections []handlers.EditSectionParams) handlers.EditCourseRequest {
	r.EditedCourse.Sections = sections
	return r
}
