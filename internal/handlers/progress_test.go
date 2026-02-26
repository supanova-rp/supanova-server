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

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
	"github.com/supanova-rp/supanova-server/internal/services/email"
)

func TestGetProgress_HappyPath(t *testing.T) {
	t.Run("returns progress successfully", func(t *testing.T) {
		expected := testhelpers.Progress

		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params domain.GetProgressParams) (*domain.Progress, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Progress: mockRepo}

		req := handlers.GetProgressParams{
			CourseID: testhelpers.Course.ID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "progress")

		err := h.GetProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
		}

		var actual domain.Progress
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetProgressCalls()), 1, testhelpers.GetProgressHandlerName)
	})

	t.Run("progress not found returns empty struct", func(t *testing.T) {
		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params domain.GetProgressParams) (*domain.Progress, error) {
				return nil, pgx.ErrNoRows
			},
		}

		h := &handlers.Handlers{Progress: mockRepo}

		req := handlers.GetProgressParams{
			CourseID: testhelpers.Course.ID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "progress")

		err := h.GetProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
		}

		expected := &domain.Progress{
			CompletedIntro:      false,
			CompletedSectionIDs: []uuid.UUID{},
		}

		var actual domain.Progress
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetProgressCalls()), 1, testhelpers.GetProgressHandlerName)
	})
}

