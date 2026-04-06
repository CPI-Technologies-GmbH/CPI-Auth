package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/policy"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
)

type contextKey string

const (
	ContextKeyTenantID      contextKey = "tenant_id"
	ContextKeyUserID        contextKey = "user_id"
	ContextKeyClaims        contextKey = "claims"
	ContextKeyCorrelationID contextKey = "correlation_id"
)

// GetTenantID extracts tenant ID from context.
func GetTenantID(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ContextKeyTenantID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ContextKeyUserID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// GetClaims extracts JWT claims from context.
func GetClaims(ctx context.Context) *tokens.AccessTokenClaims {
	if v, ok := ctx.Value(ContextKeyClaims).(*tokens.AccessTokenClaims); ok {
		return v
	}
	return nil
}

// GetCorrelationID extracts the correlation ID from context.
func GetCorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyCorrelationID).(string); ok {
		return v
	}
	return ""
}

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// errorLogger is the package-level logger used to surface 5xx errors that
// would otherwise vanish into a generic JSON response. Set via SetErrorLogger
// during server bootstrap.
var errorLogger *zap.Logger

// SetErrorLogger registers the logger that WriteError uses to log 5xx errors.
func SetErrorLogger(l *zap.Logger) {
	errorLogger = l
}

// WriteError writes an error response. 5xx errors (including any wrapped
// inner cause) are also logged so they don't disappear silently.
func WriteError(w http.ResponseWriter, err error) {
	if appErr := models.GetAppError(err); appErr != nil {
		if appErr.HTTPStatus >= 500 && errorLogger != nil {
			fields := []zap.Field{
				zap.String("code", appErr.Code),
				zap.String("message", appErr.Message),
			}
			if appErr.Inner != nil {
				fields = append(fields, zap.Error(appErr.Inner))
			}
			errorLogger.Error("server error response", fields...)
		}
		WriteJSON(w, appErr.HTTPStatus, appErr)
		return
	}
	if errorLogger != nil {
		errorLogger.Error("unhandled error response", zap.Error(err))
	}
	WriteJSON(w, http.StatusInternalServerError, models.ErrInternal)
}

// --- Middleware Implementations ---

