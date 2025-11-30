package testhelpers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
)

const (
	testUserID                   = "test-user-id"
	AddCourseHandlerName         = "AddCourse"
	GetCourseHandlerName         = "GetCourse"
	GetVideoURLHandlerName       = "GetVideoURL"
	GetVideoUploadURLHandlerName = "GetVideoUploadURL"
)

var Course = &domain.Course{
	ID:                uuid.New(),
	Title:             "Test Course",
	Description:       "Test Description",
	CompletionTitle:   "Completion Title",
	CompletionMessage: "Completion Message",
	Sections:          []domain.CourseSection{},
	Materials:         []domain.CourseMaterial{},
}

var VideoURLParams = &handlers.VideoURLParams{
	CourseID:   uuid.New().String(),
	StorageKey: uuid.New().String(),
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func SetupEchoContext(t *testing.T, reqBody, endpoint string, withContextValue bool) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/v2/%s", endpoint), strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	var ctx context.Context
	if withContextValue {
		ctx = context.WithValue(req.Context(), config.UserIDContextKey, testUserID)
	} else {
		ctx = req.Context()
	}

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func AssertHTTPError(t *testing.T, err error, expectedCode int, expectedMsg string) {
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

func AssertRepoCalls(t *testing.T, got, expected int, handlerName string) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %d calls to %s, got %d", expected, handlerName, got)
	}
}
