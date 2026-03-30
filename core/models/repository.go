package models

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines storage operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams, search string) (*PaginatedResult[User], error)
	Block(ctx context.Context, tenantID, id uuid.UUID) error
	Unblock(ctx context.Context, tenantID, id uuid.UUID) error
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	GetPasswordHistory(ctx context.Context, userID uuid.UUID, limit int) ([]PasswordHistory, error)
	AddPasswordHistory(ctx context.Context, entry *PasswordHistory) error
}

// TenantRepository defines storage operations for tenants.
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params PaginationParams) (*PaginatedResult[Tenant], error)
}

// ApplicationRepository defines storage operations for OAuth applications.
type ApplicationRepository interface {
	Create(ctx context.Context, app *Application) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Application, error)
	GetByClientID(ctx context.Context, clientID string) (*Application, error)
	Update(ctx context.Context, app *Application) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[Application], error)
}

// IdentityRepository defines storage operations for external identities.
type IdentityRepository interface {
	Create(ctx context.Context, identity *Identity) error
	GetByID(ctx context.Context, id uuid.UUID) (*Identity, error)
	GetByProvider(ctx context.Context, provider, providerUserID string) (*Identity, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Identity, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, identity *Identity) error
}

// OrganizationRepository defines storage operations for organizations.
type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Organization, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[Organization], error)
	AddMember(ctx context.Context, member *OrganizationMember) error
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error
	ListMembers(ctx context.Context, orgID uuid.UUID, params PaginationParams) (*PaginatedResult[OrganizationMember], error)
	GetMember(ctx context.Context, orgID, userID uuid.UUID) (*OrganizationMember, error)
}

// RoleRepository defines storage operations for roles.
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[Role], error)
	GetRolesForUser(ctx context.Context, userID uuid.UUID) ([]Role, error)
	GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID, organizationID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID, organizationID uuid.UUID) error
}

// PermissionRepository defines storage operations for dynamic permissions.
type PermissionRepository interface {
	Create(ctx context.Context, perm *Permission) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Permission, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*Permission, error)
	Update(ctx context.Context, perm *Permission) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[Permission], error)
	ListAll(ctx context.Context, tenantID uuid.UUID) ([]Permission, error)
	EnsureSystemDefaults(ctx context.Context, tenantID uuid.UUID) error
}

// ApplicationPermissionRepository defines storage operations for application permission whitelists.
type ApplicationPermissionRepository interface {
	SetPermissions(ctx context.Context, appID, tenantID uuid.UUID, permissions []string) error
	GetPermissions(ctx context.Context, appID uuid.UUID) ([]string, error)
}

// SessionRepository defines storage operations for sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Session, error)
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	DeleteByTenant(ctx context.Context, tenantID uuid.UUID) error
}

// OAuthGrantRepository defines storage operations for authorization grants.
type OAuthGrantRepository interface {
	Create(ctx context.Context, grant *OAuthGrant) error
	GetByCode(ctx context.Context, code string) (*OAuthGrant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// RefreshTokenRepository defines storage operations for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByFamily(ctx context.Context, family string) error
	RevokeByUser(ctx context.Context, userID uuid.UUID) error
	RevokeByApplication(ctx context.Context, appID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// MFAEnrollmentRepository defines storage operations for MFA enrollments.
type MFAEnrollmentRepository interface {
	Create(ctx context.Context, enrollment *MFAEnrollment) error
	GetByID(ctx context.Context, id uuid.UUID) (*MFAEnrollment, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]MFAEnrollment, error)
	Update(ctx context.Context, enrollment *MFAEnrollment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RecoveryCodeRepository defines storage operations for recovery codes.
type RecoveryCodeRepository interface {
	Create(ctx context.Context, code *RecoveryCode) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]RecoveryCode, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	GetByUserAndHash(ctx context.Context, userID uuid.UUID, codeHash string) (*RecoveryCode, error)
}

// WebAuthnCredentialRepository defines storage operations for WebAuthn credentials.
type WebAuthnCredentialRepository interface {
	Create(ctx context.Context, cred *WebAuthnCredential) error
	GetByID(ctx context.Context, id uuid.UUID) (*WebAuthnCredential, error)
	GetByCredentialID(ctx context.Context, credentialID []byte) (*WebAuthnCredential, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]WebAuthnCredential, error)
	Update(ctx context.Context, cred *WebAuthnCredential) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AuditLogRepository defines storage operations for audit logs.
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams, action string) (*PaginatedResult[AuditLog], error)
}

// WebhookRepository defines storage operations for webhooks.
type WebhookRepository interface {
	Create(ctx context.Context, webhook *Webhook) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Webhook, error)
	Update(ctx context.Context, webhook *Webhook) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[Webhook], error)
	ListByEvent(ctx context.Context, tenantID uuid.UUID, event string) ([]Webhook, error)
}

