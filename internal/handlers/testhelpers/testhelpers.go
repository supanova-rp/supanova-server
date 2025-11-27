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

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/middleware"
)

const testUserID = "test-user-id"

var Course = &domain.Course{
	ID:          uuid.New(),
	Title:       "Test Course",
	Description: "Test Description",
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
		ctx = context.WithValue(req.Context(), middleware.UserIDContextKey, testUserID)
	} else {
		ctx = req.Context()
	}

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}
