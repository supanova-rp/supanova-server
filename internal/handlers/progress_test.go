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
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func TestGetProgress(t *testing.T) {
	t.Run("returns progress successfully", func(t *testing.T) {
		expected := testhelpers.Progress

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params sqlc.GetProgressParams) (*domain.Progress, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{
			Progress: mockProgressRepo,
		}

		params := handlers.GetProgressParams{
			CourseID: uuid.New().String(),
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
		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params sqlc.GetProgressParams) (*domain.Progress, error) {
				return nil, nil
			},
		}

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
		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params sqlc.GetProgressParams) (*domain.Progress, error) {
				return nil, nil
			},
		}

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

	t.Run("progress not found", func(t *testing.T) {
		courseID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params sqlc.GetProgressParams) (*domain.Progress, error) {
				return nil, pgx.ErrNoRows
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

		testhelpers.AssertHTTPError(t, err, http.StatusNotFound, errors.NotFound("user progress"))
		testhelpers.AssertRepoCalls(t, len(mockRepo.GetProgressCalls()), 1, testhelpers.GetProgressHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		courseID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{
			GetProgressFunc: func(ctx context.Context, params sqlc.GetProgressParams) (*domain.Progress, error) {
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
		courseID := uuid.New()
		sectionID := uuid.New()

		mockProgressRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params sqlc.UpdateProgressParams) error {
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

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockProgressRepo.UpdateProgressCalls()), 1, testhelpers.UpdateProgressHandlerName)
	})

	t.Run("validation error - missing courseId", func(t *testing.T) {
		sectionID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params sqlc.UpdateProgressParams) error {
				return nil
			},
		}

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
		courseID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params sqlc.UpdateProgressParams) error {
				return nil
			},
		}

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

		mockRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params sqlc.UpdateProgressParams) error {
				return nil
			},
		}

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
		courseID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params sqlc.UpdateProgressParams) error {
				return nil
			},
		}

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
		courseID := uuid.New()
		sectionID := uuid.New()

		mockRepo := &mocks.ProgressRepositoryMock{
			UpdateProgressFunc: func(ctx context.Context, params sqlc.UpdateProgressParams) error {
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
