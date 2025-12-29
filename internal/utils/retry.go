package utils

import (
	"context"
	"log/slog"
	"time"
)

func RetryWithExponentialBackoff[T any](
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
