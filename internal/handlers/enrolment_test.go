package handlers_test

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
	"github.com/supanova-rp/supanova-server/internal/handlers/testhelpers"
)

func TestUpdateCourseEnrolment_HappyPath(t *testing.T) {
	t.Run("enrols user successfully when IsEnrolled is false", func(t *testing.T) {
		courseID := testhelpers.Course.ID.String()

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			EnrolInCourseFunc: func(ctx context.Context, params domain.EnrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{Enrolment: mockEnrolmentRepo}

		req := handlers.UpdateCourseEnrolmentParams{
			CourseID:   courseID,
			IsEnrolled: false,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "enrolment")
		err := h.UpdateCourseEnrolment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 1, testhelpers.EnrolUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 0, testhelpers.DisenrolUserInCourseHandlerName)
	})

	t.Run("disenrols user successfully when IsEnrolled is true", func(t *testing.T) {
		courseID := testhelpers.Course.ID.String()

		mockEnrolmentRepo := &mocks.EnrolmentRepositoryMock{
			DisenrolInCourseFunc: func(ctx context.Context, params domain.DisenrolInCourseParams) error {
				return nil
			},
		}

		h := &handlers.Handlers{Enrolment: mockEnrolmentRepo}

		req := handlers.UpdateCourseEnrolmentParams{
			CourseID:   courseID,
			IsEnrolled: true,
		}

		ctx, rec := testhelpers.SetupEchoContext(t, req, "enrolment")
		err := h.UpdateCourseEnrolment(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.DisenrolInCourseCalls()), 1, testhelpers.DisenrolUserInCourseHandlerName)
		testhelpers.AssertRepoCalls(t, len(mockEnrolmentRepo.EnrolInCourseCalls()), 0, testhelpers.EnrolUserInCourseHandlerName)
	})
}

func TestUpdateCourseEnrolment_UnhappyPath(t *testing.T) {
	type testCase struct {
		name           string
		reqBody        handlers.UpdateCourseEnrolmentParams
		setup          func() *handlers.Handlers
		wantStatus     int
		expectedErrMsg string
	}

	courseID := testhelpers.Course.ID.String()

	tests := []testCase{
		{
			name: "validation error - missing course id",
			reqBody: handlers.UpdateCourseEnrolmentParams{
				IsEnrolled: false,
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.Validation,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Enrolment: &mocks.EnrolmentRepositoryMock{}}
			},
		},
		{
			name: "validation error - invalid uuid format",
			reqBody: handlers.UpdateCourseEnrolmentParams{
				CourseID:   "invalid-uuid",
				IsEnrolled: false,
			},
			wantStatus:     http.StatusBadRequest,
			expectedErrMsg: errors.InvalidUUID,
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{Enrolment: &mocks.EnrolmentRepositoryMock{}}
			},
		},
		{
			name: "internal server error",
			reqBody: handlers.UpdateCourseEnrolmentParams{
				CourseID:   courseID,
				IsEnrolled: false,
			},
			wantStatus:     http.StatusInternalServerError,
			expectedErrMsg: errors.Creating("enrolment"),
			setup: func() *handlers.Handlers {
				return &handlers.Handlers{
					Enrolment: &mocks.EnrolmentRepositoryMock{
						EnrolInCourseFunc: func(ctx context.Context, params domain.EnrolInCourseParams) error {
							return stdErrors.New("db error")
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			ctx, _ := testhelpers.SetupEchoContext(t, tt.reqBody, "enrolment")
			err := h.UpdateCourseEnrolment(ctx)
			testhelpers.AssertHTTPError(t, err, tt.wantStatus, tt.expectedErrMsg)
		})
	}
}
