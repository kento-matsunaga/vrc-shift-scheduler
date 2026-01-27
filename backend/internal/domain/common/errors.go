package common

import "fmt"

// Error codes
const (
	ErrInvalidInput = "INVALID_INPUT"
	ErrNotFound     = "NOT_FOUND"
	ErrConflict     = "CONFLICT"
	ErrUnauthorized = "UNAUTHORIZED"
)

// DomainError represents a domain-specific error
type DomainError struct {
	code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.code, e.Message)
}

func (e *DomainError) Code() string {
	return e.code
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error with a code and message
func NewDomainError(code string, message string) *DomainError {
	return &DomainError{
		code:    code,
		Message: message,
	}
}

// Common domain error constructors

func NewValidationError(message string, err error) *DomainError {
	return &DomainError{
		code:    ErrInvalidInput,
		Message: message,
		Err:     err,
	}
}

func NewNotFoundError(entityType string, id string) *DomainError {
	return &DomainError{
		code:    ErrNotFound,
		Message: fmt.Sprintf("%s with id %s not found", entityType, id),
	}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{
		code:    ErrConflict,
		Message: message,
	}
}

func NewInvariantViolationError(message string) *DomainError {
	return &DomainError{
		code:    "INVARIANT_VIOLATION",
		Message: message,
	}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		code:    ErrUnauthorized,
		Message: message,
	}
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code() == ErrNotFound
	}
	return false
}
