package models

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a typed application error with HTTP status and machine-readable code.
type AppError struct {
	Code       string `json:"error"`
	Message    string `json:"error_description"`
	HTTPStatus int    `json:"-"`
	Inner      error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Inner)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Inner
}

// Predefined error codes.
var (
	ErrNotFound = &AppError{
		Code:       "not_found",
		Message:    "The requested resource was not found.",
		HTTPStatus: http.StatusNotFound,
	}
	ErrUnauthorized = &AppError{
		Code:       "unauthorized",
		Message:    "Authentication is required.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrForbidden = &AppError{
		Code:       "forbidden",
		Message:    "You do not have permission to perform this action.",
		HTTPStatus: http.StatusForbidden,
	}
	ErrConflict = &AppError{
		Code:       "conflict",
		Message:    "The resource already exists.",
		HTTPStatus: http.StatusConflict,
	}
	ErrValidation = &AppError{
		Code:       "validation_error",
		Message:    "The request body contains invalid data.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrBadRequest = &AppError{
		Code:       "bad_request",
		Message:    "The request is malformed.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInternal = &AppError{
		Code:       "internal_error",
		Message:    "An internal error occurred.",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrRateLimited = &AppError{
		Code:       "rate_limited",
		Message:    "Too many requests. Please try again later.",
		HTTPStatus: http.StatusTooManyRequests,
	}
	ErrInvalidCredentials = &AppError{
		Code:       "invalid_credentials",
		Message:    "The email or password is incorrect.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrAccountBlocked = &AppError{
		Code:       "account_blocked",
		Message:    "This account has been blocked.",
		HTTPStatus: http.StatusForbidden,
	}
	ErrEmailNotVerified = &AppError{
		Code:       "email_not_verified",
		Message:    "Email address has not been verified.",
		HTTPStatus: http.StatusForbidden,
	}
	ErrMFARequired = &AppError{
		Code:       "mfa_required",
		Message:    "Multi-factor authentication is required.",
		HTTPStatus: http.StatusForbidden,
	}
	ErrMFAInvalidCode = &AppError{
		Code:       "mfa_invalid_code",
		Message:    "The MFA code is invalid or expired.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrTokenExpired = &AppError{
		Code:       "token_expired",
		Message:    "The token has expired.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrTokenRevoked = &AppError{
		Code:       "token_revoked",
		Message:    "The token has been revoked.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrInvalidGrant = &AppError{
		Code:       "invalid_grant",
		Message:    "The authorization grant is invalid.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInvalidClient = &AppError{
		Code:       "invalid_client",
		Message:    "Client authentication failed.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrInvalidScope = &AppError{
		Code:       "invalid_scope",
		Message:    "The requested scope is invalid.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrUnsupportedGrantType = &AppError{
		Code:       "unsupported_grant_type",
		Message:    "The grant type is not supported.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrPasswordBreached = &AppError{
		Code:       "password_breached",
		Message:    "This password has been found in a data breach. Please choose a different password.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrPasswordPolicy = &AppError{
		Code:       "password_policy",
		Message:    "The password does not meet the requirements.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrPasswordReused = &AppError{
		Code:       "password_reused",
		Message:    "This password has been used recently. Please choose a different password.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrSessionExpired = &AppError{
		Code:       "session_expired",
		Message:    "The session has expired.",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrWebAuthnFailed = &AppError{
		Code:       "webauthn_failed",
		Message:    "WebAuthn ceremony failed.",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrSAMLFailed = &AppError{
		Code:       "saml_failed",
		Message:    "SAML authentication failed.",
		HTTPStatus: http.StatusBadRequest,
	}
)

// NewAppError creates a new application error with a custom message.
func NewAppError(code string, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// Wrap wraps an inner error.
func (e *AppError) Wrap(inner error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Inner:      inner,
	}
}

// WithMessage creates a copy with a custom message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
		Inner:      e.Inner,
	}
}

// IsAppError checks if an error is an AppError and optionally matches a code.
func IsAppError(err error, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		if target != nil {
			return appErr.Code == target.Code
		}
		return true
	}
	return false
}

// GetAppError extracts AppError from an error chain.
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
