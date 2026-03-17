package handlers

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/services/auth"
	"github.com/supanova-rp/supanova-server/internal/services/email"
)

type Handlers struct {
	System    domain.SystemRepository
	Course    domain.CourseRepository
	Progress  domain.ProgressRepository
	Enrolment domain.EnrolmentRepository
	User      domain.UserRepository
	Auth      domain.AuthRepository
	Quiz      domain.QuizRepository

	ObjectStorage ObjectStorage
	EmailService  EmailService
	AuthProvider  auth.AuthProvider
}

//go:generate moq -out ../handlers/mocks/objectstorage_mock.go -pkg mocks . ObjectStorage

type ObjectStorage interface {
	GenerateUploadURL(ctx context.Context, key string, contentType *string) (string, error)
	GetCDNURL(ctx context.Context, key string) (string, error)
}

//go:generate moq -out ../handlers/mocks/emailservice_mock.go -pkg mocks . EmailService

type EmailService interface {
	Send(ctx context.Context, params email.EmailParams, templateName, emailName string) error
	SetupRetry() (context.CancelFunc, error)
	GetTemplateNames() *email.TemplateNames
	GetEmailNames() *email.EmailNames
	StopRetry(ctx context.Context)
}

func NewHandlers(
	system domain.SystemRepository,
	course domain.CourseRepository,
	progress domain.ProgressRepository,
	enrolment domain.EnrolmentRepository,
	user domain.UserRepository,
	authentication domain.AuthRepository,
	quiz domain.QuizRepository,
	objectStorage ObjectStorage,
	emailService EmailService,
	authProvider auth.AuthProvider,
) *Handlers {
	return &Handlers{
		System:        system,
		Course:        course,
		Progress:      progress,
		Enrolment:     enrolment,
		User:          user,
		Auth:          authentication,
		Quiz:          quiz,
		ObjectStorage: objectStorage,
		EmailService:  emailService,
		AuthProvider:  authProvider,
	}
}
