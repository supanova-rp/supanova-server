package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

//go:generate moq -out ../handlers/mocks/progress_mock.go -pkg mocks . ProgressRepository

type ProgressRepository interface {
	GetProgress(context.Context, sqlc.GetProgressParams) (*Progress, error)
	UpdateProgress(context.Context, sqlc.UpdateProgressParams) error
	HasCompletedCourse(context.Context, sqlc.HasCompletedCourseParams) (bool, error)
	SetCourseCompleted(context.Context, sqlc.SetCourseCompletedParams) error
}

type Progress struct {
	CompletedSectionIDs []uuid.UUID `json:"completedSectionIds"`
	CompletedIntro      bool        `json:"completedIntro"`
}
