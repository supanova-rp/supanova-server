package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type Store struct {
	pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func NewStore(ctx context.Context, dbUrl string, shouldRunMigrations bool) (*Store, error) {
	poolConfig, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %v", err)
	}

	// Currently we don't want to run migrations in prod because they are handled by node app.
	// In future we want to stop handling them with the node app and handle them here.
	if shouldRunMigrations {
		err = runMigrations(dbUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to run migrations: %v", err)
		}
	}

	return &Store{
		pool:    pool,
		Queries: sqlc.New(pool),
	}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) PingDB(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
