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
	userMocks "github.com/supanova-rp/supanova-server/internal/domain/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
	"github.com/supanova-rp/supanova-server/internal/services/email"
)

func TestGetProgress(t *testing.T) {
	t.Run("returns progress successfully", func(t *testing.T) {
		expected := testhelpers.Progress

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params domain.GetProgressParams) (*domain.Progress, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
		}

		params := handlers.GetProgressParams{
			CourseID: testhelpers.Course.ID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.GetProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual domain.Progress
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.GetProgressCalls()), 1, testhelpers.GetProgressHandlerName)
	})

	t.Run("validation error - missing courseId", func(t *testing.T) {
		mockRepo := &mocks.ProgressRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.GetProgressParams{}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.GetProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetProgressCalls()), 0, testhelpers.GetProgressHandlerName)
	})

	t.Run("validation error - invalid uuid format", func(t *testing.T) {
		mockRepo := &mocks.ProgressRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.GetProgressParams{
			CourseID: "invalid-uuid",
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.GetProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidUUID)
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetProgressCalls()), 0, testhelpers.GetProgressHandlerName)
	})

	t.Run("progress not found - returns empty progress struct", func(t *testing.T) {
		courseID := testhelpers.Course.ID

		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params domain.GetProgressParams) (*domain.Progress, error) {
				return nil, pgx.ErrNoRows
			},
		}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.GetProgressParams{
			CourseID: courseID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.GetProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		expected := &domain.Progress{
			CompletedIntro:      false,
			CompletedSectionIDs: []uuid.UUID{},
		}

		var actual domain.Progress
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetProgressCalls()), 1, testhelpers.GetProgressHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		courseID := testhelpers.Course.ID

		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params domain.GetProgressParams) (*domain.Progress, error) {
				return nil, stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.GetProgressParams{
			CourseID: courseID.String(),
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.GetProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("user progress"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetProgressCalls()), 1, testhelpers.GetProgressHandlerName)
	})
}

func TestUpdateProgress(t *testing.T) {
	t.Run("updates progress successfully", func(t *testing.T) {
		courseID := testhelpers.Course.ID
		sectionID := uuid.New()

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params domain.UpdateProgressParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
		}

		params := handlers.UpdateProgressParams{
			CourseID:  courseID.String(),
			SectionID: sectionID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.UpdateProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.UpdateProgressCalls()), 1, testhelpers.UpdateProgressHandlerName)
	})

	t.Run("validation error - missing courseId", func(t *testing.T) {
		sectionID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.UpdateProgressParams{
			SectionID: sectionID.String(),
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.UpdateProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.UpdateProgressCalls()), 0, testhelpers.UpdateProgressHandlerName)
	})

	t.Run("validation error - missing sectionId", func(t *testing.T) {
		courseID := testhelpers.Course.ID

		mockRepo := &mocks.ProgressRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.UpdateProgressParams{
			CourseID: courseID.String(),
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.UpdateProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.UpdateProgressCalls()), 0, testhelpers.UpdateProgressHandlerName)
	})

	t.Run("validation error - invalid courseId uuid format", func(t *testing.T) {
		sectionID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.UpdateProgressParams{
			CourseID:  "invalid-uuid",
			SectionID: sectionID.String(),
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.UpdateProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidUUID)
		testhelpers.AssertRepoCalls(t, len(mockRepo.UpdateProgressCalls()), 0, testhelpers.UpdateProgressHandlerName)
	})

	t.Run("validation error - invalid sectionId uuid format", func(t *testing.T) {
		courseID := testhelpers.Course.ID

		mockRepo := &mocks.ProgressRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.UpdateProgressParams{
			CourseID:  courseID.String(),
			SectionID: "invalid-uuid",
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.UpdateProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidUUID)
		testhelpers.AssertRepoCalls(t, len(mockRepo.UpdateProgressCalls()), 0, testhelpers.UpdateProgressHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		courseID := testhelpers.Course.ID
		sectionID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params domain.UpdateProgressParams) error {
				return stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		params := handlers.UpdateProgressParams{
			CourseID:  courseID.String(),
			SectionID: sectionID.String(),
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "progress")

		err := h.UpdateProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Updating("user progress"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.UpdateProgressCalls()), 1, testhelpers.UpdateProgressHandlerName)
	})
}

