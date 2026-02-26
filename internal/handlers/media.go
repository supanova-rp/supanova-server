package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

const courseMaterialsResource = "course materials"

type VideoURLParams struct {
	CourseID   string `json:"courseId" validate:"required"`
	StorageKey string `json:"storageKey" validate:"required"`
}

func (h *Handlers) GetVideoUploadURL(e echo.Context) error {
	ctx := e.Request().Context()

	var params VideoURLParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	videoKey := getVideoKey(params)
	URL, err := h.ObjectStorage.GenerateUploadURL(ctx, videoKey, nil)
	if err != nil {
		return internalError(
			ctx,
			errors.Getting("upload url"),
			err,
			slog.String("course_id", params.CourseID), slog.String("storage_key", params.StorageKey),
		)
	}

	return e.JSON(http.StatusOK, &domain.VideoUploadURL{
		UploadURL: URL,
	})
}

func (h *Handlers) GetVideoURL(e echo.Context) error {
	ctx := e.Request().Context()

	var params VideoURLParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	videoKey := getVideoKey(params)
	URL, err := h.ObjectStorage.GetCDNURL(ctx, videoKey)
	if err != nil {
		return internalError(
			ctx,
			errors.Getting("video url"),
			err,
			slog.String("course_id", params.CourseID), slog.String("storage_key", params.StorageKey),
		)
	}

	return e.JSON(http.StatusOK, &domain.VideoURL{
		URL: URL,
	})
}

type GetCourseMaterialsParams struct {
	CourseID string `json:"courseId" validate:"required"`
}

func (h *Handlers) GetCourseMaterials(e echo.Context) error {
	ctx := e.Request().Context()

	var params GetCourseMaterialsParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	courseID, err := uuid.Parse(params.CourseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.InvalidUUID)
	}

	enrolled, err := h.isEnrolled(ctx, courseID)
	if err != nil {
		return internalError(ctx, errors.Getting(courseMaterialsResource), err, slog.String("course_id", params.CourseID))
	}
	if !enrolled {
		return echo.NewHTTPError(http.StatusForbidden, errors.Forbidden(courseMaterialsResource))
	}

	materials, err := h.Course.GetCourseMaterials(ctx, courseID)
	if err != nil {
		return internalError(ctx, errors.Getting(courseMaterialsResource), err, slog.String("course_id", params.CourseID))
	}

	materialsWithURL := make([]domain.CourseMaterialWithURL, 0, len(materials))
	for _, m := range materials {
		key := getMaterialKey(params.CourseID, m.StorageKey.String())
		url, err := h.ObjectStorage.GetCDNURL(ctx, key)
		if err != nil {
			return internalError(ctx, errors.Getting(courseMaterialsResource), err, slog.String("course_id", params.CourseID))
		}

		materialsWithURL = append(materialsWithURL, domain.CourseMaterialWithURL{
			ID:       m.ID,
			Name:     m.Name,
			Position: m.Position,
			URL:      url,
		})
	}

	return e.JSON(http.StatusOK, materialsWithURL)
}

func (h *Handlers) GetMaterialUploadURL(e echo.Context) error {
	ctx := e.Request().Context()

	var params VideoURLParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	materialKey := getMaterialKey(params.CourseID, params.StorageKey)
	contentType := "application/pdf"
	URL, err := h.ObjectStorage.GenerateUploadURL(ctx, materialKey, &contentType)
	if err != nil {
		return internalError(
			ctx,
			errors.Getting("upload url"),
			err,
			slog.String("course_id", params.CourseID), slog.String("storage_key", params.StorageKey),
		)
	}

	return e.JSON(http.StatusOK, &domain.CourseMaterialUploadURL{
		UploadURL: URL,
	})
}

func getVideoKey(params VideoURLParams) string {
	return fmt.Sprintf("%s/videos/%s", params.CourseID, params.StorageKey)
}

func getMaterialKey(courseID, storageKey string) string {
	return fmt.Sprintf("%s/materials/%s.pdf", courseID, storageKey)
}
