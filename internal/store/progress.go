package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

func (s *Store) GetProgress(ctx context.Context, args sqlc.GetProgressParams) (*domain.Progress, error) {
	progress, err := ExecQuery(
		ctx,
		func() (sqlc.GetProgressRow, error) { return s.Queries.GetProgress(ctx, args) },
	)
	if err != nil {
		return nil, err
	}

	return progressFrom(progress), nil
}

func (s *Store) UpdateProgress(ctx context.Context, args sqlc.UpdateProgressParams) error {
	return ExecCommand(ctx, func() error {
		return s.Queries.UpdateProgress(ctx, args)
	})
}

func (s *Store) HasCompletedCourse(ctx context.Context, args sqlc.HasCompletedCourseParams) (bool, error) {
	completed, err := ExecQuery(ctx, func() (pgtype.Bool, error) {
		return s.Queries.HasCompletedCourse(ctx, args)
	})

	return completed.Bool, err
}

func (s *Store) SetCourseCompleted(ctx context.Context, args sqlc.SetCourseCompletedParams) error {
	return ExecCommand(ctx, func() error {
		return s.Queries.SetCourseCompleted(ctx, args)
	})
}

func progressFrom(row sqlc.GetProgressRow) *domain.Progress {
	var sectionUUIDs []uuid.UUID
	for _, sectionID := range row.CompletedSectionIds {
		sectionUUIDs = append(sectionUUIDs, uuid.UUID(sectionID.Bytes))
	}

	return &domain.Progress{
		CompletedSectionIDs: sectionUUIDs,
		CompletedIntro:      row.CompletedIntro.Bool,
	}
}
