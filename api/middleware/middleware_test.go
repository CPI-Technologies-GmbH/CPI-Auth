package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
)

// --- Context Helpers Tests ---

func TestGetTenantID(t *testing.T) {
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), ContextKeyTenantID, tenantID)

	got := GetTenantID(ctx)
	if got != tenantID {
		t.Errorf("GetTenantID = %v, want %v", got, tenantID)
	}
}

func TestGetTenantID_Missing(t *testing.T) {
	got := GetTenantID(context.Background())
	if got != uuid.Nil {
		t.Errorf("GetTenantID should return uuid.Nil when not set, got %v", got)
	}
}

func TestGetUserID(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), ContextKeyUserID, userID)

	got := GetUserID(ctx)
	if got != userID {
		t.Errorf("GetUserID = %v, want %v", got, userID)
	}
}

func TestGetUserID_Missing(t *testing.T) {
	got := GetUserID(context.Background())
	if got != uuid.Nil {
		t.Errorf("GetUserID should return uuid.Nil when not set, got %v", got)
	}
}

func TestGetClaims(t *testing.T) {
	claims := &tokens.AccessTokenClaims{
		TenantID: "tenant-1",
		Email:    "user@test.com",
	}
	ctx := context.WithValue(context.Background(), ContextKeyClaims, claims)

	got := GetClaims(ctx)
	if got == nil {
		t.Fatal("GetClaims returned nil")
	}
	if got.Email != "user@test.com" {
		t.Errorf("Email = %q, want %q", got.Email, "user@test.com")
	}
}

func TestGetClaims_Missing(t *testing.T) {
	got := GetClaims(context.Background())
	if got != nil {
		t.Error("GetClaims should return nil when not set")
	}
}

func TestGetCorrelationID(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyCorrelationID, "corr-123")

	got := GetCorrelationID(ctx)
	if got != "corr-123" {
		t.Errorf("GetCorrelationID = %q, want %q", got, "corr-123")
	}
}

func TestGetCorrelationID_Missing(t *testing.T) {
	got := GetCorrelationID(context.Background())
	if got != "" {
		t.Errorf("GetCorrelationID should return empty string when not set, got %q", got)
	}
}

// --- WriteJSON ---

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"hello": "world"}
	WriteJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["hello"] != "world" {
		t.Errorf("response body hello = %q, want %q", result["hello"], "world")
	}
}

func TestWriteJSON_DifferentStatus(t *testing.T) {
	w := httptest.NewRecorder()

	WriteJSON(w, http.StatusCreated, map[string]string{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

// --- WriteError ---

func TestWriteError_AppError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, models.ErrNotFound)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["error"] != "not_found" {
		t.Errorf("error = %q, want %q", result["error"], "not_found")
	}
}

func TestWriteError_WrappedAppError(t *testing.T) {
	w := httptest.NewRecorder()

	err := models.ErrUnauthorized.WithMessage("Token expired")
	WriteError(w, err)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestWriteError_GenericError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, fmt.Errorf("something broke"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestWriteError_NilError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, nil)

	// nil error should result in 500 since GetAppError returns nil
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// --- CorrelationID Middleware ---

func TestCorrelationID_GeneratesNew(t *testing.T) {
	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := GetCorrelationID(r.Context())
		if corrID == "" {
			t.Error("correlation ID should be set in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	corrID := w.Header().Get("X-Correlation-ID")
	if corrID == "" {
		t.Error("X-Correlation-ID header should be set in response")
	}
	// Should be a valid UUID
	if _, err := uuid.Parse(corrID); err != nil {
		t.Errorf("generated correlation ID should be a valid UUID, got %q", corrID)
	}
}

func TestCorrelationID_UsesExisting(t *testing.T) {
	existingCorrID := "my-custom-correlation-id"

	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := GetCorrelationID(r.Context())
		if corrID != existingCorrID {
			t.Errorf("correlation ID = %q, want %q", corrID, existingCorrID)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", existingCorrID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Correlation-ID") != existingCorrID {
		t.Errorf("response should echo existing correlation ID")
	}
}

// --- RateLimit Middleware ---

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	middleware := RateLimit(10)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	middleware := RateLimit(3) // 3 requests per second
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if i < 3 && w.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want %d", i, w.Code, http.StatusOK)
		}
		if i >= 3 && w.Code != http.StatusTooManyRequests {
			t.Errorf("request %d: status = %d, want %d", i, w.Code, http.StatusTooManyRequests)
		}
	}
}

func TestRateLimit_DifferentClients(t *testing.T) {
	middleware := RateLimit(2)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Client 1: 2 requests (should work)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "1.1.1.1:100"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("client1 request %d: status = %d, want %d", i, w.Code, http.StatusOK)
		}
	}

	// Client 2: should still be able to request (different IP)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "2.2.2.2:200"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("client2: status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --- CSPHeaders Middleware ---

func TestCSPHeaders(t *testing.T) {
	csp := "default-src 'self'; script-src 'none'"
	middleware := CSPHeaders(csp)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	tests := []struct {
		header string
		want   string
	}{
		{"Content-Security-Policy", csp},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Strict-Transport-Security", "max-age=31536000; includeSubDomains"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := w.Header().Get(tt.header)
			if got != tt.want {
				t.Errorf("%s = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

// --- Recovery Middleware ---

func TestRecovery(t *testing.T) {
	logger := zap.NewNop()
	middleware := Recovery(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something terrible happened")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["error"] != "internal_error" {
		t.Errorf("error = %q, want %q", result["error"], "internal_error")
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	logger := zap.NewNop()
	middleware := Recovery(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --- RequestLogger Middleware ---

func TestRequestLogger(t *testing.T) {
	logger := zap.NewNop()
	middleware := RequestLogger(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRequestLogger_StatusCapture(t *testing.T) {
	logger := zap.NewNop()
	middleware := RequestLogger(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- statusWriter ---

func TestStatusWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

	sw.WriteHeader(http.StatusNotFound)

	if sw.status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", sw.status, http.StatusNotFound)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("underlying writer status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- maskPathPII ---

func TestMaskPathPII_UUID(t *testing.T) {
	id := uuid.New().String()
	path := "/api/v1/users/" + id

	masked := maskPathPII(path)

	if masked == path {
		t.Error("UUID in path should be masked")
	}
	// UUID should be partially masked: first 8 chars + ***
	expected := "/api/v1/users/" + id[:8] + "***"
	if masked != expected {
		t.Errorf("masked = %q, want %q", masked, expected)
	}
}

func TestMaskPathPII_Email(t *testing.T) {
	path := "/api/v1/users/user@example.com"
	masked := maskPathPII(path)

	if masked == path {
		t.Error("email in path should be masked")
	}
}

func TestMaskPathPII_NoSensitiveData(t *testing.T) {
	path := "/api/v1/health"
	masked := maskPathPII(path)

	if masked != path {
		t.Errorf("path without sensitive data should not be changed, got %q", masked)
	}
}
