package middleware

import (
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
		path := c.Path()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request().Method, path, http.StatusText(status)).Observe(latency)

		return err
	}
}
