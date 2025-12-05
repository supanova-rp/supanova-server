package server

import (
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers"
)

func RegisterCourseRoutes(e *echo.Echo, h *handlers.Handlers, apiVersion string) {
	e.POST(getRoute(apiVersion, "course"), h.GetCourse)
	e.POST(getRoute(apiVersion, "add-course"), h.AddCourse)
}

func RegisterProgressRoutes(e *echo.Echo, h *handlers.Handlers, apiVersion string) {
	e.POST(getRoute(apiVersion, "get-progress"), h.GetProgress)
	e.POST(getRoute(apiVersion, "update-progress"), h.UpdateProgress)
}

func RegisterMediaRoutes(e *echo.Echo, h *handlers.Handlers, apiVersion string) {
	e.POST(getRoute(apiVersion, "video-url"), h.GetVideoURL)
	e.POST(getRoute(apiVersion, "get-video-upload-url"), h.GetVideoUploadURL)
}
