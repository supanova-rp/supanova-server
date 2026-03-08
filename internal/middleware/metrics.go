package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/services/metrics"
)

func Metrics(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiPrefix := "/" + config.APIVersion
		reqPath := c.Request().URL.Path

		// ignore anything without the apiPrefix, also ignore the /health endpoint
		if !strings.HasPrefix(reqPath, apiPrefix) || reqPath == apiPrefix+"/health" {
			return next(c)
		}

		start := time.Now()
		err := next(c)
		latency := time.Since(start).Seconds() // prometheus measures latency in seconds

		status := responseStatus(c, err)
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
