package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
)

// Kind categorizes the class of error for transport mapping.
type Kind string

const (
	KindValidation   Kind = "validation"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindUnavailable  Kind = "unavailable"
	KindInternal     Kind = "internal"
)

// AppError is a rich domain/application error with metadata useful across services.
type AppError struct {
	Kind    Kind
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// New creates an AppError with optional wrapped cause.
func New(kind Kind, code, message string, cause error) *AppError {
	return &AppError{Kind: kind, Code: code, Message: message, Cause: cause}
}

// IsKind determines whether err or any wrapped error is AppError with a specific kind.
func IsKind(err error, kind Kind) bool {
	var appErr *AppError
	if !stderrors.As(err, &appErr) {
		return false
	}
	return appErr.Kind == kind
}

// HTTPStatus returns canonical status code for the given error.
func HTTPStatus(err error) int {
	var appErr *AppError
	if !stderrors.As(err, &appErr) {
		return http.StatusInternalServerError
	}

	switch appErr.Kind {
	case KindValidation:
		return http.StatusBadRequest
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
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
