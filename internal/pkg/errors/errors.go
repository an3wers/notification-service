package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrDatabaseOperation = errors.New("database operation failed")
	ErrEmailSendFailed   = errors.New("email send failed")
	ErrQueueOperation    = errors.New("queue operation failed")
	ErrStorageOperation  = errors.New("storage operation failed")
	ErrDuplicateMessage  = errors.New("duplicate message")
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
