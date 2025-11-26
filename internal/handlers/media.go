package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
)

type GetVideoURLParams struct {
	CourseID   string `json:"courseId" validate:"required"`
	StorageKey string `json:"storageKey" validate:"required"`
}

const videoResource = "video"

func (h *Handlers) GetVideoURL(e echo.Context) error {
	ctx := e.Request().Context()

	var params GetVideoURLParams
	if err := bindAndValidate(e, &params); err != nil {
		return err
	}

	videoKey := fmt.Sprintf("%s/videos/%s", params.CourseID, params.StorageKey)
	URL, err := h.ObjectStorage.GetCDNURL(ctx, videoKey)
	if err != nil {
		internalError(ctx, errors.Getting(videoResource), err, slog.String("id", params.CourseID), slog.String("storageKey", params.StorageKey))
	}

	return e.JSON(http.StatusOK, map[string]string{
		"url": URL,
	})
}
