package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		req := c.Request()

		// Read request body if present
		var bodyBytes []byte
		if req.Body != nil {
			// No need to handle error, if it fails bodyBytes will be empty and then won't be logged.
			// A request shouldn't be failed over a logging issue.
			bodyBytes, _ = io.ReadAll(req.Body)

			// Restore the request body
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		allParams := getRequestParams(c)

		// Execute the handler
		err := next(c)

		latency := time.Since(start)

		attrs := []any{
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.String("params", allParams),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("ip", c.RealIP()),
		}

		if uid := c.Request().Context().Value(UserIDContextKey); uid != nil {
			if userID, ok := uid.(string); ok {
				attrs = append(attrs, slog.String("user_id", userID))
			}
		}

		if len(bodyBytes) > 0 {
			var bodyMap map[string]interface{}

			err := json.Unmarshal(bodyBytes, &bodyMap)
			if err == nil {
				delete(bodyMap, "access_token") // Redact the access_token from logs

				sanitised, err := json.Marshal(bodyMap)
				if err == nil {
					attrs = append(attrs, slog.String("body", string(sanitised)))
				}
			}
		}

		if err != nil {
			attrs = append(attrs, slog.Any("error", err))
			slog.ErrorContext(c.Request().Context(), "request", attrs...)
		} else {
			slog.DebugContext(c.Request().Context(), "request", attrs...)
		}

		return err
	}
}

func getRequestParams(c echo.Context) string {
	// Start with query parameters (e.g. ?someId=123)
	params := c.Request().URL.Query().Encode()

	// Add path parameters (e.g. /:someId)
	for i, name := range c.ParamNames() {
		if params != "" {
			params += ", "
		}
		params += fmt.Sprintf("%s=%s", name, c.ParamValues()[i])
	}

	return params
}