func TestCourseCompleted(t *testing.T) {
	t.Run("sets course to completed with previous completion", func(t *testing.T) {
		courseID := testhelpers.Course.ID
		courseName := testhelpers.Course.Title

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			HasCompletedCourseFunc: func(ctx context.Context, params domain.HasCompletedCourseParams) (bool, error) {
				return true, nil
			},
		}
		mockUserRepo := &userMocks.UserRepositoryMock{}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
		}

		params := &handlers.SetCourseCompletedParams{
			CourseID:   courseID.String(),
			CourseName: courseName,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, params, "set-course-completed")

		err := h.SetCourseCompleted(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 1, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 0, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 0, testhelpers.GetUserHandlerName)
	})

	t.Run("sets course to completed with no previous completion", func(t *testing.T) {
		courseID := testhelpers.Course.ID
		courseName := testhelpers.Course.Title

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			HasCompletedCourseFunc: func(ctx context.Context, params domain.HasCompletedCourseParams) (bool, error) {
				return false, nil
			},
			SetCourseCompletedFunc: func(ctx context.Context, params domain.SetCourseCompletedParams) error {
				return nil
			},
		}
		mockUserRepo := &userMocks.UserRepositoryMock{
			GetUserFunc: func(ctx context.Context, id string) (*domain.User, error) {
				return testhelpers.User, nil
			},
		}
		mockEmailRepo := &mocks.EmailServiceMock{
			SendFunc: func(ctx context.Context, params email.EmailParams, templateName, emailName string) error {
				return nil
			},
			GetTemplateNamesFunc: func() *email.TemplateNames {
				return &email.TemplateNames{
					CourseCompletion: "",
				}
			},
			GetEmailNamesFunc: func() *email.EmailNames {
				return &email.EmailNames{
					CourseCompletion: "",
				}
			},
		}

		h := &handlers.Handlers{
			Progress:     mockProgressRepo,
			User:         mockUserRepo,
			EmailService: mockEmailRepo,
		}

		params := &handlers.SetCourseCompletedParams{
			CourseID:   courseID.String(),
			CourseName: courseName,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, params, "set-course-completed")

		err := h.SetCourseCompleted(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 1, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 1, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 1, testhelpers.GetUserHandlerName)
	})

	t.Run("validation error - missing courseId", func(t *testing.T) {
		courseName := testhelpers.Course.Title

		mockProgressRepo := &mocks.ProgressRepositoryMock{}
		mockUserRepo := &userMocks.UserRepositoryMock{}

		params := &handlers.SetCourseCompletedParams{
			CourseName: courseName,
		}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "set-course-completed")

		err := h.SetCourseCompleted(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 0, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 0, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 0, testhelpers.GetUserHandlerName)
	})

	t.Run("validation error - missing courseName", func(t *testing.T) {
		courseID := testhelpers.Course.ID

		mockProgressRepo := &mocks.ProgressRepositoryMock{}
		mockUserRepo := &userMocks.UserRepositoryMock{}

		params := &handlers.SetCourseCompletedParams{
			CourseID: courseID.String(),
		}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "set-course-completed")

		err := h.SetCourseCompleted(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 0, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 0, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 0, testhelpers.GetUserHandlerName)
	})

	t.Run("validation error - invalid courseId uuid format", func(t *testing.T) {
		mockProgressRepo := &mocks.ProgressRepositoryMock{}
		mockUserRepo := &userMocks.UserRepositoryMock{}

		params := &handlers.SetCourseCompletedParams{
			CourseID: "invalid-uuid",
		}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "set-course-completed")

		err := h.SetCourseCompleted(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 0, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 0, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 0, testhelpers.GetUserHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		courseID := testhelpers.Course.ID
		courseName := testhelpers.Course.Title

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			HasCompletedCourseFunc: func(ctx context.Context, params domain.HasCompletedCourseParams) (bool, error) {
				return false, nil
			},
			SetCourseCompletedFunc: func(ctx context.Context, params domain.SetCourseCompletedParams) error {
				return nil
			},
		}
		mockUserRepo := &userMocks.UserRepositoryMock{
			GetUserFunc: func(ctx context.Context, id string) (*domain.User, error) {
				return nil, pgx.ErrNoRows
			},
		}

		params := &handlers.SetCourseCompletedParams{
			CourseID:   courseID.String(),
			CourseName: courseName,
		}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
			User:     mockUserRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, params, "set-course-completed")

		err := h.SetCourseCompleted(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Updating("user progress"))
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.HasCompletedCourseCalls()), 1, testhelpers.HasCompletedCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.SetCourseCompletedCalls()), 1, testhelpers.SetCourseCompletedHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUserCalls()), 1, testhelpers.GetUserHandlerName)
	})
}

func TestGetAllProgress(t *testing.T) {
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

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, nil, "progress/all")

		err := h.GetAllProgress(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []*domain.FullProgress
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAllProgressCalls()), 1, testhelpers.GetAllProgressHandlerName)
	})

	t.Run("progress not found", func(t *testing.T) {
		mockRepo := &mocks.ProgressRepositoryMock{
			GetAllProgressFunc: func(ctx context.Context) ([]*domain.FullProgress, error) {
				return nil, pgx.ErrNoRows
			},
		}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, nil, "progress/all")

		err := h.GetAllProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusNotFound, errors.NotFound("user progress"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAllProgressCalls()), 1, testhelpers.GetAllProgressHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mocks.ProgressRepositoryMock{
			GetAllProgressFunc: func(ctx context.Context) ([]*domain.FullProgress, error) {
				return nil, stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Progress: mockRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, nil, "progress/all")

		err := h.GetAllProgress(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("user progress"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetAllProgressCalls()), 1, testhelpers.GetAllProgressHandlerName)
	})
}
