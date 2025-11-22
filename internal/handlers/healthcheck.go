package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) HealthCheck(e echo.Context) error {
	dbStatus := "ok"
	httpStatus := http.StatusOK
	err := h.System.PingDB(e.Request().Context())
	if err != nil {
		httpStatus = http.StatusServiceUnavailable
		dbStatus = "unreachable"
	}

	return e.JSON(httpStatus, map[string]string{
		"status": "ok",
		"db":     dbStatus,
	})
}
