// ─── API Types ────────────────────────────────────────────────

export type PageType =
  | 'login' | 'signup' | 'verification' | 'password_reset'
  | 'mfa_challenge' | 'error' | 'consent' | 'profile' | 'custom'

export interface PageTemplate {
  id: string
  tenant_id: string
  page_type: PageType
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

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
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

// ─── SDK Config ───────────────────────────────────────────────

export interface SDKConfig {
  server: string
  tenantId?: string
  token?: string
  credentials?: { email: string; password: string }
}

// ─── Design Tokens ────────────────────────────────────────────

export interface DesignTokens {
  colors?: Record<string, string>
  spacing?: Record<string, string>
  radius?: Record<string, string>
  typography?: Record<string, string>
}

// ─── Project Config (cpi-auth.config.yaml) ───────────────────

export interface CustomFieldConfig {
  label: string
  type: 'text' | 'email' | 'tel' | 'number' | 'select' | 'checkbox' | 'date' | 'textarea' | 'url'
  required?: boolean
  placeholder?: string
  options?: string[]
}

export interface ProjectConfig {
  server: string
  tenant_id: string
  tokens?: DesignTokens
  templates?: Record<string, { html: string; css: string }>
  locales?: string[]
  preview?: {
    custom_fields?: CustomFieldConfig[]
    sample_data?: Record<string, string>
  }
}

// ─── Sync Types ───────────────────────────────────────────────

export interface ContrastResult {
  pair: [string, string]
  ratio: number
  aa: boolean
  aaLarge: boolean
  aaa: boolean
}

export interface SyncDiff {
  templates: {
    added: string[]
    modified: string[]
    unchanged: string[]
  }
  strings: {
    added: number
    modified: number
    deleted: number
    unchanged: number
  }
}
