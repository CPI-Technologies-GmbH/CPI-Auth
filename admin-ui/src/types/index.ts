// ==================== Auth ====================
export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface AdminUser {
  id: string
  email: string
  name: string
  avatar_url?: string
  role: 'super_admin' | 'admin' | 'viewer'
  tenant_id: string
  created_at: string
}

// ==================== Users ====================
export interface User {
  id: string
  email: string
  name: string
  avatar_url?: string
  status: 'active' | 'blocked' | 'inactive' | 'pending'
  email_verified: boolean
  phone?: string
  phone_verified?: boolean
  last_login?: string
  login_count: number
  created_at: string
  updated_at: string
  metadata?: Record<string, unknown>
  app_metadata?: Record<string, unknown>
}

export interface UserSession {
  id: string
  user_id: string
  ip_address: string
  user_agent: string
  created_at: string
  expires_at: string
  last_active: string
}

export interface UserMfaEnrollment {
  id: string
  user_id: string
  method: 'totp' | 'sms' | 'email' | 'webauthn'
  status: 'confirmed' | 'pending'
  created_at: string
}

export interface LinkedIdentity {
  provider: string
  provider_id: string
  email?: string
  name?: string
  linked_at: string
}

export interface AuditLogEntry {
  id: string
  action: string
  actor_id: string
  actor_email?: string
  target_type?: string
  target_id?: string
  ip_address?: string
  user_agent?: string
  metadata?: Record<string, unknown>
  tenant_id: string
  created_at: string
}

// ==================== Applications ====================
export type ApplicationType = 'spa' | 'web' | 'native' | 'm2m'
export type GrantType = 'authorization_code' | 'client_credentials' | 'refresh_token' | 'implicit' | 'password'

export interface Application {
  id: string
  name: string
  type: ApplicationType
  client_id: string
  client_secret?: string
  description?: string
  logo_url?: string
  redirect_uris: string[]
  allowed_origins: string[]
  allowed_logout_urls: string[]
  grant_types: GrantType[]
  token_endpoint_auth_method: string
  access_token_ttl: number
  refresh_token_ttl: number
  id_token_ttl: number
  is_active: boolean
  created_at: string
  updated_at: string
}

// ==================== Tenants ====================
export interface Tenant {
  id: string
  name: string
  slug: string
  domain?: string
  custom_domain?: string
  parent_id?: string
  logo_url?: string
  is_active: boolean
  created_at: string
  updated_at: string
  settings?: TenantSettings
}

export interface TenantSettings {
  branding?: BrandingConfig
  security?: SecuritySettings
  mfa?: MfaSettings
}

export interface BrandingConfig {
  primary_color: string
  secondary_color: string
  background_color: string
  text_color: string
  logo_url?: string
  logo_dark_url?: string
  font_family: string
  border_radius: number
  layout_mode: 'centered' | 'split-screen' | 'sidebar'
  custom_css?: string
}

export interface SecuritySettings {
  password_min_length: number
  password_require_uppercase: boolean
  password_require_lowercase: boolean
  password_require_numbers: boolean
  password_require_special: boolean
  brute_force_protection: boolean
  max_login_attempts: number
  lockout_duration: number
  session_lifetime: number
  session_idle_timeout: number
}

export interface MfaSettings {
  enabled: boolean
  required: boolean
  allowed_methods: ('totp' | 'sms' | 'email' | 'webauthn')[]
}

// ==================== Organizations ====================
export interface Organization {
  id: string
  name: string
  display_name: string
  logo_url?: string
  tenant_id: string
  member_count: number
  created_at: string
  updated_at: string
  metadata?: Record<string, unknown>
}

export interface OrganizationMember {
  id: string
  user_id: string
  user_email: string
  user_name: string
  role: string
  joined_at: string
}

// ==================== Roles & Permissions ====================
export interface Role {
  id: string
  name: string
  description?: string
  is_system: boolean
  parent_id?: string
  permissions: string[]
  tenant_id: string
  created_at: string
  updated_at: string
}

export interface Permission {
  id: string
  name: string
  display_name: string
  description?: string
  group: string
  is_system: boolean
  created_at: string
  updated_at: string
}

