package server

import (
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers"
)

func RegisterCourseRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/course", h.GetCourse)
	private.POST("/add-course", h.AddCourse)
}

func RegisterProgressRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/get-progress", h.GetProgress)
	private.POST("/update-progress", h.UpdateProgress)
}

func RegisterMediaRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/video-url", h.GetVideoURL)
	private.POST("/get-video-upload-url", h.GetVideoUploadURL)
}
