package expections

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeValidation   Code = "VALIDATION"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeInternal     Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Message string
	Cause   error
}

func New(code Code, message string) error {
	return &Error{Code: code, Message: message}
}

func Wrap(code Code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func IsCode(err error, code Code) bool {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Code == code
}

func CodeOf(err error) Code {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return CodeInternal
	}
	if appErr.Code == "" {
		return CodeInternal
	}
	return appErr.Code
}

func Validation(message string) error {
	return New(CodeValidation, message)
}

func NotFound(message string, cause error) error {
	return Wrap(CodeNotFound, message, cause)
}

func Conflict(message string, cause error) error {
	return Wrap(CodeConflict, message, cause)
}

func Internal(message string, cause error) error {
	if message == "" && cause != nil {
		message = fmt.Sprint(cause)
	}
	return Wrap(CodeInternal, message, cause)
}
