package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	_ "time/tzdata" // embed timezone database in binary so time.LoadLocation works on ubuntu

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/services/email"
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
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params GetProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	progress, err := h.Progress.GetProgress(e.Request().Context(), domain.GetProgressParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		// If there is no progress on this course, return a default entry
		if errors.IsNotFoundErr(err) {
			return e.JSON(http.StatusOK, &domain.Progress{
				CompletedIntro:      false,
				CompletedSectionIDs: []uuid.UUID{},
			})
		}

		return httpError(http.StatusInternalServerError, errors.Getting(progressResource), err)
	}

	return e.JSON(http.StatusOK, progress)
}

func (h *Handlers) UpdateProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params UpdateProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	sectionID, err := uuid.Parse(params.SectionID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	err = h.Progress.UpdateProgress(ctx, domain.UpdateProgressParams{
		UserID:    userID,
		CourseID:  courseID,
		SectionID: sectionID,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Updating(progressResource), err)
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) SetCourseCompleted(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params SetCourseCompletedParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	prevCompleted, err := h.Progress.HasCompletedCourse(ctx, domain.HasCompletedCourseParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Updating(progressResource), err)
	}

	if prevCompleted {
		return e.NoContent(http.StatusNoContent)
	}

	err = h.Progress.SetCourseCompleted(ctx, domain.SetCourseCompletedParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Updating(progressResource), err)
	}

	user, err := h.User.GetUser(ctx, userID)
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Updating(progressResource), err)
	}

	emailParams := &email.CourseCompletionParams{
		UserName:            user.Name,
		UserEmail:           user.Email,
		CourseName:          params.CourseName,
		CompletionTimestamp: time.Now().In(location).Format("02/01/2006 15:04:05"),
	}

	go func() {
		emailName := h.EmailService.GetEmailNames().CourseCompletion
		templateName := h.EmailService.GetTemplateNames().CourseCompletion

		err = h.EmailService.Send(
			context.WithoutCancel(ctx),
			emailParams,
			templateName,
			emailName,
		)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to send email",
				slog.Any("error", err),
				slog.String("email_name", emailName),
				slog.String("template_name", templateName),
				slog.String("course_id", params.CourseID),
				slog.String("user_id", userID),
			)
		} else {
			slog.InfoContext(
				ctx,
				"course completion email sent",
				slog.String("email_name", emailName),
				slog.String("template_name", templateName),
				slog.String("course_id", params.CourseID),
				slog.String("user_id", userID),
			)
		}
	}()

	return e.NoContent(http.StatusNoContent)
}

type SetIntroCompletedParams struct {
	CourseID string `json:"courseId" validate:"required"`
}

type SetIntroCompletedResponse struct {
	CourseID       string `json:"courseId"`
	CompletedIntro bool   `json:"completed_intro"`
}

func (h *Handlers) SetIntroCompleted(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params SetIntroCompletedParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	err = h.Progress.SetIntroCompleted(ctx, domain.SetIntroCompletedParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Updating(progressResource), err)
	}

	return e.JSON(http.StatusOK, SetIntroCompletedResponse{
		CourseID:       params.CourseID,
		CompletedIntro: true,
	})
}

type ResetProgressParams struct {
	CourseID string `json:"courseId" validate:"required"`
}

func (h *Handlers) ResetProgress(e echo.Context) error {
	ctx := e.Request().Context()

	userID, ok := getUserID(ctx)
	if !ok {
		return httpError(http.StatusInternalServerError, errors.NotFoundInCtx("user"), nil)
	}

	var params ResetProgressParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return httpError(http.StatusBadRequest, errors.InvalidUUID, err)
	}

	err = h.Progress.ResetProgress(ctx, domain.ResetProgressParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return httpError(http.StatusInternalServerError, errors.Updating(progressResource), err)
	}

	return e.NoContent(http.StatusNoContent)
}

func (h *Handlers) GetAllProgress(e echo.Context) error {
	ctx := e.Request().Context()

	progress, err := h.Progress.GetAllProgress(ctx)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return httpError(http.StatusNotFound, errors.NotFound(progressResource), err)
		}

		return httpError(http.StatusInternalServerError, errors.Getting(progressResource), err)
	}

	return e.JSON(http.StatusOK, progress)
}
