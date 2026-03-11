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

func TestGetCourses_HappyPath(t *testing.T) {
	t.Run("returns courses with sections and materials successfully", func(t *testing.T) {
		videoSectionID := uuid.New()
		quizSectionID := uuid.New()
		materialID := uuid.New()
		storageKey := uuid.New()
		materialStorageKey := uuid.New()

		expected := []*domain.Course{
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
					&domain.QuizSection{
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
			GetAllCoursesFunc: func(ctx context.Context) ([]*domain.Course, error) {
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

		var actual []*domain.Course
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
			GetAllCoursesFunc: func(ctx context.Context) ([]*domain.Course, error) {
				return []*domain.Course{}, nil
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

		var actual []*domain.Course
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
			GetAllCoursesFunc: func(ctx context.Context) ([]*domain.Course, error) {
				return nil, stdErrors.New("db error")
			},
		}

		h := &handlers.Handlers{Course: mockRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, nil, "courses")

		err := h.GetCourses(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("courses"))
	})
}
