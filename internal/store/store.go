package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/store/cache"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type Store struct {
	pool        *pgxpool.Pool
	Queries     *sqlc.Queries
	courseCache *cache.Cache[domain.Course]
}

func New(ctx context.Context, dbUrl string) (*Store, error) {
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

	err = runMigrations(dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	cache := cache.New[domain.Course]()

	return &Store{
		pool:    pool,
		Queries: sqlc.New(pool),
		courseCache:   cache,
	}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) PingDB(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
