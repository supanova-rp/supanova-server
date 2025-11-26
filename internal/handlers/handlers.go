package handlers

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/domain"
)

type Handlers struct {
	System   domain.SystemRepository
	Course   domain.CourseRepository
	Progress domain.ProgressRepository

	ObjectStorage ObjectStorage
}

type ObjectStorage interface {
	GenerateUploadURL(ctx context.Context, key string, contentType *string) (string, error)
	GetCDNURL(ctx context.Context, key string) (string, error)
}

func NewHandlers(
	system domain.SystemRepository,
	course domain.CourseRepository,
	progress domain.ProgressRepository,
	objectStorage ObjectStorage,
) *Handlers {
	return &Handlers{
		System:        system,
		Course:        course,
		Progress:      progress,
		ObjectStorage: objectStorage,
	}
}
