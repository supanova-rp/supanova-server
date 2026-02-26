package handlers_test

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/config"
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

func TestGetMaterialUploadURL(t *testing.T) {
	t.Run("returns material upload URL successfully", func(t *testing.T) {
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

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "get-material-upload-url")

		err := h.GetMaterialUploadURL(ctx)
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
			testhelpers.GetMaterialUploadURLHandlerName,
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

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "get-material-upload-url")

		err := h.GetMaterialUploadURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(
			t,
			len(objectStorageMock.GenerateUploadURLCalls()),
			0,
			testhelpers.GetMaterialUploadURLHandlerName,
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

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "get-material-upload-url")

		err := h.GetMaterialUploadURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusBadRequest, errors.Validation)
		testhelpers.AssertRepoCalls(
			t,
			len(objectStorageMock.GenerateUploadURLCalls()),
			0,
			testhelpers.GetMaterialUploadURLHandlerName,
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

		ctx, _ := testhelpers.SetupEchoContext(t, reqBody, "get-material-upload-url")

		err := h.GetMaterialUploadURL(ctx)

		testhelpers.AssertHTTPError(t, err, http.StatusInternalServerError, errors.Getting("upload url"))
		testhelpers.AssertRepoCalls(
			t,
			len(objectStorageMock.GenerateUploadURLCalls()),
			1,
			testhelpers.GetMaterialUploadURLHandlerName,
		)
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

func TestGetCourseMaterials_HappyPath(t *testing.T) {
	courseID := testhelpers.Course.ID

	material1 := domain.CourseMaterial{
		ID:         uuid.New(),
		Name:       "Material 1",
		Position:   0,
		StorageKey: uuid.New(),
	}
	material2 := domain.CourseMaterial{
		ID:         uuid.New(),
		Name:       "Material 2",
		Position:   1,
		StorageKey: uuid.New(),
	}

	t.Run("returns materials with urls successfully", func(t *testing.T) {
		mockCourseRepo := &mocks.CourseRepositoryMock{
			GetCourseMaterialsFunc: func(ctx context.Context, id uuid.UUID) ([]domain.CourseMaterial, error) {
				return []domain.CourseMaterial{material1, material2}, nil
			},
		}

		mockObjectStorage := &mocks.ObjectStorageMock{
			GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
				return "https://cdn.example.com/" + key, nil
			},
		}

		h := &handlers.Handlers{
			Course:        mockCourseRepo,
			ObjectStorage: mockObjectStorage,
		}

		reqBody := handlers.GetCourseMaterialsParams{
			CourseID: courseID.String(),
		}

		ctx, rec := testhelpers.SetupEchoContext(t, reqBody, "materials")

		err := h.GetCourseMaterials(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var actual []domain.CourseMaterialWithURL
		if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(actual) != 2 {
			t.Fatalf("expected 2 materials, got %d", len(actual))
		}

		if actual[0].ID != material1.ID || actual[0].Name != material1.Name || actual[0].Position != material1.Position {
			t.Errorf("unexpected first material: %+v", actual[0])
		}
		if actual[0].URL == "" {
			t.Error("expected non-empty URL for material 1")
		}

		testhelpers.AssertRepoCalls(t, len(mockCourseRepo.GetCourseMaterialsCalls()), 1, testhelpers.GetCourseMaterialsHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockObjectStorage.GetCDNURLCalls()), 2, "GetCDNURL")
	})
}

func TestGetCourseMaterials_UnhappyPath(t *testing.T) {
	courseID := testhelpers.Course.ID
	userRole := config.UserRole

	material1 := domain.CourseMaterial{
		ID:         uuid.New(),
		Name:       "Material 1",
		Position:   0,
		StorageKey: uuid.New(),
	}

	type testCase struct {
		name           string
		reqBody        handlers.GetCourseMaterialsParams
		userRole       *config.Role
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name:           "validation error - missing course id",
			reqBody:        handlers.GetCourseMaterialsParams{},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name:           "validation error - invalid uuid format",
			reqBody:        handlers.GetCourseMaterialsParams{CourseID: "invalid-uuid"},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Course: &mocks.CourseRepositoryMock{}}
			},
		},
		{
			name:           "forbidden - user not enrolled",
			reqBody:        handlers.GetCourseMaterialsParams{CourseID: courseID.String()},
			userRole:       &userRole,
			wantStatus:     http.StatusForbidden,
			expectedErrMsg: errors.Forbidden("course materials"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Course: &mocks.CourseRepositoryMock{},
					Enrolment: &mocks.EnrolmentRepositoryMock{
						IsEnrolledFunc: func(ctx context.Context, params domain.IsEnrolledParams) (bool, error) {
							return false, nil
						},
					},
				}
			},
		},
		{
			name:           "internal server error from repo",
			reqBody:        handlers.GetCourseMaterialsParams{CourseID: courseID.String()},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("course materials"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Course: &mocks.CourseRepositoryMock{
						GetCourseMaterialsFunc: func(ctx context.Context, id uuid.UUID) ([]domain.CourseMaterial, error) {
							return nil, stdErrors.New("database connection failed")
						},
					},
				}
			},
		},
		{
			name:           "internal server error from object storage",
			reqBody:        handlers.GetCourseMaterialsParams{CourseID: courseID.String()},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Getting("course materials"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Course: &mocks.CourseRepositoryMock{
						GetCourseMaterialsFunc: func(ctx context.Context, id uuid.UUID) ([]domain.CourseMaterial, error) {
							return []domain.CourseMaterial{material1}, nil
						},
					},
					ObjectStorage: &mocks.ObjectStorageMock{
						GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
							return "", stdErrors.New("cdn error")
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()

			var opts []testhelpers.EchoTestOption
			if tt.userRole != nil {
				opts = append(opts, testhelpers.WithRole(*tt.userRole))
			}

			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "materials", opts...)
			err := h.GetCourseMaterials(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}
