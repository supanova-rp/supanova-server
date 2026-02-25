package testhelpers

import (
	"context"
	"encoding/json"
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
	"github.com/supanova-rp/supanova-server/internal/middleware"
)

const (
	AddCourseHandlerName            = "AddCourse"
	DeleteCourseHandlerName         = "DeleteCourse"
	GetCourseHandlerName            = "GetCourse"
	GetCourseMaterialsHandlerName   = "GetCourseMaterials"
	GetCoursesOverviewHandlerName   = "GetCoursesOverview"
	IsEnrolledHandlerName           = "IsEnrolled"
	GetVideoURLHandlerName          = "GetVideoURL"
	GetVideoUploadURLHandlerName    = "GetVideoUploadURL"
	UpdateCourseEnrolmentHandler    = "UpdateCourseEnrolment"
	EnrolUserInCourseHandlerName    = "EnrolInCourse"
	DisenrolUserInCourseHandlerName = "DisenrolInCourse"
	GetAllProgressHandlerName       = "GetAllProgress"
	GetProgressHandlerName          = "GetProgress"
	UpdateProgressHandlerName       = "UpdateProgress"
	SetCourseCompletedHandlerName   = "SetCourseCompleted"
	HasCompletedCourseHandlerName   = "HasCompletedCourse"
	GetUserHandlerName              = "GetUser"
	SendEmailHandlerName            = "SendEmail"
	GetTemplateNamesHandlerName     = "GetEmailTemplateNames"

	TestUserID = "test-user-id"
)

var User = &domain.User{
	ID:    uuid.New().String(),
	Name:  "User A",
	Email: "usera@gmail.com",
}

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

var Progress = &domain.Progress{
	CompletedSectionIDs: []uuid.UUID{uuid.New(), uuid.New()},
	CompletedIntro:      true,
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

type setupOptions struct {
	userID *string
	role   *config.Role
}

type EchoTestOption func(*setupOptions)

func WithUserID(id string) EchoTestOption {
	return func(o *setupOptions) {
		o.userID = &id
	}
}

func WithRole(role config.Role) EchoTestOption {
	return func(o *setupOptions) {
		o.role = &role
	}
}

func SetupEchoContext(t *testing.T, reqBody interface{}, endpoint string, opts ...EchoTestOption) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	adminRole := config.AdminRole
	defaultUserID := TestUserID

	o := &setupOptions{
		userID: &defaultUserID,
		role:   &adminRole,
	}

	for _, opt := range opts {
		opt(o)
	}

	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal reqBody: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s/%s", config.APIVersion, endpoint), strings.NewReader(string(jsonBytes)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := req.Context()

	ctx = context.WithValue(ctx, middleware.UserIDContextKey, *o.userID)
	ctx = context.WithValue(ctx, middleware.RoleContextKey, *o.role)

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
