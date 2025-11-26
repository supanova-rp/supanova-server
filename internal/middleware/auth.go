package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
)

type ContextKey string

const UserIDContextKey ContextKey = "userID"

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: implement auth middleware using firebase
		return next(c)
	}
}

func TestAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Request().Header.Get("X-Test-User-ID")

		ctx := context.WithValue(c.Request().Context(), UserIDContextKey, userID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
