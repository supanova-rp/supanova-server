package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type ProgressRespository interface {
	GetProgress(context.Context, sqlc.GetProgressByIDParams) (*Progress, error)
}

type Progress struct {
	CompletedSectionIDs []uuid.UUID `json:"completedSectionIds"`
	CompletedIntro      bool        `json:"completedIntro"`
}
