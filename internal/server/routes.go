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
}

func RegisterMediaRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/video-url", h.GetVideoURL)
	private.POST("/get-video-upload-url", h.GetVideoUploadURL)
}

func RegisterEnrollmentRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/update-users-to-courses", h.UpdateUserCourseEnrollment)
}