// ActionRepository defines storage operations for actions.
type ActionRepository interface {
	Create(ctx context.Context, action *Action) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Action, error)
	Update(ctx context.Context, action *Action) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[Action], error)
	ListByTrigger(ctx context.Context, tenantID uuid.UUID, trigger string) ([]Action, error)
}

// EmailTemplateRepository defines storage operations for email templates.
type EmailTemplateRepository interface {
	Create(ctx context.Context, tmpl *EmailTemplate) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*EmailTemplate, error)
	GetByTypeAndLocale(ctx context.Context, tenantID uuid.UUID, typ, locale string) (*EmailTemplate, error)
	Update(ctx context.Context, tmpl *EmailTemplate) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[EmailTemplate], error)
}

// APIKeyRepository defines storage operations for API keys.
type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)
	Update(ctx context.Context, key *APIKey) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, params PaginationParams) (*PaginatedResult[APIKey], error)
}

// FGATupleRepository defines storage operations for FGA tuples.
type FGATupleRepository interface {
	Create(ctx context.Context, tuple *FGATuple) error
	Delete(ctx context.Context, id uuid.UUID) error
	Check(ctx context.Context, tenantID uuid.UUID, userType, userID, relation, objectType, objectID string) (bool, error)
	ListByObject(ctx context.Context, tenantID uuid.UUID, objectType, objectID string) ([]FGATuple, error)
	ListByUser(ctx context.Context, tenantID uuid.UUID, userType, userID string) ([]FGATuple, error)
}

// CustomFieldDefinitionRepository defines storage operations for custom field definitions.
type CustomFieldDefinitionRepository interface {
	Create(ctx context.Context, field *CustomFieldDefinition) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*CustomFieldDefinition, error)
	Update(ctx context.Context, field *CustomFieldDefinition) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID) ([]CustomFieldDefinition, error)
}

// DomainVerificationRepository defines storage operations for domain verifications.
type DomainVerificationRepository interface {
	Create(ctx context.Context, dv *DomainVerification) error
	GetByID(ctx context.Context, id uuid.UUID) (*DomainVerification, error)
	GetByDomain(ctx context.Context, domain string) (*DomainVerification, error)
	GetByTenant(ctx context.Context, tenantID uuid.UUID) (*DomainVerification, error)
	Update(ctx context.Context, dv *DomainVerification) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PageTemplateRepository defines storage operations for page templates.
type PageTemplateRepository interface {
	Create(ctx context.Context, tmpl *PageTemplate) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*PageTemplate, error)
	GetByType(ctx context.Context, tenantID uuid.UUID, pageType string) (*PageTemplate, error)
	Update(ctx context.Context, tmpl *PageTemplate) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID) ([]PageTemplate, error)
	Duplicate(ctx context.Context, tenantID, sourceID uuid.UUID, newName string) (*PageTemplate, error)
}

// LanguageStringRepository defines storage operations for language strings.
type LanguageStringRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, locale string) ([]LanguageString, error)
	Upsert(ctx context.Context, ls *LanguageString) error
	Delete(ctx context.Context, tenantID uuid.UUID, stringKey, locale string) error
	GetByKeys(ctx context.Context, tenantID uuid.UUID, keys []string, locale string) (map[string]string, error)
}

// DeviceCodeRepository defines storage operations for RFC 8628 device authorization codes.
type DeviceCodeRepository interface {
	Create(ctx context.Context, dc *DeviceCode) error
	GetByDeviceCode(ctx context.Context, deviceCode string) (*DeviceCode, error)
	GetByUserCode(ctx context.Context, userCode string) (*DeviceCode, error)
	Authorize(ctx context.Context, userCode string, userID uuid.UUID) error
	Deny(ctx context.Context, userCode string) error
	DeleteExpired(ctx context.Context) error
}
