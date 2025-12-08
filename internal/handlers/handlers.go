package handlers

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/services/email"
)

type Handlers struct {
	System    domain.SystemRepository
	Course    domain.CourseRepository
	Progress  domain.ProgressRepository
	Enrolment domain.EnrolmentRepository

	ObjectStorage ObjectStorage
	EmailService  EmailService
}

//go:generate moq -out ../handlers/mocks/objectstorage_mock.go -pkg mocks . ObjectStorage

type ObjectStorage interface {
	GenerateUploadURL(ctx context.Context, key string, contentType *string) (string, error)
	GetCDNURL(ctx context.Context, key string) (string, error)
}

type EmailService interface {
	SendCourseCompletion(ctx echo.Context, params *email.CourseCompletionParams) error
}

func NewHandlers(
	system domain.SystemRepository,
	course domain.CourseRepository,
	progress domain.ProgressRepository,
	enrolment domain.EnrolmentRepository,
	objectStorage ObjectStorage,
	emailService EmailService,
) *Handlers {
	return &Handlers{
		System:        system,
		Course:        course,
		Progress:      progress,
		Enrolment:     enrolment,
		ObjectStorage: objectStorage,
		EmailService:  emailService,
	}
}
