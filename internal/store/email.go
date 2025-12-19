package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/supanova-rp/supanova-server/internal/services/email"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (s *Store) AddEmailFailure(ctx context.Context, params sqlc.AddEmailFailureParams) error {
	return s.Queries.AddEmailFailure(ctx, params)
}

func (s *Store) UpdateEmailFailure(ctx context.Context, params sqlc.UpdateEmailFailureParams) error {
	return s.Queries.UpdateEmailFailure(ctx, params)
}

func (s *Store) GetEmailFailures(ctx context.Context) ([]email.FailedEmail, error) {
	rows, err := s.Queries.GetEmailFailures(ctx)
	if err != nil {
		return nil, err
	}

	return utils.Map(rows, emailFailuresFrom), nil
}

func emailFailuresFrom(row sqlc.GetEmailFailuresRow) email.FailedEmail {
	return email.FailedEmail{
		ID:             utils.UUIDFrom(row.ID),
		TemplateName:   row.TemplateName,
		TemplateParams: row.TemplateParams,
		EmailName:      row.EmailName,
	}
}

func (s *Store) DeleteEmailFailures(ctx context.Context, ids []pgtype.UUID) error {
	return s.Queries.DeleteEmailFailures(ctx, ids)
}
