// Package apierror provides a small typed error used across services so that
// HTTP handlers can map domain errors to consistent JSON responses and status
// codes.
package apierror

import "net/http"

// Error is an API-facing error carrying an HTTP status, a stable machine code
// and a human-readable message.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

// New builds an Error.
func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// Common constructors for frequently used error classes.
func ErrBadRequest(msg string) *Error   { return New(http.StatusBadRequest, "bad_request", msg) }
func ErrUnauthorized(msg string) *Error { return New(http.StatusUnauthorized, "unauthorized", msg) }
func ErrForbidden(msg string) *Error    { return New(http.StatusForbidden, "forbidden", msg) }
func ErrNotFound(msg string) *Error     { return New(http.StatusNotFound, "not_found", msg) }
func ErrConflict(msg string) *Error     { return New(http.StatusConflict, "conflict", msg) }
func ErrInternal(msg string) *Error     { return New(http.StatusInternalServerError, "internal_error", msg) }
