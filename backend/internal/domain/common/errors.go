package common

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Common domain error constructors

func NewValidationError(message string, err error) *DomainError {
	return &DomainError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Err:     err,
	}
}

func NewNotFoundError(entityType string, id string) *DomainError {
	return &DomainError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s with id %s not found", entityType, id),
	}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{
		Code:    "CONFLICT",
		Message: message,
	}
}

func NewInvariantViolationError(message string) *DomainError {
	return &DomainError{
		Code:    "INVARIANT_VIOLATION",
		Message: message,
	}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

