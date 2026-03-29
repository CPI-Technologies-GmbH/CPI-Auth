-- CPI Auth IAM Platform - Initial Schema Rollback
-- 001_initial_schema.down.sql
-- Drops all tables in reverse dependency order

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_sso_connections_updated_at ON sso_connections;
DROP TRIGGER IF EXISTS trg_social_providers_updated_at ON social_providers;
DROP TRIGGER IF EXISTS trg_branding_configs_updated_at ON branding_configs;
DROP TRIGGER IF EXISTS trg_email_templates_updated_at ON email_templates;
DROP TRIGGER IF EXISTS trg_actions_updated_at ON actions;
DROP TRIGGER IF EXISTS trg_webhooks_updated_at ON webhooks;
DROP TRIGGER IF EXISTS trg_organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS trg_applications_updated_at ON applications;
DROP TRIGGER IF EXISTS trg_identities_updated_at ON identities;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP TRIGGER IF EXISTS trg_tenants_updated_at ON tenants;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at();

-- Drop tables in reverse dependency order
-- (leaf tables first, then tables with foreign key references)

-- Registration flows (depends on: tenants)
DROP TABLE IF EXISTS registration_flows;

-- Consent records (depends on: users, applications)
DROP TABLE IF EXISTS consent_records;

-- Branding configs (depends on: tenants)
DROP TABLE IF EXISTS branding_configs;

-- Social providers (depends on: tenants)
DROP TABLE IF EXISTS social_providers;

-- JWKS keys (no foreign keys)
DROP TABLE IF EXISTS jwks_keys;

-- Login attempts (no foreign keys)
DROP TABLE IF EXISTS login_attempts;

-- Password history (depends on: users)
DROP TABLE IF EXISTS password_history;

-- SCIM tokens (depends on: organizations)
DROP TABLE IF EXISTS scim_tokens;

-- SSO connections (depends on: organizations, tenants)
DROP TABLE IF EXISTS sso_connections;

-- FGA tuples (depends on: tenants)
DROP TABLE IF EXISTS fga_tuples;

-- Audit logs partition first, then parent
DROP TABLE IF EXISTS audit_logs_default;
DROP TABLE IF EXISTS audit_logs;

-- API keys (depends on: tenants)
DROP TABLE IF EXISTS api_keys;

-- Email templates (depends on: tenants)
DROP TABLE IF EXISTS email_templates;

-- Actions (depends on: tenants)
DROP TABLE IF EXISTS actions;

-- Webhooks (depends on: tenants)
DROP TABLE IF EXISTS webhooks;

-- WebAuthn credentials (depends on: users)
DROP TABLE IF EXISTS webauthn_credentials;

-- Recovery codes (depends on: users)
DROP TABLE IF EXISTS recovery_codes;

-- MFA enrollments (depends on: users)
DROP TABLE IF EXISTS mfa_enrollments;

-- Refresh tokens (depends on: users, applications, tenants)
DROP TABLE IF EXISTS refresh_tokens;

-- OAuth grants (depends on: users, applications, tenants)
DROP TABLE IF EXISTS oauth_grants;

-- Sessions (depends on: users, tenants)
DROP TABLE IF EXISTS sessions;

-- Organization members (depends on: organizations, users)
DROP TABLE IF EXISTS organization_members;

-- User roles (depends on: users, roles, organizations)
DROP TABLE IF EXISTS user_roles;

-- Roles (depends on: tenants, self-referencing)
DROP TABLE IF EXISTS roles;

-- Organizations (depends on: tenants)
DROP TABLE IF EXISTS organizations;

-- Applications (depends on: tenants)
DROP TABLE IF EXISTS applications;

-- Identities (depends on: users)
DROP TABLE IF EXISTS identities;

-- Users (depends on: tenants)
DROP TABLE IF EXISTS users;

-- Tenants (self-referencing parent_id)
DROP TABLE IF EXISTS tenants;

-- Drop extensions (optional, commented out to avoid affecting other databases)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
