package middleware

import (
	"context"
	"os"

	"github.com/labstack/echo/v4"
)

type ContextKey string

var UserIDContextKey ContextKey = "userID"

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if testUserID := os.Getenv("TEST_ENVIRONMENT_USER_ID"); testUserID != "" {
			ctx := context.WithValue(c.Request().Context(), UserIDContextKey, testUserID)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}

		// TODO: implement auth middleware using firebase
		return next(c)
	}
}
