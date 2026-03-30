package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status represents user account status.
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusBlocked   Status = "blocked"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

// ApplicationType represents the type of OAuth application.
type ApplicationType string

const (
	AppTypeSPA    ApplicationType = "spa"
	AppTypeNative ApplicationType = "native"
	AppTypeWeb    ApplicationType = "web"
	AppTypeM2M    ApplicationType = "m2m"
)

// MFAMethod represents a multi-factor authentication method.
type MFAMethod string

const (
	MFAMethodTOTP     MFAMethod = "totp"
	MFAMethodSMS      MFAMethod = "sms"
	MFAMethodEmail    MFAMethod = "email"
	MFAMethodWebAuthn MFAMethod = "webauthn"
)

// Tenant represents an isolated tenant in the multi-tenant system.
type Tenant struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Name      string          `json:"name" db:"name" validate:"required,min=1,max=255"`
	Slug      string          `json:"slug" db:"slug" validate:"required,min=1,max=100"`
	Domain    string          `json:"domain,omitempty" db:"domain"`
	ParentID  *uuid.UUID      `json:"parent_id,omitempty" db:"parent_id"`
	Settings  json.RawMessage `json:"settings,omitempty" db:"settings"`
	Branding  json.RawMessage `json:"branding,omitempty" db:"branding"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// User represents a user account within a tenant.
type User struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Email         string          `json:"email" db:"email" validate:"required,email"`
	Phone         string          `json:"phone,omitempty" db:"phone"`
	Name          string          `json:"name,omitempty" db:"name"`
	AvatarURL     string          `json:"avatar_url,omitempty" db:"avatar_url"`
	PasswordHash  string          `json:"-" db:"password_hash"`
	Locale        string          `json:"locale,omitempty" db:"locale"`
	Metadata      json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	AppMetadata   json.RawMessage `json:"app_metadata,omitempty" db:"app_metadata"`
	Status        Status          `json:"status" db:"status"`
	EmailVerified bool            `json:"email_verified" db:"email_verified"`
	PhoneVerified bool            `json:"phone_verified" db:"phone_verified"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// Identity represents an external identity linked to a user (social, SAML, etc.).
type Identity struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	Provider        string          `json:"provider" db:"provider" validate:"required"`
	ProviderUserID  string          `json:"provider_user_id" db:"provider_user_id" validate:"required"`
	TokensEncrypted []byte          `json:"-" db:"tokens_encrypted"`
	Profile         json.RawMessage `json:"profile,omitempty" db:"profile"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// Application represents an OAuth/OIDC client application.
type Application struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	TenantID         uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name             string          `json:"name" db:"name" validate:"required,min=1,max=255"`
	Description      string          `json:"description" db:"description"`
	Type             ApplicationType `json:"type" db:"type" validate:"required,oneof=spa native web m2m"`
	ClientID         string          `json:"client_id" db:"client_id"`
	ClientSecretHash string          `json:"-" db:"client_secret_hash"`
	LogoURL          string          `json:"logo_url,omitempty" db:"logo_url"`
	RedirectURIs     []string        `json:"redirect_uris" db:"redirect_uris"`
	AllowedOrigins   []string        `json:"allowed_origins" db:"allowed_origins"`
	AllowedLogoutURLs []string       `json:"allowed_logout_urls" db:"post_logout_redirect_uris"`
	GrantTypes       []string        `json:"grant_types" db:"grant_types"`
	AccessTokenTTL   *int            `json:"access_token_ttl,omitempty" db:"access_token_ttl"`
	RefreshTokenTTL  *int            `json:"refresh_token_ttl,omitempty" db:"refresh_token_ttl"`
	IDTokenTTL       *int            `json:"id_token_ttl,omitempty" db:"id_token_ttl"`
	IsActive         bool            `json:"is_active" db:"is_active"`
	Settings         json.RawMessage `json:"settings,omitempty" db:"settings"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// Organization represents a group/company within a tenant.
type Organization struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	TenantID  uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name      string          `json:"name" db:"name" validate:"required,min=1,max=255"`
	Slug      string          `json:"slug" db:"slug" validate:"required,min=1,max=100"`
	Domains   []string        `json:"domains,omitempty" db:"domains"`
	Settings  json.RawMessage `json:"settings,omitempty" db:"settings"`
	Branding  json.RawMessage `json:"branding,omitempty" db:"branding"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// OrganizationMember represents user membership within an organization.
type OrganizationMember struct {
	OrgID     uuid.UUID `json:"organization_id" db:"organization_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Role represents an RBAC role with permissions.
