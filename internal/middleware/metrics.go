package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

// histogram for request latency
var HTTPRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_latency_seconds",
		Help:    "Latency of HTTP requests in seconds",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"method", "path", "status"},
)

func RegisterMetrics() {
	prometheus.MustRegister(HTTPRequestDuration)
}

func Metrics(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		latency := time.Since(start).Seconds() // prometheus measures latency in seconds

		status := c.Response().Status
		path := c.Path()
		HTTPRequestDuration.WithLabelValues(c.Request().Method, path, http.StatusText(status)).Observe(latency)

		return err
	}
}
