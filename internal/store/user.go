package store

import (
	"context"

	"github.com/supanova-rp/supanova-server/internal/domain"
)

func (s *Store) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.Queries.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:    user.ID,
		Name:  user.Name.String,
		Email: user.Email.String,
	}, nil
}
