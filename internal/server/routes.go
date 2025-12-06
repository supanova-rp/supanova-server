package server

import (
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers"
)

func RegisterCourseRoutes(private *echo.Group, h *handlers.Handlers, apiVersion string) {
	private.POST(getRoute(apiVersion, "course"), h.GetCourse)
	private.POST(getRoute(apiVersion, "add-course"), h.AddCourse)
}

func RegisterProgressRoutes(private *echo.Group, h *handlers.Handlers, apiVersion string) {
	private.POST(getRoute(apiVersion, "get-progress"), h.GetProgress)
}

func RegisterMediaRoutes(private *echo.Group, h *handlers.Handlers, apiVersion string) {
	private.POST(getRoute(apiVersion, "video-url"), h.GetVideoURL)
	private.POST(getRoute(apiVersion, "get-video-upload-url"), h.GetVideoUploadURL)
}
