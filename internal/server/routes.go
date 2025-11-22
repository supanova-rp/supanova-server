package server

import (
	"github.com/labstack/echo/v4"
	"github.com/supanova-rp/supanova-server/internal/handlers"
)

func RegisterCourseRoutes(e *echo.Echo, h *handlers.Handlers) {
	e.GET(getRoute("v2", "course/:id"), h.GetCourse)
	e.POST(getRoute("v2", "course"), h.AddCourse)
}
