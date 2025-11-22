package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func (s *Store) GetProgress(ctx context.Context, args sqlc.GetProgressByIdParams) (*domain.Progress, error) {
	progress, err := s.Queries.GetProgressById(ctx, args)
	if err != nil {
		return nil, err
	}

	var sectionUuids []uuid.UUID
	for _, sectionId := range progress.CompletedSectionIds {
		sectionUuids = append(sectionUuids, uuid.UUID(sectionId.Bytes))
	}

	return &domain.Progress{
		CompletedSectionIds: sectionUuids,
		CompletedIntro:      progress.CompletedIntro.Bool,
	}, nil
}
