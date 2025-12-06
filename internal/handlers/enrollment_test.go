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

func TestUpdateUserCourseEnrollment(t *testing.T) {
	t.Run("enrolls user successfully when isAssigned is false", func(t *testing.T) {
		courseID := uuid.New()

		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			EnrollUserInCourseFunc: func(ctx context.Context, params sqlc.EnrollUserInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrollment: mockEnrollmentRepo,
		}

		reqBody := handlers.UpdateUserCourseEnrollmentParams{
			CourseID:   courseID.String(),
			IsEnrolled: false,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "enrollment")

		err := h.UpdateUserCourseEnrollment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.EnrollUserInCourseCalls()), 1, testhelpers.EnrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.DisenrollUserInCourseCalls()), 0, testhelpers.DisenrollUserInCourseHandlerName)
	})

	t.Run("disenrolls user successfully when isAssigned is true", func(t *testing.T) {
		courseID := uuid.New()

		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			DisenrollUserInCourseFunc: func(ctx context.Context, params sqlc.DisenrollUserInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrollment: mockEnrollmentRepo,
		}

		reqBody := handlers.UpdateUserCourseEnrollmentParams{
			CourseID:   courseID.String(),
			IsEnrolled: true,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "enrollment")

		err := h.UpdateUserCourseEnrollment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.DisenrollUserInCourseCalls()), 1, testhelpers.DisenrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.EnrollUserInCourseCalls()), 0, testhelpers.EnrollUserInCourseHandlerName)
	})

	t.Run("validation error - missing course_id", func(t *testing.T) {
		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			EnrollUserInCourseFunc: func(ctx context.Context, params sqlc.EnrollUserInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrollment: mockEnrollmentRepo,
		}

		reqBody := map[string]interface{}{
			"isAssigned": false,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "enrollment")

		err := h.UpdateUserCourseEnrollment(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.EnrollUserInCourseCalls()), 0, testhelpers.EnrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.DisenrollUserInCourseCalls()), 0, testhelpers.DisenrollUserInCourseHandlerName)
	})

	t.Run("validation error - invalid uuid format", func(t *testing.T) {
		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			EnrollUserInCourseFunc: func(ctx context.Context, params sqlc.EnrollUserInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{
			Enrollment: mockEnrollmentRepo,
		}

		reqBody := handlers.UpdateUserCourseEnrollmentParams{
			CourseID:   "invalid-uuid",
			IsEnrolled: false,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "enrollment")

		err := h.UpdateUserCourseEnrollment(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidUUID)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.EnrollUserInCourseCalls()), 0, testhelpers.EnrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.DisenrollUserInCourseCalls()), 0, testhelpers.DisenrollUserInCourseHandlerName)
	})

	t.Run("internal server error when enrolling", func(t *testing.T) {
		courseID := uuid.New()

		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			EnrollUserInCourseFunc: func(ctx context.Context, params sqlc.EnrollUserInCourseParams) error {
				return stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Enrollment: mockEnrollmentRepo,
		}

		reqBody := handlers.UpdateUserCourseEnrollmentParams{
			CourseID:   courseID.String(),
			IsEnrolled: false,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "enrollment")

		err := h.UpdateUserCourseEnrollment(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Creating("enrollment"))
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.EnrollUserInCourseCalls()), 1, testhelpers.EnrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.DisenrollUserInCourseCalls()), 0, testhelpers.DisenrollUserInCourseHandlerName)
	})

	t.Run("internal server error when disenrolling", func(t *testing.T) {
		courseID := uuid.New()

		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			DisenrollUserInCourseFunc: func(ctx context.Context, params sqlc.DisenrollUserInCourseParams) error {
				return stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Enrollment: mockEnrollmentRepo,
		}

		reqBody := handlers.UpdateUserCourseEnrollmentParams{
			CourseID:   courseID.String(),
			IsEnrolled: true,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "enrollment")

		err := h.UpdateUserCourseEnrollment(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Deleting("enrollment"))
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.DisenrollUserInCourseCalls()), 1, testhelpers.DisenrollUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.EnrollUserInCourseCalls()), 0, testhelpers.EnrollUserInCourseHandlerName)
	})
}
