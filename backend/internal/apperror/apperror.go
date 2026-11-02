package apperror

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Kind int

const (
	KindInternal Kind = iota
	KindValidation
	KindNotFound
	KindConflict
	KindUnavailable
)

// Error is a sanitized error safe to show to API clients.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Details map[string]string
}

func (e *Error) Error() string { return e.Message }

func Validation(msg string, details map[string]string) *Error {
	return &Error{Kind: KindValidation, Code: "VALIDATION_ERROR", Message: msg, Details: details}
}

func NotFound(msg string) *Error {
	return &Error{Kind: KindNotFound, Code: "NOT_FOUND", Message: msg}
}

func Conflict(msg string) *Error {
	return &Error{Kind: KindConflict, Code: "CONFLICT", Message: msg}
}

func Unavailable() *Error {
	return &Error{Kind: KindUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Service temporarily unavailable. Please try again later."}
}

func Internal() *Error {
	return &Error{Kind: KindInternal, Code: "INTERNAL_ERROR", Message: "An unexpected error occurred."}
}

// HTTPStatus maps the error kind to an HTTP status code.
func (e *Error) HTTPStatus() int {
	switch e.Kind {
	case KindValidation:
		return http.StatusBadRequest
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// As converts any error into an *Error, never leaking the original message.
func As(err error) *Error {
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return FromMongo(err)
}

// FromMongo classifies a driver (or context) error into a sanitized *Error.
func FromMongo(err error) *Error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return NotFound("User not found.")
	case mongo.IsDuplicateKeyError(err):
		return Conflict("A user with this email already exists.")
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled),
		mongo.IsTimeout(err), mongo.IsNetworkError(err),
		strings.Contains(err.Error(), "server selection"),
		strings.Contains(err.Error(), "connection"):
		return Unavailable()
	default:
		return Internal()
	}
}
