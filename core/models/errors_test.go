package models

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestAppError_Error_WithInner(t *testing.T) {
	inner := fmt.Errorf("database connection lost")
	appErr := &AppError{
		Code:       "db_error",
		Message:    "Database failed",
		HTTPStatus: 500,
		Inner:      inner,
	}

	got := appErr.Error()
	want := "db_error: Database failed (database connection lost)"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAppError_Error_WithoutInner(t *testing.T) {
	appErr := &AppError{
		Code:       "not_found",
		Message:    "Resource not found",
		HTTPStatus: 404,
	}

	got := appErr.Error()
	want := "not_found: Resource not found"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("root cause")
	appErr := &AppError{
		Code:       "test",
		Message:    "test error",
		HTTPStatus: 500,
		Inner:      inner,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != inner {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, inner)
	}
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	appErr := &AppError{
		Code:       "test",
		Message:    "no inner",
		HTTPStatus: 500,
	}

	if appErr.Unwrap() != nil {
		t.Error("Unwrap() should return nil when there is no inner error")
	}
}

func TestNewAppError(t *testing.T) {
	appErr := NewAppError("custom_code", "custom message", http.StatusTeapot)

	if appErr.Code != "custom_code" {
		t.Errorf("Code = %q, want %q", appErr.Code, "custom_code")
	}
	if appErr.Message != "custom message" {
		t.Errorf("Message = %q, want %q", appErr.Message, "custom message")
	}
	if appErr.HTTPStatus != http.StatusTeapot {
		t.Errorf("HTTPStatus = %d, want %d", appErr.HTTPStatus, http.StatusTeapot)
	}
	if appErr.Inner != nil {
		t.Error("Inner should be nil for NewAppError")
	}
}

func TestAppError_Wrap(t *testing.T) {
	original := ErrNotFound
	inner := fmt.Errorf("user ID 123")

	wrapped := original.Wrap(inner)

	if wrapped.Code != original.Code {
		t.Errorf("wrapped Code = %q, want %q", wrapped.Code, original.Code)
	}
	if wrapped.Message != original.Message {
		t.Errorf("wrapped Message = %q, want %q", wrapped.Message, original.Message)
	}
	if wrapped.HTTPStatus != original.HTTPStatus {
		t.Errorf("wrapped HTTPStatus = %d, want %d", wrapped.HTTPStatus, original.HTTPStatus)
	}
	if wrapped.Inner != inner {
		t.Errorf("wrapped Inner = %v, want %v", wrapped.Inner, inner)
	}
	// Ensure original is not mutated
	if original.Inner != nil {
		t.Error("original error should not be mutated by Wrap")
	}
}

func TestAppError_WithMessage(t *testing.T) {
	original := ErrValidation
	customMsg := "Field 'email' is required."

	modified := original.WithMessage(customMsg)

	if modified.Code != original.Code {
		t.Errorf("Code = %q, want %q", modified.Code, original.Code)
	}
	if modified.Message != customMsg {
		t.Errorf("Message = %q, want %q", modified.Message, customMsg)
	}
	if modified.HTTPStatus != original.HTTPStatus {
		t.Errorf("HTTPStatus = %d, want %d", modified.HTTPStatus, original.HTTPStatus)
	}
	// Ensure original is not mutated
	if original.Message == customMsg {
		t.Error("original error message should not be mutated by WithMessage")
	}
}

func TestIsAppError_Match(t *testing.T) {
	err := ErrNotFound.WithMessage("User not found")

	if !IsAppError(err, ErrNotFound) {
		t.Error("IsAppError should return true when codes match")
	}
}

func TestIsAppError_NoMatch(t *testing.T) {
	err := ErrNotFound.WithMessage("User not found")

	if IsAppError(err, ErrUnauthorized) {
		t.Error("IsAppError should return false when codes don't match")
	}
}

func TestIsAppError_NilTarget(t *testing.T) {
	err := ErrNotFound.WithMessage("something")

	if !IsAppError(err, nil) {
		t.Error("IsAppError with nil target should return true for any AppError")
	}
}

func TestIsAppError_NonAppError(t *testing.T) {
	err := fmt.Errorf("plain error")

	if IsAppError(err, ErrNotFound) {
		t.Error("IsAppError should return false for non-AppError")
	}
}

func TestIsAppError_Wrapped(t *testing.T) {
	inner := ErrNotFound.WithMessage("deep inside")
	wrapped := fmt.Errorf("outer: %w", inner)

	if !IsAppError(wrapped, ErrNotFound) {
		t.Error("IsAppError should work through error wrapping")
	}
}

func TestGetAppError(t *testing.T) {
	err := ErrUnauthorized.WithMessage("Token expired")
	result := GetAppError(err)

	if result == nil {
		t.Fatal("GetAppError should return non-nil for AppError")
	}
	if result.Code != "unauthorized" {
		t.Errorf("Code = %q, want %q", result.Code, "unauthorized")
	}
}

func TestGetAppError_NonAppError(t *testing.T) {
	err := fmt.Errorf("not an app error")
	result := GetAppError(err)

	if result != nil {
		t.Error("GetAppError should return nil for non-AppError")
	}
}

func TestGetAppError_WrappedAppError(t *testing.T) {
	appErr := ErrForbidden.WithMessage("no access")
	wrapped := fmt.Errorf("wrapping: %w", appErr)
	result := GetAppError(wrapped)

	if result == nil {
		t.Fatal("GetAppError should find AppError in wrapped error")
	}
	if result.Code != "forbidden" {
		t.Errorf("Code = %q, want %q", result.Code, "forbidden")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		code       string
		httpStatus int
	}{
		{"ErrNotFound", ErrNotFound, "not_found", http.StatusNotFound},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized", http.StatusUnauthorized},
		{"ErrForbidden", ErrForbidden, "forbidden", http.StatusForbidden},
		{"ErrConflict", ErrConflict, "conflict", http.StatusConflict},
		{"ErrValidation", ErrValidation, "validation_error", http.StatusBadRequest},
		{"ErrBadRequest", ErrBadRequest, "bad_request", http.StatusBadRequest},
		{"ErrInternal", ErrInternal, "internal_error", http.StatusInternalServerError},
		{"ErrRateLimited", ErrRateLimited, "rate_limited", http.StatusTooManyRequests},
		{"ErrInvalidCredentials", ErrInvalidCredentials, "invalid_credentials", http.StatusUnauthorized},
		{"ErrAccountBlocked", ErrAccountBlocked, "account_blocked", http.StatusForbidden},
		{"ErrEmailNotVerified", ErrEmailNotVerified, "email_not_verified", http.StatusForbidden},
		{"ErrMFARequired", ErrMFARequired, "mfa_required", http.StatusForbidden},
		{"ErrMFAInvalidCode", ErrMFAInvalidCode, "mfa_invalid_code", http.StatusUnauthorized},
		{"ErrTokenExpired", ErrTokenExpired, "token_expired", http.StatusUnauthorized},
		{"ErrTokenRevoked", ErrTokenRevoked, "token_revoked", http.StatusUnauthorized},
		{"ErrInvalidGrant", ErrInvalidGrant, "invalid_grant", http.StatusBadRequest},
		{"ErrInvalidClient", ErrInvalidClient, "invalid_client", http.StatusUnauthorized},
		{"ErrInvalidScope", ErrInvalidScope, "invalid_scope", http.StatusBadRequest},
		{"ErrUnsupportedGrantType", ErrUnsupportedGrantType, "unsupported_grant_type", http.StatusBadRequest},
		{"ErrPasswordBreached", ErrPasswordBreached, "password_breached", http.StatusBadRequest},
		{"ErrPasswordPolicy", ErrPasswordPolicy, "password_policy", http.StatusBadRequest},
		{"ErrPasswordReused", ErrPasswordReused, "password_reused", http.StatusBadRequest},
		{"ErrSessionExpired", ErrSessionExpired, "session_expired", http.StatusUnauthorized},
		{"ErrWebAuthnFailed", ErrWebAuthnFailed, "webauthn_failed", http.StatusBadRequest},
		{"ErrSAMLFailed", ErrSAMLFailed, "saml_failed", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("Code = %q, want %q", tt.err.Code, tt.code)
			}
			if tt.err.HTTPStatus != tt.httpStatus {
				t.Errorf("HTTPStatus = %d, want %d", tt.err.HTTPStatus, tt.httpStatus)
			}
			if tt.err.Message == "" {
				t.Error("Message should not be empty")
			}
		})
	}
}

func TestAppError_Is(t *testing.T) {
	// Test that errors.As works with AppError
	original := ErrNotFound.Wrap(fmt.Errorf("inner"))
	var target *AppError
	if !errors.As(original, &target) {
		t.Error("errors.As should find AppError")
	}
	if target.Code != "not_found" {
		t.Errorf("Code = %q, want %q", target.Code, "not_found")
	}
}

func TestAppError_ImplementsError(t *testing.T) {
	var _ error = &AppError{}
}
