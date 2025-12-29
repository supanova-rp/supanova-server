package store

import (
	"context"
	stdErrors "errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/store/cache"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

const (
	DbMaxRetries = 5
	DbBaseDelay  = 100 * time.Millisecond
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

	courseCache := cache.New[domain.Course]()

	return &Store{
		pool:        pool,
		Queries:     sqlc.New(pool),
		courseCache: courseCache,
	}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) PingDB(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func ExecQuery[T any](ctx context.Context, query func() (T, error)) (T, error) {
	return utils.RetryWithExponentialBackoff(ctx, query, DbMaxRetries, DbBaseDelay, isRetryableDbError)
}

func ExecCommand(ctx context.Context, command func() error) error {
	_, err := ExecQuery(ctx, func() (*struct{}, error) { return nil, command() })
	return err
}

var transientPostgresErrorCodes = []string{
	"08", // Connection exceptions (network problems, can't reach database)
	"40", // Transaction rollback (like deadlocks or serialisation failures)
	"53", // Insufficient resources (out of memory, disk full)
	"55", // Object not in prerequisite state (like trying to use a prepared statement that doesn't exist)
	"57", // Operator intervention (admin killed the query, database shutting down)
}

func isRetryableDbError(err error) bool {
	if err == nil {
		return false
	}

	if errors.IsNotFoundErr(err) {
		return false
	}

	var pgErr *pgconn.PgError
	if stdErrors.As(err, &pgErr) {
		errClass := pgErr.Code[:2]
		return slices.Contains(transientPostgresErrorCodes, errClass)
	}

	return false
}
