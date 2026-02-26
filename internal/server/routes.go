package server

import (
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers"
)

func RegisterCourseRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/course", h.GetCourse)
	private.POST("/materials", h.GetCourseMaterials)

	// admin routes
	private.POST("/course-titles", h.GetCoursesOverview)
	private.POST("/add-course", h.AddCourse)
	private.POST("/delete-course", h.DeleteCourse)
}

func RegisterProgressRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/get-progress", h.GetProgress)
	private.POST("/update-progress", h.UpdateProgress)
	private.POST("/set-course-completed", h.SetCourseCompleted)

	// admin routes
	private.POST("/admin/get-all-progress", h.GetAllProgress)
	private.POST("/reset-progress", h.ResetProgress)
}

func RegisterQuizRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/quiz/save-attempt", h.SaveQuizAttempt)
	private.POST("/quiz/save-state", h.SaveQuizState)
	private.POST("/quiz/get-all-sections", h.GetAllQuizSections)

	// admin routes
	private.POST("/admin/quiz/get-attempts", h.GetQuizAttemptsByUserID)
	private.POST("/admin/quiz/reset-progress", h.ResetQuizProgress)
}

func RegisterMediaRoutes(private *echo.Group, h *handlers.Handlers) {
	private.POST("/video-url", h.GetVideoURL)

	// admin routes
	private.POST("/get-video-upload-url", h.GetVideoUploadURL)
	private.POST("/get-material-upload-url", h.GetMaterialUploadURL)
}

func RegisterEnrolmentRoutes(private *echo.Group, h *handlers.Handlers) {
	// admin routes
	private.POST("/update-users-to-courses", h.UpdateCourseEnrolment)
}
