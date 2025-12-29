package errors

import (
	stdErrors "errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	InvalidUUID        = "invalid uuid format"
	Validation         = "validation failed"
	InvalidRequestBody = "invalid request body"
	Unauthorised       = "Unauthorised"
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
