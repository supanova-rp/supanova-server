package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
)

type VideoURL struct {
	URL string `json:"url"`
}

// TODO: unit tests for GetVideoUploadURL
func TestGetVideoURL(t *testing.T) {
	t.Run("returns video URL successfully", func(t *testing.T) {
		expected := &VideoURL{URL: "https://mycdnurl.com"}

		objectStorageMock := &mocks.ObjectStorageMock{
			GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
				return expected.URL, nil
			},
		}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := fmt.Sprintf(
			`{"courseId":%q,"storageKey":%q}`,
			testhelpers.VideoURLParams.CourseID,
			testhelpers.VideoURLParams.StorageKey,
		)
		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "video-url", false)

		err := h.GetVideoURL(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual VideoURL
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("url mismatch (-want +got):\n%s", diff)
		}
	})

	// TODO: 404 not found case
	// TODO: 400 courseId missing
	// TODO: 400 storageKey missing
	// TODO: 500 internal server error
}