func TestGetProgress_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		reqBody        handlers.GetProgressParams
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "validation - missing courseId",
			reqBody:        handlers.GetProgressParams{},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "validation - invalid uuid",
			reqBody: handlers.GetProgressParams{
				CourseID: "invalid-uuid",
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.GetProgressParams{
				CourseID: testhelpers.Course.ID.String(),
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("user progress"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Progress: &mocks.ProgressRepositoryMock{
						GetProgressFunc: func(ctx context.Context, params domain.GetProgressParams) (*domain.Progress, error) {
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
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "progress")
			err := h.GetProgress(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}

func TestUpdateProgress_HappyPath(t *testing.T) {
	t.Run("updates progress successfully", func(t *testing.T) {
		mockRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params domain.UpdateProgressParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{Progress: mockRepo}

		req := handlers.UpdateProgressParams{
			CourseID:  testhelpers.Course.ID.String(),
			SectionID: uuid.New().String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "progress")

		err := h.UpdateProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.UpdateProgressCalls()), 1, testhelpers.UpdateProgressHandlerName)
	})
}

func TestUpdateProgress_UnhappyPath(t *testing.T) {
	courseID := testhelpers.Course.ID.String()
	sectionID := uuid.New().String()

	type testCase struct {
		name           string
		reqBody        handlers.UpdateProgressParams
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name: "validation - missing courseId",
			reqBody: handlers.UpdateProgressParams{
				SectionID: sectionID,
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "validation - missing sectionId",
			reqBody: handlers.UpdateProgressParams{
				CourseID: courseID,
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "validation - invalid courseId",
			reqBody: handlers.UpdateProgressParams{
				CourseID:  "invalid-uuid",
				SectionID: sectionID,
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "validation - invalid sectionId",
			reqBody: handlers.UpdateProgressParams{
				CourseID:  courseID,
				SectionID: "invalid-uuid",
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.UpdateProgressParams{
				CourseID:  courseID,
				SectionID: sectionID,
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Updating("user progress"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Progress: &mocks.ProgressRepositoryMock{
						UpdateProgressFunc: func(ctx context.Context, params domain.UpdateProgressParams) error {
							return stdErrors.New("db error")
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "progress")
			err := h.UpdateProgress(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}
func TestSetCourseCompleted_HappyPath(t *testing.T) {
	t.Run("already completed course - no update", func(t *testing.T) {
		courseID := testhelpers.Course.ID.String()
		courseName := testhelpers.Course.Title

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			HasCompletedCourseFunc: func(ctx context.Context, params domain.HasCompletedCourseParams) (bool, error) {
				return true, nil
			},
		}
		mockUserRepo := &mocks.UserRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
			User:     mockUserRepo,
		}

		req := &handlers.SetCourseCompletedParams{
			CourseID:   courseID,
			CourseName: courseName,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "set-course-completed")

		err := h.SetCourseCompleted(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 1, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 0, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 0, testhelpers.GetUserHandlerName)
	})

	t.Run("sets course completed - first time", func(t *testing.T) {
		courseID := testhelpers.Course.ID.String()
		courseName := testhelpers.Course.Title

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			HasCompletedCourseFunc: func(ctx context.Context, params domain.HasCompletedCourseParams) (bool, error) {
				return false, nil
			},
			SetCourseCompletedFunc: func(ctx context.Context, params domain.SetCourseCompletedParams) error {
				return nil
			},
		}
		mockUserRepo := &mocks.UserRepositoryMock{
			GetUserFunc: func(ctx context.Context, id string) (*domain.User, error) {
				return testhelpers.User, nil
			},
		}
		mockEmailRepo := &mocks.EmailServiceMock{
			SendFunc: func(ctx context.Context, params email.EmailParams, templateName, emailName string) error {
				return nil
			},
			GetTemplateNamesFunc: func() *email.TemplateNames {
				return &email.TemplateNames{CourseCompletion: ""}
			},
			GetEmailNamesFunc: func() *email.EmailNames {
				return &email.EmailNames{CourseCompletion: ""}
			},
		}

		h := &handlers.Handlers{
			Progress:     mockProgressRepo,
			User:         mockUserRepo,
			EmailService: mockEmailRepo,
		}

		req := &handlers.SetCourseCompletedParams{
			CourseID:   courseID,
			CourseName: courseName,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "set-course-completed")

		err := h.SetCourseCompleted(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 1, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 1, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 1, testhelpers.GetUserHandlerName)
	})
}

func TestSetCourseCompleted_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		reqBody        *handlers.SetCourseCompletedParams
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	courseID := testhelpers.Course.ID.String()
	courseName := testhelpers.Course.Title

	tests := []testCase{
		{
			name:           "validation - missing courseId",
			reqBody:        &handlers.SetCourseCompletedParams{CourseName: courseName},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name:           "validation - missing courseName",
			reqBody:        &handlers.SetCourseCompletedParams{CourseID: courseID},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name:           "validation - invalid courseId uuid",
			reqBody:        &handlers.SetCourseCompletedParams{CourseID: "invalid-uuid", CourseName: courseName},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: &handlers.SetCourseCompletedParams{
				CourseID:   courseID,
				CourseName: courseName,
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Updating("user progress"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Progress: &mocks.ProgressRepositoryMock{
						HasCompletedCourseFunc: func(ctx context.Context, params domain.HasCompletedCourseParams) (bool, error) {
							return false, nil
						},
						SetCourseCompletedFunc: func(ctx context.Context, params domain.SetCourseCompletedParams) error {
							return nil
						},
					},
					User: &mocks.UserRepositoryMock{
						GetUserFunc: func(ctx context.Context, id string) (*domain.User, error) {
							return nil, pgx.ErrNoRows
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "set-course-completed")
			err := h.SetCourseCompleted(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}

func TestResetProgress_HappyPath(t *testing.T) {
	t.Run("resets progress successfully", func(t *testing.T) {
		mockRepo := &mocks.ProgressRepositoryMock{
			ResetProgressFunc: func(ctx context.Context, params domain.ResetProgressParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{Progress: mockRepo}

		req := handlers.ResetProgressParams{
			CourseID: testhelpers.Course.ID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "reset-progress")

		err := h.ResetProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.ResetProgressCalls()), 1, testhelpers.ResetProgressHandlerName)
	})
}

func TestResetProgress_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		reqBody        handlers.ResetProgressParams
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "validation - missing courseId",
			reqBody:        handlers.ResetProgressParams{},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "validation - invalid uuid",
			reqBody: handlers.ResetProgressParams{
				CourseID: "invalid-uuid",
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Progress: &mocks.ProgressRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.ResetProgressParams{
				CourseID: testhelpers.Course.ID.String(),
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Updating("user progress"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Progress: &mocks.ProgressRepositoryMock{
						ResetProgressFunc: func(ctx context.Context, params domain.ResetProgressParams) error {
							return stdErrors.New("db error")
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "admin/reset-progress")
			err := h.ResetProgress(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}

func TestGetAllProgress_HappyPath(t *testing.T) {
	t.Run("returns all progress successfully", func(t *testing.T) {
		sectionTitle := "Section 1"
		expected := []*domain.FullProgress{
			{
				UserID:   "user-1",
				UserName: "User A",
				Email:    "usera@test.com",
				Progress: []*domain.FullUserProgress{
					{
						CourseID:        testhelpers.Course.ID,
						CourseName:      testhelpers.Course.Title,
						CompletedIntro:  true,
						CompletedCourse: false,
						CourseSectionProgress: []domain.CourseSectionProgress{
							{
								ID:        uuid.New(),
								Title:     &sectionTitle,
								Type:      "video",
								Completed: true,
							},
						},
					},
				},
			},
		}

		mockRepo := &mocks.ProgressRepositoryMock{
			GetAllProgressFunc: func(ctx context.Context) ([]*domain.FullProgress, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Progress: mockRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, nil, "progress/all")

		err := h.GetAllProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []*domain.FullProgress
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAllProgressCalls()), 1, testhelpers.GetAllProgressHandlerName)
	})
}

func TestGetAllProgress_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "progress not found",
			wantStatus:     http.StatusNotFound,
			expectedErrMsg: errors.NotFound("user progress"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Progress: &mocks.ProgressRepositoryMock{
						GetAllProgressFunc: func(ctx context.Context) ([]*domain.FullProgress, error) {
							return nil, pgx.ErrNoRows
						},
					},
				}
			},
		},
		{
			name:           "internal server error",
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("user progress"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Progress: &mocks.ProgressRepositoryMock{
						GetAllProgressFunc: func(ctx context.Context) ([]*domain.FullProgress, error) {
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
			ctx, _ := testhelpers.SetupEchoContext(t, nil, "progress/all")
			err := h.GetAllProgress(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}
