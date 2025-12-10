package handlers_test

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
)

func TestGetVideoURL(t *testing.T) {
	t.Run("returns video URL successfully", func(t *testing.T) {
		expected := &domain.VideoURL{URL: "https://mycdnurl.com"}

		objectStorageMock := &mocks.ObjectStorageMock{
			GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
				return expected.URL, nil
			},
		}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			CourseID:   testhelpers.VideoURLParams.CourseID,
			StorageKey: testhelpers.VideoURLParams.StorageKey,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "video-url")

		err := h.GetVideoURL(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual domain.VideoURL
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("url mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(t, len(objectStorageMock.GetCDNURLCalls()), 1, testhelpers.GetVideoURLHandlerName)
	})

	t.Run("validation error - missing courseId", func(t *testing.T) {
		objectStorageMock := &mocks.ObjectStorageMock{}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			StorageKey: testhelpers.VideoURLParams.StorageKey,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "video-url")

		err := h.GetVideoURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(objectStorageMock.GetCDNURLCalls()), 0, testhelpers.GetVideoURLHandlerName)
	})

	t.Run("validation error - missing storageKey", func(t *testing.T) {
		objectStorageMock := &mocks.ObjectStorageMock{}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			CourseID: testhelpers.VideoURLParams.CourseID,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "video-url")

		err := h.GetVideoURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(t, len(objectStorageMock.GetCDNURLCalls()), 0, testhelpers.GetVideoURLHandlerName)
	})

	t.Run("internal server error", func(t *testing.T) {
		objectStorageMock := &mocks.ObjectStorageMock{
			GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
				return "", stdErrors.New("crypto/rsa: message too long for RSA key size")
			},
		}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			CourseID:   testhelpers.VideoURLParams.CourseID,
			StorageKey: testhelpers.VideoURLParams.StorageKey,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "video-url")

		err := h.GetVideoURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("video url"))
		testhelpers.AssertRepoCalls(t, len(objectStorageMock.GetCDNURLCalls()), 1, testhelpers.GetVideoURLHandlerName)
	})
}

func TestGetVideoUploadURL(t *testing.T) {
	t.Run("returns video upload URL successfully", func(t *testing.T) {
		expected := &domain.VideoUploadURL{UploadURL: "https://s3uploadurl.com"}

		objectStorageMock := &mocks.ObjectStorageMock{
			GenerateUploadURLFunc: func(ctx context.Context, key string, contentType *string) (string, error) {
				return expected.UploadURL, nil
			},
		}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			CourseID:   testhelpers.VideoURLParams.CourseID,
			StorageKey: testhelpers.VideoURLParams.StorageKey,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "get-video-upload-url")

		err := h.GetVideoUploadURL(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual domain.VideoUploadURL
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if diff := cmp.Diff(expected, &actual); diff != "" {
			t.Errorf("url mismatch (-want +got):\n%s", diff)
		}

		testhelpers.AssertRepoCalls(
			t,
			len(objectStorageMock.GenerateUploadURLCalls()),
			1,
			testhelpers.GetVideoUploadURLHandlerName,
		)
	})

	t.Run("validation error - missing courseId", func(t *testing.T) {
		objectStorageMock := &mocks.ObjectStorageMock{}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			StorageKey: testhelpers.VideoURLParams.StorageKey,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "get-video-upload-url")

		err := h.GetVideoUploadURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(
			t,
			len(objectStorageMock.GenerateUploadURLCalls()),
			0,
			testhelpers.GetVideoUploadURLHandlerName,
		)
	})

	t.Run("validation error - missing storageKey", func(t *testing.T) {
		objectStorageMock := &mocks.ObjectStorageMock{}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			CourseID: testhelpers.VideoURLParams.CourseID,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "get-video-upload-url")

		err := h.GetVideoUploadURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(
			t,
			len(objectStorageMock.GenerateUploadURLCalls()),
			0,
			testhelpers.GetVideoUploadURLHandlerName,
		)
	})

	t.Run("internal server error", func(t *testing.T) {
		objectStorageMock := &mocks.ObjectStorageMock{
			GenerateUploadURLFunc: func(ctx context.Context, key string, contentType *string) (string, error) {
				return "", stdErrors.New("InvalidBucketName: The specified bucket is not valid.")
			},
		}

		h := &handlers.Handlers{
			ObjectStorage: objectStorageMock,
		}

		reqBody := handlers.VideoURLParams{
			CourseID:   testhelpers.VideoURLParams.CourseID,
			StorageKey: testhelpers.VideoURLParams.StorageKey,
		}

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "get-video-upload-url")

		err := h.GetVideoUploadURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("upload url"))
		testhelpers.AssertRepoCalls(
			t,
			len(objectStorageMock.GenerateUploadURLCalls()),
			1,
			testhelpers.GetVideoUploadURLHandlerName,
		)
	})
}