type Role struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name         string     `json:"name" db:"name" validate:"required,min=1,max=100"`
	Description  string     `json:"description,omitempty" db:"description"`
	IsSystem     bool       `json:"is_system" db:"is_system"`
	Permissions  []string   `json:"permissions" db:"permissions"`
	ParentRoleID *uuid.UUID `json:"parent_role_id,omitempty" db:"parent_role_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Permission represents a dynamic permission definition stored in the database.
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	DisplayName string    `json:"display_name" db:"display_name" validate:"required,min=1,max=255"`
	Description string    `json:"description,omitempty" db:"description"`
	GroupName   string    `json:"group_name" db:"group_name" validate:"required,min=1,max=100"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ApplicationPermission represents a permission whitelist entry for an application.
type ApplicationPermission struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ApplicationID  uuid.UUID `json:"application_id" db:"application_id"`
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	PermissionName string    `json:"permission_name" db:"permission_name"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// CustomFieldDefinition represents a configurable user field for a tenant.
type CustomFieldDefinition struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	TenantID        uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name            string          `json:"name" db:"name"`
	Label           string          `json:"label" db:"label"`
	FieldType       string          `json:"field_type" db:"field_type"`
	Placeholder     string          `json:"placeholder,omitempty" db:"placeholder"`
	Description     string          `json:"description,omitempty" db:"description"`
	Options         json.RawMessage `json:"options,omitempty" db:"options"`
	Required        bool            `json:"required" db:"required"`
	VisibleOn       string          `json:"visible_on" db:"visible_on"`
	Position        int             `json:"position" db:"position"`
	ValidationRules json.RawMessage `json:"validation_rules,omitempty" db:"validation_rules"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// DomainVerification tracks custom domain DNS verification.
type DomainVerification struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Domain             string     `json:"domain" db:"domain"`
	VerificationToken  string     `json:"verification_token" db:"verification_token"`
	VerificationMethod string     `json:"verification_method" db:"verification_method"`
	IsVerified         bool       `json:"is_verified" db:"is_verified"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// Session represents an active user session.
type Session struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	DeviceInfo   json.RawMessage `json:"device_info,omitempty" db:"device_info"`
	IP           string          `json:"ip" db:"ip"`
	UserAgent    string          `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time       `json:"expires_at" db:"expires_at"`
	LastActiveAt time.Time       `json:"last_active_at" db:"last_active_at"`
}

// AuditLog represents an immutable audit trail entry.
type AuditLog struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	TenantID   uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	ActorID    *uuid.UUID      `json:"actor_id,omitempty" db:"actor_id"`
	Action     string          `json:"action" db:"action" validate:"required"`
	TargetType string          `json:"target_type,omitempty" db:"target_type"`
	TargetID   string          `json:"target_id,omitempty" db:"target_id"`
	Metadata   json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	IP         string          `json:"ip,omitempty" db:"ip"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// OAuthGrant represents an authorization code grant.
type OAuthGrant struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	UserID              uuid.UUID `json:"user_id" db:"user_id"`
	ApplicationID       uuid.UUID `json:"application_id" db:"application_id"`
	TenantID            uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Scopes              []string  `json:"scopes" db:"scopes"`
	Code                string    `json:"-" db:"code"`
	CodeChallenge       string    `json:"-" db:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method" db:"code_challenge_method"`
	RedirectURI         string    `json:"redirect_uri" db:"redirect_uri"`
	Nonce               string    `json:"nonce,omitempty" db:"nonce"`
	ExpiresAt           time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// RefreshToken represents an opaque refresh token.
type RefreshToken struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	ApplicationID uuid.UUID `json:"application_id" db:"application_id"`
	TenantID      uuid.UUID `json:"tenant_id" db:"tenant_id"`
	TokenHash     string    `json:"-" db:"token_hash"`
	Family        string    `json:"family" db:"family"`
	Revoked       bool      `json:"revoked" db:"revoked"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// MFAEnrollment represents a user's MFA enrollment.
type MFAEnrollment struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	Method          MFAMethod `json:"method" db:"method" validate:"required,oneof=totp sms email webauthn"`
	SecretEncrypted []byte    `json:"-" db:"secret_encrypted"`
	Verified        bool      `json:"verified" db:"verified"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// RecoveryCode represents a one-time use recovery code.
