package testhelpers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/supanova-rp/supanova-server/internal/config"
)

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func SetupEchoContext(t *testing.T, reqBody, endpoint string) echo.Context {
	t.Helper()

	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s/%s", config.APIVersion, endpoint), strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}