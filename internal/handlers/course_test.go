package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func TestGetCourse(t *testing.T) {
	t.Run("returns course successfully", func(t *testing.T) {
		courseID := uuid.New()
		expected := &domain.Course{
			ID:          courseID,
			Title:       "Test Course",
			Description: "Test Description",
		}

		mockRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		c, rec := setupEchoContext(t, `{"id":"`+courseID.String()+`"}`)

		err := h.GetCourse(c)
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

		c, _ := setupEchoContext(t, `{}`) // missing id

		err := h.GetCourse(c)

		assertHTTPError(t, err, http.StatusBadRequest, "validation failed")
		assertRepoCalls(t, len(mockRepo.GetCourseCalls()), 0)
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

		c, _ := setupEchoContext(t, `{"id":"invalid-uuid"}`)

		err := h.GetCourse(c)

		assertHTTPError(t, err, http.StatusBadRequest, "invalid uuid format")
		assertRepoCalls(t, len(mockRepo.GetCourseCalls()), 0)
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

		c, _ := setupEchoContext(t, `{"id":"`+courseID.String()+`"}`)

		err := h.GetCourse(c)

		assertHTTPError(t, err, http.StatusNotFound, "course not found")
		assertRepoCalls(t, len(mockRepo.GetCourseCalls()), 1)
	})

	t.Run("internal server error", func(t *testing.T) {
		courseID := uuid.New()

		mockRepo := &mocks.CourseRepositoryMock{
			GetCourseFunc: func(ctx context.Context, id pgtype.UUID) (*domain.Course, error) {
				return nil, errors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{
			Course: mockRepo,
		}

		c, _ := setupEchoContext(t, `{"id":"`+courseID.String()+`"}`)

		err := h.GetCourse(c)

		assertHTTPError(t, err, http.StatusInternalServerError, "Error getting course")
		assertRepoCalls(t, len(mockRepo.GetCourseCalls()), 1)
	})
}

func TestAddCourse(t *testing.T) {
	t.Run("adds course successfully", func(t *testing.T) {
		courseID := uuid.New()
		expected := &domain.Course{
			ID:          courseID,
			Title:       "New Course",
			Description: "New Description",
		}

		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return expected, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		c, rec := setupEchoContext(t, `{"title":"New Course","description":"New Description"}`)

		err := h.AddCourse(c)
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

		if len(mockRepo.AddCourseCalls()) != 1 {
			t.Errorf("expected 1 call to AddCourse, got %d", len(mockRepo.AddCourseCalls()))
		}
	})

	t.Run("validation error - missing title", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return nil, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		c, _ := setupEchoContext(t, `{"description":"Test Description"}`)

		err := h.AddCourse(c)

		assertHTTPError(t, err, http.StatusBadRequest, "validation failed")
		assertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0)
	})

	t.Run("validation error - missing description", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return nil, nil
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		c, _ := setupEchoContext(t, `{"title":"Test Title"}`)

		err := h.AddCourse(c)

		assertHTTPError(t, err, http.StatusBadRequest, "validation failed")
		assertRepoCalls(t, len(mockRepo.AddCourseCalls()), 0)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mocks.CourseRepositoryMock{
			AddCourseFunc: func(ctx context.Context, params sqlc.AddCourseParams) (*domain.Course, error) {
				return nil, errors.New("database connection failed")
			},
		}

		h := &handlers.Handlers{Course: mockRepo}
		c, _ := setupEchoContext(t, `{"title":"Test Title","description":"Test Description"}`)

		err := h.AddCourse(c)

		assertHTTPError(t, err, http.StatusInternalServerError, "Error getting course")
		if len(mockRepo.AddCourseCalls()) != 1 {
			t.Errorf("expected 1 call to AddCourse, got %d", len(mockRepo.AddCourseCalls()))
		}
	})
}

func setupEchoContext(t *testing.T, reqBody string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}

	req := httptest.NewRequest(http.MethodPost, "/v2/course", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func assertHTTPError(t *testing.T, err error, expectedCode int, expectedMsg string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}

	if httpErr.Code != expectedCode {
		t.Errorf("expected status %d, got %d", expectedCode, httpErr.Code)
	}

	if httpErr.Message != expectedMsg {
		t.Errorf("expected message %q, got %v", expectedMsg, httpErr.Message)
	}
}

func assertRepoCalls(t *testing.T, got, expected int) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %d calls to GetCourse, got %d", expected, got)
	}
}
