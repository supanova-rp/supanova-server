package errors

import (
	"context"
	stdErrors "errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	InvalidUUID        = "invalid uuid format"
	Validation         = "validation failed"
	InvalidRequestBody = "invalid request body"
	Unauthorised       = "Unauthorised"
	DbMaxRetries       = 5
	DbBaseDelay        = 100 * time.Millisecond
)

func Getting(resource string) string {
	return fmt.Sprintf("Error getting %s", resource)
}

func Creating(resource string) string {
	return fmt.Sprintf("Error creating %s", resource)
}

func Deleting(resource string) string {
	return fmt.Sprintf("Error deleting %s", resource)
}

func Updating(resource string) string {
	return fmt.Sprintf("Error updating %s", resource)
}

func NotFound(resource string) string {
	return fmt.Sprintf("%s not found", resource)
}

func NotFoundInCtx(resource string) string {
	return fmt.Sprintf("%s not found in context", resource)
}

func IsNotFoundErr(err error) bool {
	return stdErrors.Is(err, pgx.ErrNoRows)
}

func InvalidFormat(resource string) string {
	return fmt.Sprintf("Invalid %s format", resource)
}

func Forbidden(resource string) string {
	return fmt.Sprintf("No permissions for %s", resource)
}

func Wrap(text string) error {
	return stdErrors.New(text)
}

func RetryDbQueryWithExponentialBackoff[T any](ctx context.Context, query func() (T, error)) (T, error) {
	return retryWithExponentialBackoff(ctx, query, DbMaxRetries, DbBaseDelay, isRetryableDbError)
}

func isRetryableDbError(err error) bool {
	if err == nil {
		return false
	}

	if IsNotFoundErr(err) {
		return false
	}

	var pgErr *pgconn.PgError
	if stdErrors.As(err, &pgErr) {
		errClass := pgErr.Code[:2]
		// Retry on transient error classes
		switch errClass {
		case "08",
			"40",
			"53",
			"55",
			"57":
			return true
		}

		return false
	}

	return false
}

func retryWithExponentialBackoff[T any](
	ctx context.Context,
	callback func() (T, error),
	maxRetries int,
	baseDelay time.Duration,
	shouldRetry func(error) bool,
) (T, error) {
	var result T
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		result, err = callback()
		if err == nil {
			slog.InfoContext(ctx, "operation succeeded after retry", "attempt", attempt)
			return result, err
		}

		if shouldRetry != nil && !shouldRetry(err) {
			return result, err
		}

		// don't sleep after the last attempt
		if attempt == maxRetries {
			slog.ErrorContext(ctx, "operation failed after all retries", "attempts", attempt, "error", err)
			break
		}

		delay := baseDelay * time.Duration(1<<attempt)
		slog.ErrorContext(ctx, "operation failed, retrying", "attempt", attempt, "next_retry_in", delay, "error", err)

		select {
		case <-time.After(delay):
			// continue to next retry
		case <-ctx.Done():
			return result, ctx.Err()
		}
	}

	return result, err
}
