package model

import (
	"errors"
	"fmt"
)

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when a user with the same email already exists.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidInput is returned when the input is invalid.
	ErrInvalidInput = errors.New("invalid input")

	// ErrPasswordMismatch is returned when a password does not match its hash.
	ErrPasswordMismatch = errors.New("password mismatch")

	// ErrPasswordTooLong is returned when a password exceeds the hasher's byte limit
	// (72 bytes for bcrypt; multi-byte characters count more than once).
	ErrPasswordTooLong = errors.New("password too long")

	// ErrUnauthorized is returned when the user is not authorized.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInternal is returned for internal server errors.
	ErrInternal = errors.New("internal server error")
)

// DomainError represents a domain error with additional context.
type DomainError struct {
	Err     error
	Message string
	Code    string
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

// Unwrap implements the errors.Unwrap interface.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new DomainError.
func NewDomainError(code, message string, err error) *DomainError {
	return &DomainError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}
