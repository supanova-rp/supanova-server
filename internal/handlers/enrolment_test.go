package handlers_test

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func TestUpdateCourseEnrolment(t *testing.T) {
	t.Run("enrols user successfully when IsEnrolled is false", func(t *testing.T) {
		courseID := uuid.New()

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			EnrolInCourseFunc: func(ctx context.Context, params sqlc.EnrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrolment: mockEnrolmentRepo,
		}

		reqBody := handlers.UpdateCourseEnrolmentParams{
			CourseID:   courseID.String(),
			IsEnrolled: false,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "update-users-to-courses")

		err := h.UpdateCourseEnrolment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 1, testhelpers.EnrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 0, testhelpers.DisenrollUserInCourseHandlerName)
	})

	t.Run("disenrolls user successfully when IsEnrolled is true", func(t *testing.T) {
		courseID := uuid.New()

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			DisenrolInCourseFunc: func(ctx context.Context, params sqlc.DisenrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrolment: mockEnrolmentRepo,
		}

		reqBody := handlers.UpdateCourseEnrolmentParams{
			CourseID:   courseID.String(),
			IsEnrolled: true,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "update-users-to-courses")

		err := h.UpdateCourseEnrolment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 1, testhelpers.DisenrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 0, testhelpers.EnrollUserInCourseHandlerName)
	})

	t.Run("validation error - missing course_id", func(t *testing.T) {
		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			EnrolInCourseFunc: func(ctx context.Context, params sqlc.EnrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrolment: mockEnrolmentRepo,
		}

		reqBody := handlers.UpdateCourseEnrolmentParams{
			IsEnrolled: false,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "update-users-to-courses")

		err := h.UpdateCourseEnrolment(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 0, testhelpers.EnrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 0, testhelpers.DisenrollUserInCourseHandlerName)
	})

	t.Run("validation error - invalid uuid format", func(t *testing.T) {
		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			EnrolInCourseFunc: func(ctx context.Context, params sqlc.EnrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrolment: mockEnrolmentRepo,
		}

		reqBody := handlers.UpdateCourseEnrolmentParams{
			CourseID:   "invalid-uuid",
			IsEnrolled: false,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "update-users-to-courses")

		err := h.UpdateCourseEnrolment(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidUUID)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 0, testhelpers.EnrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 0, testhelpers.DisenrollUserInCourseHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		courseID := uuid.New()

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			EnrolInCourseFunc: func(ctx context.Context, params sqlc.EnrolInCourseParams) error {
				return stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Enrolment: mockEnrolmentRepo,
		}

		reqBody := handlers.UpdateCourseEnrolmentParams{
			CourseID:   courseID.String(),
			IsEnrolled: false,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "update-users-to-courses")

		err := h.UpdateCourseEnrolment(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Creating("update-users-to-courses"))
	})
}
