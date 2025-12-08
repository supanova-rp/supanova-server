package handlers

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/services/email"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

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

		return internalError(ctx, errors.Getting(progressResource), err, slog.String("id", params.CourseID))
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
			slog.String("courseId", params.CourseID),
			slog.String("sectionId", params.SectionID))
	}

	return e.NoContent(http.StatusOK)
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

	completed, err := h.Progress.HasCompletedCourse(ctx, sqlc.HasCompletedCourseParams{
		UserID:   userID,
		CourseID: courseID,
	})

	if completed.CompletedCourse {
		return e.JSON(http.StatusOK, completed)
	}

	err = h.Progress.SetCourseCompleted(ctx, sqlc.SetCourseCompletedParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return internalError(ctx, errors.Updating(progressResource), err,
			slog.String("courseId", params.CourseID),
			slog.String("userId", userID))
	}

	user, err := h.User.GetUser(ctx, userID)

	// TODO: get timestamp

	emailParams := &email.CourseCompletionParams{
		UserName: user.Name,
		UserEmail: user.Email,
		CourseName: params.CourseName,
		CompletionTimestamp: "",

	}
	h.EmailService.SendCourseCompletion(ctx, emailParams)

	return e.JSON(http.StatusOK, completed)
}
