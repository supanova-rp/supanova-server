package errors

import (
	stdErrors "errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	ErrInvalidUuid       = "invalid uuid format"
	ErrValidation        = "validation failed"
	ErrUserIDCtxNotFound = "user not found in context"
)

func Getting(resource string) string {
	return fmt.Sprintf("Error getting %s", resource)
}

func Creating(resource string) string {
	return fmt.Sprintf("Error adding %s", resource)
}

func Deleting(resource string) string {
	return fmt.Sprintf("Error deleting %s", resource)
}

func NotFound(resource string) string {
	return fmt.Sprintf("%s not found", resource)
}

func IsNotFoundErr(err error) bool {
	return stdErrors.Is(err, pgx.ErrNoRows)
}

func InvalidFormat(resource string) string {
	return fmt.Sprintf("Invalid %s format", resource)
}
