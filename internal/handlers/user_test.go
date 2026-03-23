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

func TestGetUsersAndAssignedCourses_HappyPath(t *testing.T) {
	t.Run("returns users with assigned courses successfully", func(t *testing.T) {
		expected := []domain.UserWithAssignedCourses{
			{
				ID:        "user-1",
				Name:      "Alice",
				Email:     "alice@example.com",
				CourseIDs: []uuid.UUID{uuid.New(), uuid.New()},
			},
			{
				ID:        "user-2",
				Name:      "Bob",
				Email:     "bob@example.com",
				CourseIDs: []uuid.UUID{},
			},
		}

		mockUserRepo := &mocks.UserRepositoryMock{
			GetUsersAndAssignedCoursesFunc: func(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{User: mockUserRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, struct{}{}, "users-to-courses")

		err := h.GetUsersAndAssignedCourses(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []domain.UserWithAssignedCourses
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("users with assigned courses mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUsersAndAssignedCoursesCalls()), 1, testhelpers.GetUsersAndAssignedCoursesHandlerName)
	})

	t.Run("returns empty slice when no users exist", func(t *testing.T) {
		mockUserRepo := &mocks.UserRepositoryMock{
			GetUsersAndAssignedCoursesFunc: func(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
				return []domain.UserWithAssignedCourses{}, nil
			},
		}

		h := &handlers.Handlers{User: mockUserRepo}

		ctx, rec := testhelpers.SetupEchoContext(t, struct{}{}, "users-to-courses")

		err := h.GetUsersAndAssignedCourses(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []domain.UserWithAssignedCourses
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(actual) != 0 {
			t.Errorf("expected empty slice, got %v", actual)
		}

		testhelpers.AssertRepoCalls(t, len(mockUserRepo.GetUsersAndAssignedCoursesCalls()), 1, testhelpers.GetUsersAndAssignedCoursesHandlerName)
	})
}

func TestGetUsersAndAssignedCourses_UnhappyPath(t *testing.T) {
	t.Run("internal server error", func(t *testing.T) {
		mockUserRepo := &mocks.UserRepositoryMock{
			GetUsersAndAssignedCoursesFunc: func(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
				return nil, stdErrors.New("db error")
			},
		}

		h := &handlers.Handlers{User: mockUserRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, struct{}{}, "users-to-courses")

		err := h.GetUsersAndAssignedCourses(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("users with assigned courses"))
	})
}