type RecoveryCode struct {
	ID       uuid.UUID `json:"id" db:"id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	CodeHash string    `json:"-" db:"code_hash"`
	Used     bool      `json:"used" db:"used"`
}

// WebAuthnCredential represents a stored WebAuthn credential.
type WebAuthnCredential struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	CredentialID []byte    `json:"-" db:"credential_id"`
	PublicKey    []byte    `json:"-" db:"public_key"`
	SignCount    uint32    `json:"sign_count" db:"sign_count"`
	AAGUID       []byte    `json:"-" db:"aaguid"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Webhook represents a configured webhook endpoint.
type Webhook struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	URL       string    `json:"url" db:"url" validate:"required,url"`
	Events    []string  `json:"events" db:"events"`
	Secret    string    `json:"-" db:"secret"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Action represents a custom action/hook in the pipeline.
type Action struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Trigger   string    `json:"trigger" db:"trigger" validate:"required"`
	Name      string    `json:"name" db:"name" validate:"required"`
	Code      string    `json:"code" db:"code"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	Order     int       `json:"order" db:"order"`
	TimeoutMs int       `json:"timeout_ms" db:"timeout_ms"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// EmailTemplate represents a tenant-specific email template.
type EmailTemplate struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Type      string    `json:"type" db:"type" validate:"required"`
	Locale    string    `json:"locale" db:"locale"`
	Subject   string    `json:"subject" db:"subject" validate:"required"`
	BodyMJML  string    `json:"body_mjml,omitempty" db:"body_mjml"`
	BodyHTML  string    `json:"body_html" db:"body_html" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents a tenant API key for programmatic access.
type APIKey struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	TenantID  uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name      string     `json:"name" db:"name" validate:"required,min=1,max=255"`
	KeyPrefix string     `json:"key_prefix" db:"key_prefix"`
	KeyHash   string     `json:"-" db:"key_hash"`
	Scopes    []string   `json:"scopes" db:"scopes"`
	RateLimit int        `json:"rate_limit" db:"rate_limit"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// FGATuple represents a fine-grained authorization tuple (Zanzibar model).
type FGATuple struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TenantID   uuid.UUID `json:"tenant_id" db:"tenant_id"`
	UserType   string    `json:"user_type" db:"user_type" validate:"required"`
	UserID     string    `json:"user_id" db:"user_id" validate:"required"`
	Relation   string    `json:"relation" db:"relation" validate:"required"`
	ObjectType string    `json:"object_type" db:"object_type" validate:"required"`
	ObjectID   string    `json:"object_id" db:"object_id" validate:"required"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// PasswordHistory stores previous password hashes for password reuse prevention.
type PasswordHistory struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Hash      string    `json:"-" db:"hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// --- Request/Response types ---

// PaginationParams holds pagination parameters.
type PaginationParams struct {
	Page    int `json:"page" validate:"min=1"`
	PerPage int `json:"per_page" validate:"min=1,max=100"`
}

// PaginatedResult wraps a paginated response.
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// TenantSettings holds configurable tenant-level settings.
type TenantSettings struct {
	PasswordMinLength     int      `json:"password_min_length" yaml:"password_min_length"`
	PasswordRequireUpper  bool     `json:"password_require_upper" yaml:"password_require_upper"`
	PasswordRequireLower  bool     `json:"password_require_lower" yaml:"password_require_lower"`
	PasswordRequireDigit  bool     `json:"password_require_digit" yaml:"password_require_digit"`
	PasswordRequireSymbol bool     `json:"password_require_symbol" yaml:"password_require_symbol"`
	PasswordHistoryCount  int      `json:"password_history_count" yaml:"password_history_count"`
	MFARequired           bool     `json:"mfa_required" yaml:"mfa_required"`
	AllowedMFAMethods     []string `json:"allowed_mfa_methods" yaml:"allowed_mfa_methods"`
	SessionDuration       int      `json:"session_duration_minutes" yaml:"session_duration_minutes"`
	InactivityTimeout     int      `json:"inactivity_timeout_minutes" yaml:"inactivity_timeout_minutes"`
	AllowedSignUpDomains  []string `json:"allowed_signup_domains" yaml:"allowed_signup_domains"`
	EnableSelfSignUp      bool     `json:"enable_self_signup" yaml:"enable_self_signup"`
}

// PageTemplate represents a tenant-specific page template for auth flows.
type PageTemplate struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	PageType    string    `json:"page_type" db:"page_type" validate:"required"`
	Name        string    `json:"name" db:"name" validate:"required"`
	HTMLContent string    `json:"html_content" db:"html_content"`
	CSSContent  string    `json:"css_content" db:"css_content"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// LanguageString represents a tenant-specific localized string.
type LanguageString struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	StringKey string    `json:"string_key" db:"string_key" validate:"required"`
	Locale    string    `json:"locale" db:"locale"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DeviceCode represents an RFC 8628 device authorization grant.
type DeviceCode struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	DeviceCode   string     `json:"-" db:"device_code"`
	UserCode     string     `json:"user_code" db:"user_code"`
	ClientID     string     `json:"client_id" db:"client_id"`
	Scopes       []string   `json:"scopes" db:"scopes"`
	Status       string     `json:"status" db:"status"`
	UserID       *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	PollInterval int        `json:"interval" db:"poll_interval"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// DefaultTenantSettings returns sane defaults.
func DefaultTenantSettings() TenantSettings {
	return TenantSettings{
		PasswordMinLength:     8,
		PasswordRequireUpper:  true,
		PasswordRequireLower:  true,
		PasswordRequireDigit:  true,
		PasswordRequireSymbol: false,
		PasswordHistoryCount:  5,
		MFARequired:           false,
		AllowedMFAMethods:     []string{"totp", "email", "webauthn"},
		SessionDuration:       1440,
		InactivityTimeout:     60,
		EnableSelfSignUp:      true,
	}
}
