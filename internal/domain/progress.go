package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type ProgressRespository interface {
	GetProgress(context.Context, sqlc.GetProgressByIdParams) (*Progress, error)
}

type Progress struct {
	CompletedSectionIds []uuid.UUID
	CompletedIntro bool
}