// Authentication validates JWT access tokens.
func Authentication(tokenSvc *tokens.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				WriteError(w, models.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				WriteError(w, models.ErrUnauthorized.WithMessage("Invalid Authorization header format."))
				return
			}

			claims, err := tokenSvc.ValidateAccessToken(r.Context(), parts[1])
			if err != nil {
				WriteError(w, err)
				return
			}

			ctx := r.Context()
			userID, _ := uuid.Parse(claims.Subject)
			tenantID, _ := uuid.Parse(claims.TenantID)

			ctx = context.WithValue(ctx, ContextKeyClaims, claims)
			ctx = context.WithValue(ctx, ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, ContextKeyTenantID, tenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission checks if the authenticated user has a specific permission.
func RequirePermission(rbac *policy.RBACService, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == uuid.Nil {
				WriteError(w, models.ErrUnauthorized)
				return
			}

			has, err := rbac.HasPermission(r.Context(), userID, permission)
			if err != nil {
				WriteError(w, models.ErrInternal.Wrap(err))
				return
			}
			if !has {
				// Also check JWT claims permissions
				claims := GetClaims(r.Context())
				if claims != nil {
					for _, p := range claims.Permissions {
						if p == permission || p == "*" {
							next.ServeHTTP(w, r)
							return
						}
					}
				}
				WriteError(w, models.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// reservedSubdomains are the labels that must never resolve to a tenant via
// the subdomain branch of TenantResolver, even if a tenant with that slug
// exists. This prevents trivial hostname hijacks (e.g. someone creating a
// tenant with slug "auth" and capturing auth.cpi.dev).
var reservedSubdomains = map[string]bool{
	"auth":    true,
	"admin":   true,
	"api":     true,
	"www":     true,
	"app":     true,
	"login":   true,
	"account": true,
	"mail":    true,
	"static":  true,
	"assets":  true,
	"cdn":     true,
	"t":       true, // reserved for the path-based /t/{slug}/ scheme
}

// validSubdomainSlug matches the slug format we accept as a tenant subdomain
// label: 2-32 chars, lowercase alphanumerics and dashes, not starting or
// ending with a dash.
var validSubdomainSlug = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)

// TenantResolver resolves the tenant from subdomain, header, or token.
//
// Resolution order:
//  1. Existing tenant in context (e.g. from JWT middleware) — passthrough.
//  2. X-Tenant-ID header.
//  3. URL path prefix /t/{slug}/ — strips the prefix from the request URL
//     so downstream handlers see clean paths. This is the default
//     multi-tenant routing scheme.
//  4. Hostname: first checked against the configured base host (no tenant);
//     then as a subdomain label (after blacklist + format validation);
//     then as a custom domain via the verified-domains table.
func TenantResolver(tenantRepo models.TenantRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Already resolved from JWT?
			if GetTenantID(ctx) != uuid.Nil {
				next.ServeHTTP(w, r)
				return
			}

			var tenantID uuid.UUID

			// 1. X-Tenant-ID header (admin tools, internal calls)
			if headerID := r.Header.Get("X-Tenant-ID"); headerID != "" {
				parsed, err := uuid.Parse(headerID)
				if err == nil {
					tenantID = parsed
				}
			}

			// 2. Path prefix /t/{slug}/...
			if tenantID == uuid.Nil && strings.HasPrefix(r.URL.Path, "/t/") {
				rest := r.URL.Path[len("/t/"):]
				slug := rest
				remainder := "/"
				if i := strings.Index(rest, "/"); i >= 0 {
					slug = rest[:i]
					remainder = rest[i:]
				}
				if validSubdomainSlug.MatchString(slug) {
					tenant, err := tenantRepo.GetBySlug(ctx, slug)
					if err == nil {
						tenantID = tenant.ID
						// Strip the /t/{slug} prefix so downstream handlers
						// continue to see clean URLs (e.g. /oauth/authorize).
						r2 := r.Clone(r.Context())
						r2.URL.Path = remainder
						r2.URL.RawPath = ""
						r = r2
					}
				}
			}

			// 3. Hostname: subdomain or custom domain.
			if tenantID == uuid.Nil {
				host := r.Host
				if idx := strings.LastIndex(host, ":"); idx != -1 {
					host = host[:idx]
				}
				host = strings.ToLower(host)

				// 3a. Subdomain branch — only if the first label is non-reserved
				// and matches the slug format.
				parts := strings.Split(host, ".")
				if len(parts) >= 3 {
					slug := parts[0]
					if !reservedSubdomains[slug] && validSubdomainSlug.MatchString(slug) {
						tenant, err := tenantRepo.GetBySlug(ctx, slug)
						if err == nil {
							tenantID = tenant.ID
						}
					}
				}

				// 3b. Custom domain branch (e.g. login.example.com)
				if tenantID == uuid.Nil {
					tenant, err := tenantRepo.GetByDomain(ctx, host)
					if err == nil {
						tenantID = tenant.ID
					}
				}
			}

			if tenantID != uuid.Nil {
				ctx = context.WithValue(ctx, ContextKeyTenantID, tenantID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RateLimiter implements a per-key sliding window rate limiter.
type rateLimiter struct {
	mu       sync.Mutex
	counters map[string]*rateBucket
	limit    int
	window   time.Duration
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		counters: make(map[string]*rateBucket),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, ok := rl.counters[key]
	if !ok || now.After(bucket.resetAt) {
		rl.counters[key] = &rateBucket{count: 1, resetAt: now.Add(rl.window)}
		return true
	}

	if bucket.count >= rl.limit {
		return false
	}
	bucket.count++
	return true
}

// RateLimit returns a rate limiting middleware.
func RateLimit(requestsPerSecond int) func(http.Handler) http.Handler {
	limiter := newRateLimiter(requestsPerSecond, time.Second)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if userID := GetUserID(r.Context()); userID != uuid.Nil {
				key = userID.String()
			}

			if !limiter.Allow(key) {
				WriteError(w, models.ErrRateLimited)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CorrelationID adds a correlation ID to each request for tracing.
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), ContextKeyCorrelationID, correlationID)
		w.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestLogger logs each HTTP request with structured fields.
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			path := maskPathPII(r.URL.Path)

			logger.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", path),
				zap.Int("status", ww.status),
				zap.Duration("duration", duration),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("correlation_id", GetCorrelationID(r.Context())),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// CSPHeaders sets Content Security Policy and other security headers.
func CSPHeaders(cspHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", cspHeader)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	}
}

// Recovery catches panics and returns a 500 error.
func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						zap.Any("error", rec),
						zap.String("stack", string(debug.Stack())),
					)
					WriteJSON(w, http.StatusInternalServerError, models.ErrInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func maskPathPII(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if _, err := uuid.Parse(part); err == nil {
			parts[i] = part[:8] + "***"
		}
		if strings.Contains(part, "@") {
			if atIdx := strings.Index(part, "@"); atIdx > 2 {
				parts[i] = part[:2] + "***" + part[atIdx:]
			}
		}
	}
	return strings.Join(parts, "/")
}

// SuperAdminTenantOverride allows super-admins to override the tenant context
// via the X-Tenant-ID header. Runs after Authentication and RequirePermission.
func SuperAdminTenantOverride() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerID := r.Header.Get("X-Tenant-ID")
			if headerID == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims := GetClaims(r.Context())
			if claims == nil {
				next.ServeHTTP(w, r)
				return
			}

			isSuperAdmin := false
			for _, p := range claims.Permissions {
				if p == "*" {
					isSuperAdmin = true
					break
				}
			}
			if !isSuperAdmin {
				next.ServeHTTP(w, r)
				return
			}

			parsed, err := uuid.Parse(headerID)
			if err != nil {
				WriteError(w, models.ErrValidation.WithMessage("Invalid X-Tenant-ID header."))
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyTenantID, parsed)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APIKeyAuth validates API key authentication.
func APIKeyAuth(apiKeyRepo models.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}
			if apiKey == "" {
				WriteError(w, models.ErrUnauthorized.WithMessage("API key required."))
				return
			}

			h := sha256.Sum256([]byte(apiKey))
			keyHash := hex.EncodeToString(h[:])
			key, err := apiKeyRepo.GetByKeyHash(r.Context(), keyHash)
			if err != nil {
				WriteError(w, models.ErrUnauthorized.WithMessage("Invalid API key."))
				return
			}

			if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
				WriteError(w, models.ErrUnauthorized.WithMessage("API key has expired."))
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyTenantID, key.TenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
