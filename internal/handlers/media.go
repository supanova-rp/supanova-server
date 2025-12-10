package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

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

func getVideoKey(params VideoURLParams) string {
	return fmt.Sprintf("%s/videos/%s", params.CourseID, params.StorageKey)
}
