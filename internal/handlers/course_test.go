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

func TestAddCourse(t *testing.T) {
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

	t.Run("validation error - missing title", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{}

		h := &handlers.Handlers{Course: mockRepo}

		reqBody := handlers.AddCourseParams{
			Description: testhelpers.Course.Description,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0, testhelpers.AddCourseHandlerName)
	})

	t.Run("validation error - missing description", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{}

		h := &handlers.Handlers{Course: mockRepo}

		reqBody := handlers.AddCourseParams{
			Title: testhelpers.Course.Title,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0, testhelpers.AddCourseHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params *domain.AddCourseParams) (*domain.Course, error) {
				return nil, stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		reqBody := handlers.AddCourseParams{
			Title:             testhelpers.Course.Title,
			Description:       testhelpers.Course.Description,
			CompletionTitle:   testhelpers.Course.CompletionTitle,
			CompletionMessage: testhelpers.Course.CompletionMessage,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Creating("course"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 1, testhelpers.AddCourseHandlerName)
	})

	t.Run("validation error - video section missing title", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{}

		h := &handlers.Handlers{Course: mockRepo}

		reqBody := handlers.AddCourseParams{
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
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0, testhelpers.AddCourseHandlerName)
	})

	t.Run("validation error - quiz section no questions", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{}

		h := &handlers.Handlers{Course: mockRepo}

		reqBody := handlers.AddCourseParams{
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
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0, testhelpers.AddCourseHandlerName)
	})

	t.Run("unknown section type", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{}

		h := &handlers.Handlers{Course: mockRepo}

		reqBody := struct {
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
			Sections: []map[string]any{
				{"type": "unknown"},
			},
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidRequestBody)
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0, testhelpers.AddCourseHandlerName)
	})
}

func TestGetCoursesOverview(t *testing.T) {
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

		reqBody := struct{}{}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "course-titles")

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

	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			GetCoursesOverviewFunc: func(ctx context.Context) ([]domain.CourseOverview, error) {
				return nil, stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, struct{}{}, "course-titles")

		err := h.GetCoursesOverview(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("course overview"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetCoursesOverviewCalls()), 1, testhelpers.GetCoursesOverviewHandlerName)
	})
}

func TestGetCourseMaterials(t *testing.T) {
	courseID := testhelpers.Course.ID

	material1 := domain.CourseMaterial{
		ID:         uuid.New(),
		Name:       "Material 1",
		Position:   0,
		StorageKey: uuid.New(),
	}
	material2 := domain.CourseMaterial{
		ID:         uuid.New(),
		Name:       "Material 2",
		Position:   1,
		StorageKey: uuid.New(),
	}

	t.Run("returns materials with urls successfully", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseMaterialsFunc: func(ctx context.Context, id uuid.UUID) ([]domain.CourseMaterial, error) {
				return []domain.CourseMaterial{material1, material2}, nil
			},
		}

		mockObjectStorage := &mocks.ObjectStorageMock{
			GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
				return "https://cdn.example.com/" + key, nil
			},
		}

		h := &handlers.Handlers{
			Course:        mockCourseRepo,
			ObjectStorage: mockObjectStorage,
		}

		reqBody := handlers.GetCourseMaterialsParams{
			CourseID: courseID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "materials")

		err := h.GetCourseMaterials(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []domain.CourseMaterialWithURL
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(actual) != 2 {
			t.Fatalf("expected 2 materials, got %d", len(actual))
		}

		if actual[0].ID != material1.ID || actual[0].Name != material1.Name || actual[0].Position != material1.Position {
			t.Errorf("unexpected first material: %+v", actual[0])
		}
		if actual[0].URL == "" {
			t.Error("expected non-empty URL for material 1")
		}

		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseMaterialsCalls()), 1, testhelpers.GetCourseMaterialsHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockObjectStorage.GetCDNURLCalls()), 2, "GetCDNURL")
	})

	t.Run("validation error - missing course id", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{}

		h := &handlers.Handlers{Course: mockCourseRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, handlers.GetCourseMaterialsParams{}, "materials")

		err := h.GetCourseMaterials(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseMaterialsCalls()), 0, testhelpers.GetCourseMaterialsHandlerName)
	})

	t.Run("validation error - invalid uuid format", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{}

		h := &handlers.Handlers{Course: mockCourseRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, handlers.GetCourseMaterialsParams{CourseID: "invalid-uuid"}, "materials")

		err := h.GetCourseMaterials(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidUUID)
		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseMaterialsCalls()), 0, testhelpers.GetCourseMaterialsHandlerName)
	})

	t.Run("forbidden - user not enrolled", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{}

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			IsEnrolledFunc: func(ctx context.Context, params domain.IsEnrolledParams) (bool, error) {
				return false, nil
			},
		}

		h := &handlers.Handlers{
			Course:    mockCourseRepo,
			Enrolment: mockEnrolmentRepo,
		}

		reqBody := handlers.GetCourseMaterialsParams{CourseID: courseID.String()}
		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "materials", testhelpers.WithRole(config.UserRole))

		err := h.GetCourseMaterials(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusForbidden, errors.Forbidden("course materials"))
		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseMaterialsCalls()), 0, testhelpers.GetCourseMaterialsHandlerName)
	})

	t.Run("internal server error from repo", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseMaterialsFunc: func(ctx context.Context, id uuid.UUID) ([]domain.CourseMaterial, error) {
				return nil, stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{Course: mockCourseRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, handlers.GetCourseMaterialsParams{CourseID: courseID.String()}, "materials")

		err := h.GetCourseMaterials(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("course materials"))
		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseMaterialsCalls()), 1, testhelpers.GetCourseMaterialsHandlerName)
	})

	t.Run("internal server error from object storage", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseMaterialsFunc: func(ctx context.Context, id uuid.UUID) ([]domain.CourseMaterial, error) {
				return []domain.CourseMaterial{material1}, nil
			},
		}

		mockObjectStorage := &mocks.ObjectStorageMock{
			GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
				return "", stdErrors.New("cdn error")
			},
		}

		h := &handlers.Handlers{
			Course:        mockCourseRepo,
			ObjectStorage: mockObjectStorage,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, handlers.GetCourseMaterialsParams{CourseID: courseID.String()}, "materials")

		err := h.GetCourseMaterials(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("course materials"))
		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseMaterialsCalls()), 1, testhelpers.GetCourseMaterialsHandlerName)
	})
}
