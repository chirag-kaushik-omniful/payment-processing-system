package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL_ERROR"
	CodeRateLimit    Code = "RATE_LIMIT_EXCEEDED"
	CodeIdempotent   Code = "IDEMPOTENT_REPLAY"
)

type AppError struct {
	Code       Code   `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(code Code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

func Validation(msg string) *AppError {
	return New(CodeValidation, msg, http.StatusBadRequest)
}

func Unauthorized(msg string) *AppError {
	return New(CodeUnauthorized, msg, http.StatusUnauthorized)
}

func NotFound(msg string) *AppError {
	return New(CodeNotFound, msg, http.StatusNotFound)
}

func Conflict(msg string) *AppError {
	return New(CodeConflict, msg, http.StatusConflict)
}

func Internal(msg string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: msg, HTTPStatus: http.StatusInternalServerError, Err: err}
}

func RateLimit(msg string) *AppError {
	return New(CodeRateLimit, msg, http.StatusTooManyRequests)
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
