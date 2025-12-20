package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/services/email"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

var location, _ = time.LoadLocation("Europe/London")

const progressResource = "user progress"

type GetProgressParams struct {
	CourseID string `json:"courseId" validate:"required"`
}

type UpdateProgressParams struct {
	CourseID  string `json:"courseId" validate:"required"`
	SectionID string `json:"sectionId" validate:"required"`
}

type SetCourseCompletedParams struct {
	CourseID   string `json:"courseId" validate:"required"`
	CourseName string `json:"courseName" validate:"required"`
}

func (h *Handlers) GetProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params GetProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := utils.PGUUIDFrom(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sqlcParams := sqlc.GetProgressParams{
		UserID:   userID,
		CourseID: courseID,
	}

	progress, err := h.Progress.GetProgress(e.Request().Context(), sqlcParams)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return echo.NewHTTPError(http.StatusNotFound, errors.NotFound(progressResource))
		}

		return internalError(ctx, errors.Getting(progressResource), err, slog.String("course_id", params.CourseID))
	}

	return e.JSON(http.StatusOK, progress)
}

func (h *Handlers) UpdateProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params UpdateProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := utils.PGUUIDFrom(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sectionID, err := utils.PGUUIDFrom(params.SectionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	sqlcParams := sqlc.UpdateProgressParams{
		UserID:    userID,
		CourseID:  courseID,
		SectionID: sectionID,
	}

	err = h.Progress.UpdateProgress(ctx, sqlcParams)
	if err != nil {
		return internalError(ctx, errors.Updating(progressResource), err,
			slog.String("course_id", params.CourseID),
			slog.String("section_id", params.SectionID))
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) SetCourseCompleted(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.NotFoundInCtx("user"))
	}

	var params SetCourseCompletedParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := utils.PGUUIDFrom(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	prevCompleted, err := h.Progress.HasCompletedCourse(ctx, sqlc.HasCompletedCourseParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return internalError(ctx, errors.Updating(progressResource), err,
			slog.String("course_id", params.CourseID),
			slog.String("user_id", userID))
	}

	if prevCompleted {
		return e.NoContent(http.StatusNoContent)
	}

	err = h.Progress.SetCourseCompleted(ctx, sqlc.SetCourseCompletedParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return internalError(ctx, errors.Updating(progressResource), err,
			slog.String("course_id", params.CourseID),
			slog.String("user_id", userID))
	}

	user, err := h.User.GetUser(ctx, userID)
	if err != nil {
		return internalError(ctx, errors.Updating(progressResource), err,
			slog.String("course_id", params.CourseID),
			slog.String("user_id", userID))
	}

	emailParams := &email.CourseCompletionParams{
		UserName:            user.Name,
		UserEmail:           user.Email,
		CourseName:          params.CourseName,
		CompletionTimestamp: time.Now().In(location).Format("02/01/2006 15:04:05"),
	}

	go func() {
		err = h.EmailService.Send(
			ctx,
			emailParams,
			h.EmailService.GetTemplateNames().CourseCompletion,
			h.EmailService.GetEmailNames().CourseCompletion,
		)
		if err != nil {
			slog.ErrorContext(ctx, err.Error(), slog.String("course_id", params.CourseID), slog.String("user_id", userID))
		}
	}()

	return e.NoContent(http.StatusNoContent)
}
