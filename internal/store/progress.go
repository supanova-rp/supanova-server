package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func (s *Store) GetProgress(ctx context.Context, args sqlc.GetProgressParams) (*domain.Progress, error) {
	progress, err := s.Queries.GetProgress(ctx, args)
	if err != nil {
		return nil, err
	}

	var sectionUUIDs []uuid.UUID
	for _, sectionID := range progress.CompletedSectionIds {
		sectionUUIDs = append(sectionUUIDs, uuid.UUID(sectionID.Bytes))
	}

	return &domain.Progress{
		CompletedSectionIDs: sectionUUIDs,
		CompletedIntro:      progress.CompletedIntro.Bool,
	}, nil
}

func (s *Store) UpdateProgress(ctx context.Context, args sqlc.UpdateProgressParams) error {
	return s.Queries.UpdateProgress(ctx, args)
}
