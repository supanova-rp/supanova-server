package server

import (
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers"
)

func RegisterCourseRoutes(e *echo.Echo, h *handlers.Handlers) {
	e.POST(getRoute("v2", "course"), h.GetCourse)
	e.POST(getRoute("v2", "add-course"), h.AddCourse)
}

func RegisterProgressRoutes(e *echo.Echo, h *handlers.Handlers) {
	e.POST(getRoute("v2", "get-progress"), h.GetProgress)
}

func RegisterMediaRoutes(e *echo.Echo, h *handlers.Handlers) {
	e.POST(getRoute("v2", "video-url"), h.GetVideoURL)
}
