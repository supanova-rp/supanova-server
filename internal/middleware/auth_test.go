package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/middleware"
	"github.com/supanova-rp/supanova-server/internal/middleware/mocks"
	"github.com/supanova-rp/supanova-server/internal/middleware/testhelpers"
	"github.com/supanova-rp/supanova-server/internal/services/auth"
	"github.com/supanova-rp/supanova-server/internal/tests"
)

const accessToken = "test-access-token"

func TestMiddleware(t *testing.T) {
	t.Run("authorised: user is admin", func(t *testing.T) {
		mockAuthProvider := &mocks.AuthProviderMock{
			GetUserFromIDTokenFunc: func(ctx context.Context, token string) (*auth.User, error) {
				return &auth.User{
					ID:      tests.TestUserID,
					IsAdmin: true,
				}, nil
			},
		}

		reqBody := map[string]interface{}{
			"id":           uuid.New().String(),
			"access_token": accessToken,
		}

		c := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := middleware.AuthMiddleware(nextMock, mockAuthProvider)(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		role, ok := c.Request().Context().Value(middleware.RoleContextKey).(config.Role)
		if !ok || role != config.AdminRole {
			t.Fatalf("expected admin role in context, got %s", role)
		}

		userID, ok := c.Request().Context().Value(middleware.UserIDContextKey).(string)
		if !ok || userID != tests.TestUserID {
			t.Fatalf("expected userID in context, got %s", userID)
		}
	})

	t.Run("authorised: user is non admin", func(t *testing.T) {
		mockAuthProvider := &mocks.AuthProviderMock{
			GetUserFromIDTokenFunc: func(ctx context.Context, token string) (*auth.User, error) {
				return &auth.User{
					ID:      tests.TestUserID,
					IsAdmin: false,
				}, nil
			},
		}

		reqBody := map[string]interface{}{
			"id":           uuid.New().String(),
			"access_token": accessToken,
		}

		c := testhelpers.SetupEchoContext(t, reqBody, "course")

		err := middleware.AuthMiddleware(nextMock, mockAuthProvider)(c)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		role, ok := c.Request().Context().Value(middleware.RoleContextKey).(config.Role)
		if !ok || role != config.UserRole {
			t.Fatalf("expected user role in context, got %s", role)
		}

		userID, ok := c.Request().Context().Value(middleware.UserIDContextKey).(string)
		if !ok || userID != tests.TestUserID {
			t.Fatalf("expected userID in context, got %s", userID)
		}
	})

	t.Run("unauthorised: user is non admin on admin route", func(t *testing.T) {
		mockAuthProvider := &mocks.AuthProviderMock{
			GetUserFromIDTokenFunc: func(ctx context.Context, token string) (*auth.User, error) {
				return &auth.User{
					ID:      tests.TestUserID,
					IsAdmin: false,
				}, nil
			},
		}

		reqBody := map[string]interface{}{
			"id":           uuid.New().String(),
			"access_token": accessToken,
		}

		c := testhelpers.SetupEchoContext(t, reqBody, "add-course")

		err := middleware.AuthMiddleware(nextMock, mockAuthProvider)(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		httpErr, ok := err.(*echo.HTTPError)
		if !ok || httpErr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, httpErr.Code)
		}
	})

	t.Run("unauthorised: access_token is invalid", func(t *testing.T) {
		mockAuthProvider := &mocks.AuthProviderMock{
			GetUserFromIDTokenFunc: func(ctx context.Context, token string) (*auth.User, error) {
				return nil, errors.New("access_token is invalid")
			},
		}

		reqBody := map[string]interface{}{
			"id":           uuid.New().String(),
			"access_token": accessToken,
		}

		c := testhelpers.SetupEchoContext(t, reqBody, "add-course")

		err := middleware.AuthMiddleware(nextMock, mockAuthProvider)(c)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		httpErr, ok := err.(*echo.HTTPError)
		if !ok || httpErr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, httpErr.Code)
		}
	})
}

func nextMock(c echo.Context) error {
	return nil
}
