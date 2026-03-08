package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// responseStatus returns the HTTP status code for a completed request.
// echo writes the response after middleware runs, so c.Response().Status defaults to 200;
// when an error is present the status must be read from the error instead.
func responseStatus(c echo.Context, err error) int {
	if err != nil {
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			return httpErr.Code
		}
		return http.StatusInternalServerError
	}
	return c.Response().Status
}
