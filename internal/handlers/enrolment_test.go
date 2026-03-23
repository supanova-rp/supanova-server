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

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			GetUsersAndAssignedCoursesFunc: func(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Enrolment: mockEnrolmentRepo}

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

		calls := len(mockEnrolmentRepo.GetUsersAndAssignedCoursesCalls())
		testhelpers.AssertRepoCalls(t, calls, 1, testhelpers.GetUsersAndAssignedCoursesHandlerName)
	})

	t.Run("returns empty slice when no users exist", func(t *testing.T) {
		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			GetUsersAndAssignedCoursesFunc: func(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
				return []domain.UserWithAssignedCourses{}, nil
			},
		}

		h := &handlers.Handlers{Enrolment: mockEnrolmentRepo}

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

		calls := len(mockEnrolmentRepo.GetUsersAndAssignedCoursesCalls())
		testhelpers.AssertRepoCalls(t, calls, 1, testhelpers.GetUsersAndAssignedCoursesHandlerName)
	})
}

func TestGetUsersAndAssignedCourses_UnhappyPath(t *testing.T) {
	t.Run("internal server error", func(t *testing.T) {
		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			GetUsersAndAssignedCoursesFunc: func(ctx context.Context) ([]domain.UserWithAssignedCourses, error) {
				return nil, stdErrors.New("db error")
			},
		}

		h := &handlers.Handlers{Enrolment: mockEnrolmentRepo}

		ctx, _ := testhelpers.SetupEchoContext(t, struct{}{}, "users-to-courses")

		err := h.GetUsersAndAssignedCourses(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("users with assigned courses"))
	})
}

func TestUpdateCourseEnrolment_HappyPath(t *testing.T) {
	t.Run("enrols user successfully when IsEnrolled is false", func(t *testing.T) {
		courseID := testhelpers.Course.ID.String()

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			EnrolInCourseFunc: func(ctx context.Context, params domain.EnrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{Enrolment: mockEnrolmentRepo}

		req := handlers.UpdateCourseEnrolmentParams{
			UserID:     testhelpers.TestUserID,
			CourseID:   courseID,
			IsEnrolled: false,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "enrolment")
		err := h.UpdateCourseEnrolment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 1, testhelpers.EnrolUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 0, testhelpers.DisenrolUserInCourseHandlerName)
	})

	t.Run("disenrols user successfully when IsEnrolled is true", func(t *testing.T) {
		courseID := testhelpers.Course.ID.String()

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			DisenrolInCourseFunc: func(ctx context.Context, params domain.DisenrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{Enrolment: mockEnrolmentRepo}

		req := handlers.UpdateCourseEnrolmentParams{
			UserID:     testhelpers.TestUserID,
			CourseID:   courseID,
			IsEnrolled: true,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "enrolment")
		err := h.UpdateCourseEnrolment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 1, testhelpers.DisenrolUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 0, testhelpers.EnrolUserInCourseHandlerName)
	})
}

func TestUpdateCourseEnrolment_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		reqBody        handlers.UpdateCourseEnrolmentParams
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	courseID := testhelpers.Course.ID.String()

	tests := []testCase{
		{
			name: "validation error - missing course id",
			reqBody: handlers.UpdateCourseEnrolmentParams{
				UserID:     testhelpers.TestUserID,
				IsEnrolled: false,
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Enrolment: &mocks.EnrolmentRepositoryMock{}}
			},
		},
		{
			name: "validation error - invalid uuid format",
			reqBody: handlers.UpdateCourseEnrolmentParams{
				UserID:     testhelpers.TestUserID,
				CourseID:   "invalid-uuid",
				IsEnrolled: false,
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Enrolment: &mocks.EnrolmentRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.UpdateCourseEnrolmentParams{
				UserID:     testhelpers.TestUserID,
				CourseID:   courseID,
				IsEnrolled: false,
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Creating("enrolment"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Enrolment: &mocks.EnrolmentRepositoryMock{
						EnrolInCourseFunc: func(ctx context.Context, params domain.EnrolInCourseParams) error {
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
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "enrolment")
			err := h.UpdateCourseEnrolment(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}
