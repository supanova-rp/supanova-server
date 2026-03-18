package store

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

func (s *Store) RegisterUser(ctx context.Context, params domain.RegisterParams) (*domain.User, error) {
	id, err := ExecQuery(ctx, func() (string, error) {
		return s.Queries.InsertUser(ctx, sqlc.InsertUserParams{
			ID:    params.ID,
			Name:  utils.PGTextFrom(params.Name),
			Email: utils.PGTextFrom(params.Email),
		})
	})
	if err != nil {
		return nil, err
	}

	return &domain.User{ID: id, Name: params.Name, Email: params.Email}, nil
}
