package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type ProgressRepository interface {
	GetProgress(context.Context, sqlc.GetProgressParams) (*Progress, error)
	UpdateProgress(context.Context, sqlc.UpdateProgressParams) error
}

type Progress struct {
	CompletedSectionIDs []uuid.UUID `json:"completedSectionIds"`
	CompletedIntro      bool        `json:"completedIntro"`
}
