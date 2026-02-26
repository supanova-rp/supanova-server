package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/services/metrics"
)

func Metrics(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		latency := time.Since(start).Seconds() // prometheus measures latency in seconds

		status := c.Response().Status
		// echo writes response after middleware, so c.Response().Status defaults to 200; read from error instead
		if err != nil {
			var httpErr *echo.HTTPError
			if errors.As(err, &httpErr) {
				status = httpErr.Code
			} else {
				status = http.StatusInternalServerError
			}
		}

		statusText := http.StatusText(status)
		method := c.Request().Method
		path := c.Path()

		metrics.HTTPRequestDuration.WithLabelValues(method, path, statusText).Observe(latency)
		metrics.HTTPRequestsTotal.WithLabelValues(method, path).Inc()
		if err != nil {
			metrics.HTTPRequestsErrorsTotal.WithLabelValues(method, path, statusText).Inc()
		}

		return err
	}
}
