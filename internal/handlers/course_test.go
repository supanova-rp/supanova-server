package handlers_test

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func TestGetCourse(t *testing.T) {
	t.Run("returns course successfully", func(t *testing.T) {
		expected := testhelpers.Course

		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return expected, nil
			},
		}

		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			IsEnrolledFunc: func(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error) {
				return true, nil
			},
		}

		h := &handlers.Handlers{
			Course:     mockCourseRepo,
			Enrollment: mockEnrollmentRepo,
		}
		ctx, rec := testhelpers.SetupEchoContext(
			t,
			fmt.Sprintf(`{"id":%q}`, testhelpers.Course.ID),
			"course",
			true,
		)

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
	})

	t.Run("validation error - missing id", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return nil, nil
			},
		}

		h := &handlers.Handlers{
			Course: mockRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, `{}`, "course", true) // missing id

		err := h.GetCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetCourseCalls()), 0, testhelpers.GetCourseHandlerName)
	})

	t.Run("validation error - invalid uuid format", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return nil, nil
			},
		}

		h := &handlers.Handlers{
			Course: mockRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, `{"id":"invalid-uuid"}`, "course", true)

		err := h.GetCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.InvalidUUID)
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetCourseCalls()), 0, testhelpers.GetCourseHandlerName)
	})

	t.Run("course not found", func(t *testing.T) {
		courseID := uuid.New()

		mockRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return nil, pgx.ErrNoRows
			},
		}

		h := &handlers.Handlers{
			Course: mockRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(
			t,
			fmt.Sprintf(`{"id":%q}`, courseID),
			"course",
			true,
		)

		err := h.GetCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusNotFound, errors.NotFound("course"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetCourseCalls()), 1, testhelpers.GetCourseHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		courseID := uuid.New()

		mockRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return nil, stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Course: mockRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(
			t,
			fmt.Sprintf(`{"id":%q}`, courseID),
			"course",
			true,
		)

		err := h.GetCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("course"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetCourseCalls()), 1, testhelpers.GetCourseHandlerName)
	})

	t.Run("forbidden - user not enrolled", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return testhelpers.Course, nil
			},
		}

		mockEnrollmentRepo := &mocks.EnrollmentRepositoryMock{
			IsEnrolledFunc: func(ctx context.Context, params sqlc.IsUserEnrolledInCourseParams) (bool, error) {
				return false, nil
			},
		}

		h := &handlers.Handlers{
			Course:     mockCourseRepo,
			Enrollment: mockEnrollmentRepo,
		}

		ctx, _ := testhelpers.SetupEchoContext(
			t,
			fmt.Sprintf(`{"id":%q}`, testhelpers.Course.ID),
			"course",
			true,
		)

		err := h.GetCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusForbidden, errors.Forbidden("course"))
		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseCalls()), 1, testhelpers.GetCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrollmentRepo.IsEnrolledCalls()), 1, testhelpers.GetCourseHandlerName)
	})
}

func TestAddCourse(t *testing.T) {
	t.Run("adds course successfully", func(t *testing.T) {
		expected := testhelpers.Course

		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		ctx, rec := testhelpers.SetupEchoContext(
			t,
			`{"title":"New Course","description":"New Description"}`,
			"course",
			true,
		)

		err := h.AddCourse(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		var actual domain.Course
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("course mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 1, testhelpers.AddCourseHandlerName)
	})

	t.Run("validation error - missing title", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return nil, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		ctx, _ := testhelpers.SetupEchoContext(
			t,
			fmt.Sprintf(`{"description":%q}`, testhelpers.Course.Description),
			"course",
			true,
		)

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0, testhelpers.AddCourseHandlerName)
	})

	t.Run("validation error - missing description", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return nil, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		ctx, _ := testhelpers.SetupEchoContext(
			t,
			fmt.Sprintf(`{"title":%q}`, testhelpers.Course.Title),
			"course",
			true,
		)

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0, testhelpers.AddCourseHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return nil, stdErrors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		ctx, _ := testhelpers.SetupEchoContext(
			t,
			fmt.Sprintf(`{"title":%q,"description":%q}`, testhelpers.Course.Title, testhelpers.Course.Description),
			"course",
			true,
		)

		err := h.AddCourse(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Adding("course"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.AddCourseCalls()), 1, testhelpers.AddCourseHandlerName)
	})
}
