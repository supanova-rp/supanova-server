package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/supanova-rp/supanova-server/internal/services/email"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (s *Store) AddFailedEmail(ctx context.Context, params sqlc.AddFailedEmailParams) error {
	return ExecCommand(ctx, func() error {
		return s.Queries.AddFailedEmail(ctx, params)
	})
}

func (s *Store) UpdateFailedEmail(ctx context.Context, params sqlc.UpdateFailedEmailParams) error {
	return ExecCommand(ctx, func() error {
		return s.Queries.UpdateFailedEmail(ctx, params)
	})
}

func (s *Store) GetFailedEmails(ctx context.Context) ([]email.FailedEmail, error) {
	rows, err := ExecQuery(ctx, func() ([]sqlc.GetFailedEmailsRow, error) {
		return s.Queries.GetFailedEmails(ctx)
	})
	if err != nil {
		return nil, err
	}

	return utils.Map(rows, func(row sqlc.GetFailedEmailsRow) email.FailedEmail {
		return failedEmailsFrom(&row)
	}), nil
}

func failedEmailsFrom(row *sqlc.GetFailedEmailsRow) email.FailedEmail {
	return email.FailedEmail{
		ID:             utils.UUIDFrom(row.ID),
		TemplateName:   row.TemplateName,
		TemplateParams: row.TemplateParams,
		EmailName:      row.EmailName,
		Retries:        int(row.Retries),
	}
}

func (s *Store) DeleteFailedEmail(ctx context.Context, emailID pgtype.UUID) error {
	return ExecCommand(ctx, func() error {
		return s.Queries.DeleteFailedEmail(ctx, emailID)
	})
}
