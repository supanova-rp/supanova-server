package domain

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type MediaRepository interface {
	GetVideoURL(context.Context, pgtype.UUID) (string, error)
	GetVideoUploadURL(context.Context, pgtype.UUID) (string, error)
}