// ==================== Webhooks ====================
export type WebhookEvent =
  | 'user.created'
  | 'user.updated'
  | 'user.deleted'
  | 'user.login'
  | 'user.logout'
  | 'user.blocked'
  | 'user.password_changed'
  | 'user.mfa_enrolled'
  | 'application.created'
  | 'application.updated'
  | 'organization.created'
  | 'organization.member_added'
  | 'tenant.created'

export interface Webhook {
  id: string
  name: string
  url: string
  events: WebhookEvent[]
  secret: string
  is_active: boolean
  last_triggered?: string
  failure_count: number
  created_at: string
  updated_at: string
}

export interface WebhookDelivery {
  id: string
  webhook_id: string
  event: string
  status_code: number
  response_body?: string
  request_body: string
  duration_ms: number
  success: boolean
  attempts: number
  created_at: string
}

// ==================== Actions/Hooks ====================
export type ActionTrigger =
  | 'pre-login'
  | 'post-login'
  | 'pre-register'
  | 'post-register'
  | 'pre-token'
  | 'post-change-password'
  | 'pre-user-update'

export interface Action {
  id: string
  name: string
  trigger: ActionTrigger
  code: string
  is_active: boolean
  order: number
  runtime: string
  timeout_ms: number
  created_at: string
  updated_at: string
}

// ==================== Email Templates ====================
export type EmailTemplateType =
  | 'welcome'
  | 'verification'
  | 'password_reset'
  | 'mfa_code'
  | 'invitation'
  | 'blocked'
  | 'password_changed'

export interface EmailTemplate {
  id: string
  type: EmailTemplateType
  locale: string
  subject: string
  body_mjml: string
  body_html: string
  is_active: boolean
  created_at: string
  updated_at: string
}

// ==================== API Keys ====================
export interface ApiKey {
  id: string
  name: string
  key_prefix: string
  scopes: string[]
  rate_limit: number
  expires_at?: string
  last_used_at?: string
  is_active: boolean
  created_at: string
}

// ==================== Dashboard ====================
export interface DashboardMetrics {
  active_users: number
  active_users_change: number
  login_success_rate: number
  login_success_rate_change: number
  mfa_adoption: number
  mfa_adoption_change: number
  total_sessions: number
  total_sessions_change: number
  error_rate: number
  error_rate_change: number
}

export interface LoginChartData {
  date: string
  logins: number
  failures: number
}

export interface AuthMethodData {
  method: string
  count: number
}

export interface RecentEvent {
  id: string
  type: string
  description: string
  actor: string
  created_at: string
}

// ==================== Domain Verification ====================
export interface DomainVerification {
  id: string
  domain: string
  is_verified: boolean
  verification_method: string
  dns_record?: {
    record_type: string
    host: string
    value: string
  }
  status: 'none' | 'pending' | 'verified'
  verified_at?: string
  created_at?: string
}

// ==================== Custom Fields ====================
export type CustomFieldType = 'text' | 'number' | 'checkbox' | 'date' | 'select' | 'textarea' | 'url' | 'email' | 'tel'

export interface CustomFieldDefinition {
  id: string
  tenant_id: string
  name: string
  label: string
  field_type: CustomFieldType
  placeholder?: string
  description?: string
  options?: string[]
  required: boolean
  visible_on: 'registration' | 'profile' | 'both'
  position: number
  validation_rules?: Record<string, unknown>
  is_active: boolean
  created_at: string
  updated_at: string
}

// ==================== Page Templates ====================
export type PageTemplateType = 'login' | 'signup' | 'verification' | 'password_reset' | 'mfa_challenge' | 'error' | 'consent' | 'profile' | 'custom'

export interface PageTemplate {
  id: string
  tenant_id: string
  page_type: PageTemplateType
  name: string
  html_content: string
  css_content: string
  is_active: boolean
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface LanguageString {
  id: string
  tenant_id: string
  string_key: string
  locale: string
  value: string
  created_at: string
  updated_at: string
}

// ==================== API ====================
export interface PaginatedResponse<T> {
  data: T[]
  cursor?: string
  has_more: boolean
  total?: number
}

export interface ApiError {
  error: string
  message: string
  status_code: number
}
