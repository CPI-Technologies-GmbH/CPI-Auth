-- CPI Auth IAM Platform - Seed Data
-- 002_seed_data.up.sql

-- ============================================================
-- Use fixed UUIDs for seed data so we can reference them
-- and cleanly reverse in the down migration
-- ============================================================

-- Default root tenant
INSERT INTO tenants (id, name, slug, domain, settings, branding) VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Default Tenant',
    'default',
    'localhost',
    '{
        "session_lifetime_seconds": 86400,
        "idle_session_lifetime_seconds": 1800,
        "enable_signup": true,
        "enable_social_login": true,
        "enable_mfa": true,
        "mfa_policy": "optional",
        "password_policy": {
            "min_length": 8,
            "require_uppercase": true,
            "require_lowercase": true,
            "require_number": true,
            "require_special": false,
            "max_history": 5
        },
        "brute_force_protection": {
            "enabled": true,
            "max_attempts": 10,
            "block_duration_seconds": 900
        },
        "rate_limiting": {
            "login_per_minute": 20,
            "signup_per_minute": 5,
            "api_per_minute": 1000
        }
    }',
    '{
        "company_name": "CPI Auth",
        "support_email": "support@cpi-auth.local",
        "support_url": "https://docs.cpi-auth.local"
    }'
);

-- ============================================================
-- Default roles with hierarchical permissions
-- ============================================================

-- Admin role (full access)
INSERT INTO roles (id, tenant_id, name, description, permissions, is_system) VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'admin',
    'Full administrative access to all resources and settings',
    ARRAY[
        '*', 'admin:access',
        'users:read', 'users:write', 'users:delete',
        'roles:read', 'roles:write', 'roles:delete',
        'applications:read', 'applications:write', 'applications:delete',
        'organizations:read', 'organizations:write', 'organizations:delete',
        'webhooks:read', 'webhooks:write', 'webhooks:delete',
        'actions:read', 'actions:write', 'actions:delete',
        'email_templates:read', 'email_templates:write', 'email_templates:delete',
        'api_keys:read', 'api_keys:write', 'api_keys:delete',
        'audit_logs:read',
        'branding:read', 'branding:write',
        'sso:read', 'sso:write', 'sso:delete',
        'tenant:read', 'tenant:write',
        'fga:read', 'fga:write', 'fga:delete',
        'social_providers:read', 'social_providers:write', 'social_providers:delete',
        'scim:read', 'scim:write'
    ],
    TRUE
);

-- Manager role (manage users and organizations, no system config)
INSERT INTO roles (id, tenant_id, name, description, permissions, parent_role_id, is_system) VALUES (
    'b0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000001',
    'manager',
    'Manage users, organizations, and view audit logs',
    ARRAY[
        'users:read', 'users:write',
        'roles:read',
        'organizations:read', 'organizations:write',
        'audit_logs:read',
        'branding:read',
        'applications:read'
    ],
    'b0000000-0000-0000-0000-000000000001',
    TRUE
);

-- Editor role (edit content and templates)
INSERT INTO roles (id, tenant_id, name, description, permissions, parent_role_id, is_system) VALUES (
    'b0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000001',
    'editor',
    'Edit email templates, branding, and view users',
    ARRAY[
        'users:read',
        'email_templates:read', 'email_templates:write',
        'branding:read', 'branding:write',
        'organizations:read',
        'applications:read'
    ],
    'b0000000-0000-0000-0000-000000000002',
    TRUE
);

-- Viewer role (read-only)
INSERT INTO roles (id, tenant_id, name, description, permissions, parent_role_id, is_system) VALUES (
    'b0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000001',
    'viewer',
    'Read-only access to users, organizations, and logs',
    ARRAY[
        'users:read',
        'organizations:read',
        'applications:read',
        'audit_logs:read',
        'branding:read'
    ],
    'b0000000-0000-0000-0000-000000000003',
    TRUE
);

-- ============================================================
-- Default admin user
-- Password: admin123!
-- Hash generated with bcrypt (cost 10)
-- In production, change this password immediately!
-- ============================================================
INSERT INTO users (id, tenant_id, email, email_verified, name, password_hash, status, app_metadata) VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'admin@cpi-auth.local',
    TRUE,
    'System Administrator',
    -- bcrypt hash of 'admin123!' -- $2a$10$ prefix, cost factor 10
    '$2a$10$rQEY0tEMG9jqJMBXVBPOHOVbQWBzRhGVQF1U3YCKbPaGHCdWzVDWi',
    'active',
    '{"roles": ["admin"], "is_system_admin": true}'
);

-- NOTE: the admin role assignment was originally inserted here with a
-- placeholder organization_id, but the FK on user_roles.organization_id
-- means a fresh install fails this INSERT (no such org exists). Migration
-- 014 makes organization_id nullable, drops the FK, and inserts the
-- assignment idempotently with ON CONFLICT DO NOTHING. Existing
-- installations where this insert previously succeeded are unaffected.

-- ============================================================
-- Default application
-- ============================================================
INSERT INTO applications (id, tenant_id, name, type, client_id, client_secret_hash, redirect_uris, allowed_origins, post_logout_redirect_uris, grant_types, response_types, scopes, settings) VALUES (
    'd0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'Default App',
    'web',
    'authforge_default_app',
    -- placeholder hash for client secret 'authforge_default_secret'
    '$2a$10$kIzKhx7G0hN8ZJRQX5YCxOQqN0A4GfW7zDVVLwR8IYaH8jWyCLgXe',
    ARRAY['http://localhost:3000/callback', 'http://localhost:3000/auth/callback'],
    ARRAY['http://localhost:3000'],
    ARRAY['http://localhost:3000'],
    ARRAY['authorization_code', 'refresh_token', 'client_credentials'],
    ARRAY['code'],
    ARRAY['openid', 'profile', 'email', 'offline_access'],
    '{
        "logo_uri": null,
        "tos_uri": null,
        "policy_uri": null,
        "allowed_web_origins": ["http://localhost:3000"],
        "allowed_logout_urls": ["http://localhost:3000"],
        "jwt_configuration": {
            "alg": "RS256",
            "lifetime_in_seconds": 3600
        }
    }'
);

-- ============================================================
-- Default email templates
-- ============================================================

-- Email Verification
INSERT INTO email_templates (id, tenant_id, type, locale, subject, body_html, variables) VALUES (
    'e0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'verification',
    'en',
    'Verify your email address',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify your email</title>
    <style>
        body { margin: 0; padding: 0; font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: #F9FAFB; }
        .container { max-width: 560px; margin: 0 auto; padding: 40px 20px; }
        .card { background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 32px; }
        .logo img { height: 40px; }
        h1 { color: #1F2937; font-size: 24px; font-weight: 600; margin: 0 0 16px 0; text-align: center; }
        p { color: #6B7280; font-size: 16px; line-height: 24px; margin: 0 0 24px 0; text-align: center; }
        .button { display: block; width: 100%; padding: 14px 24px; background-color: #6366F1; color: #FFFFFF; text-decoration: none; border-radius: 8px; font-size: 16px; font-weight: 600; text-align: center; box-sizing: border-box; }
        .button:hover { background-color: #4F46E5; }
        .footer { text-align: center; margin-top: 32px; color: #9CA3AF; font-size: 14px; line-height: 20px; }
        .link-text { color: #6B7280; font-size: 14px; word-break: break-all; text-align: center; margin-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">
                <img src="{{logo_url}}" alt="{{company_name}}" />
            </div>
            <h1>Verify your email address</h1>
            <p>Hi {{user_name}},</p>
            <p>Thanks for signing up! Please verify your email address by clicking the button below.</p>
            <a href="{{verification_url}}" class="button">Verify Email Address</a>
            <p class="link-text">Or copy and paste this link: {{verification_url}}</p>
            <p class="footer">This link will expire in {{expiry_hours}} hours. If you did not create an account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; {{current_year}} {{company_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>',
    '{"user_name": "string", "verification_url": "string", "expiry_hours": "number", "company_name": "string", "logo_url": "string", "current_year": "string"}'
);

-- Password Reset
INSERT INTO email_templates (id, tenant_id, type, locale, subject, body_html, variables) VALUES (
    'e0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000001',
    'password_reset',
    'en',
    'Reset your password',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset your password</title>
    <style>
        body { margin: 0; padding: 0; font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: #F9FAFB; }
        .container { max-width: 560px; margin: 0 auto; padding: 40px 20px; }
        .card { background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 32px; }
        .logo img { height: 40px; }
        h1 { color: #1F2937; font-size: 24px; font-weight: 600; margin: 0 0 16px 0; text-align: center; }
        p { color: #6B7280; font-size: 16px; line-height: 24px; margin: 0 0 24px 0; text-align: center; }
        .button { display: block; width: 100%; padding: 14px 24px; background-color: #6366F1; color: #FFFFFF; text-decoration: none; border-radius: 8px; font-size: 16px; font-weight: 600; text-align: center; box-sizing: border-box; }
        .button:hover { background-color: #4F46E5; }
        .footer { text-align: center; margin-top: 32px; color: #9CA3AF; font-size: 14px; line-height: 20px; }
        .warning { background-color: #FEF3C7; border: 1px solid #FCD34D; border-radius: 8px; padding: 12px 16px; margin: 24px 0; }
        .warning p { color: #92400E; font-size: 14px; margin: 0; }
        .link-text { color: #6B7280; font-size: 14px; word-break: break-all; text-align: center; margin-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">
                <img src="{{logo_url}}" alt="{{company_name}}" />
            </div>
            <h1>Reset your password</h1>
            <p>Hi {{user_name}},</p>
            <p>We received a request to reset the password for your account. Click the button below to set a new password.</p>
            <a href="{{reset_url}}" class="button">Reset Password</a>
            <p class="link-text">Or copy and paste this link: {{reset_url}}</p>
            <div class="warning">
                <p>If you did not request a password reset, please ignore this email or contact support if you have concerns about your account security.</p>
            </div>
            <p class="footer">This link will expire in {{expiry_minutes}} minutes.</p>
        </div>
        <div class="footer">
            <p>&copy; {{current_year}} {{company_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>',
    '{"user_name": "string", "reset_url": "string", "expiry_minutes": "number", "company_name": "string", "logo_url": "string", "current_year": "string"}'
);

-- Welcome Email
INSERT INTO email_templates (id, tenant_id, type, locale, subject, body_html, variables) VALUES (
    'e0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000001',
    'welcome',
    'en',
    'Welcome to {{company_name}}!',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome</title>
    <style>
        body { margin: 0; padding: 0; font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: #F9FAFB; }
        .container { max-width: 560px; margin: 0 auto; padding: 40px 20px; }
        .card { background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 32px; }
        .logo img { height: 40px; }
        h1 { color: #1F2937; font-size: 24px; font-weight: 600; margin: 0 0 16px 0; text-align: center; }
        p { color: #6B7280; font-size: 16px; line-height: 24px; margin: 0 0 24px 0; text-align: center; }
        .button { display: block; width: 100%; padding: 14px 24px; background-color: #6366F1; color: #FFFFFF; text-decoration: none; border-radius: 8px; font-size: 16px; font-weight: 600; text-align: center; box-sizing: border-box; }
        .button:hover { background-color: #4F46E5; }
        .features { margin: 24px 0; padding: 0; }
        .feature { display: flex; align-items: flex-start; margin-bottom: 16px; }
        .feature-icon { width: 24px; height: 24px; margin-right: 12px; color: #6366F1; flex-shrink: 0; font-size: 18px; }
        .feature-text { color: #4B5563; font-size: 15px; line-height: 22px; }
        .footer { text-align: center; margin-top: 32px; color: #9CA3AF; font-size: 14px; line-height: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">
                <img src="{{logo_url}}" alt="{{company_name}}" />
            </div>
            <h1>Welcome to {{company_name}}!</h1>
            <p>Hi {{user_name}},</p>
            <p>Your account has been created successfully. We are excited to have you on board!</p>
            <a href="{{dashboard_url}}" class="button">Go to Dashboard</a>
            <p style="color: #9CA3AF; font-size: 14px; text-align: center; margin-top: 24px;">If you have any questions, feel free to reach out to our support team at {{support_email}}.</p>
        </div>
        <div class="footer">
            <p>&copy; {{current_year}} {{company_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>',
    '{"user_name": "string", "dashboard_url": "string", "company_name": "string", "logo_url": "string", "support_email": "string", "current_year": "string"}'
);

-- Magic Link
INSERT INTO email_templates (id, tenant_id, type, locale, subject, body_html, variables) VALUES (
    'e0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000001',
    'magic_link',
    'en',
    'Your sign-in link for {{company_name}}',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign in</title>
    <style>
        body { margin: 0; padding: 0; font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: #F9FAFB; }
        .container { max-width: 560px; margin: 0 auto; padding: 40px 20px; }
        .card { background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 32px; }
        .logo img { height: 40px; }
        h1 { color: #1F2937; font-size: 24px; font-weight: 600; margin: 0 0 16px 0; text-align: center; }
        p { color: #6B7280; font-size: 16px; line-height: 24px; margin: 0 0 24px 0; text-align: center; }
        .button { display: block; width: 100%; padding: 14px 24px; background-color: #6366F1; color: #FFFFFF; text-decoration: none; border-radius: 8px; font-size: 16px; font-weight: 600; text-align: center; box-sizing: border-box; }
        .button:hover { background-color: #4F46E5; }
        .footer { text-align: center; margin-top: 32px; color: #9CA3AF; font-size: 14px; line-height: 20px; }
        .security-note { background-color: #EEF2FF; border-radius: 8px; padding: 16px; margin: 24px 0; }
        .security-note p { color: #4338CA; font-size: 14px; margin: 0; }
        .link-text { color: #6B7280; font-size: 14px; word-break: break-all; text-align: center; margin-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">
                <img src="{{logo_url}}" alt="{{company_name}}" />
            </div>
            <h1>Sign in to {{company_name}}</h1>
            <p>Hi {{user_name}},</p>
            <p>Click the button below to securely sign in to your account. No password needed!</p>
            <a href="{{magic_link_url}}" class="button">Sign In</a>
            <p class="link-text">Or copy and paste this link: {{magic_link_url}}</p>
            <div class="security-note">
                <p>This link is valid for {{expiry_minutes}} minutes and can only be used once.</p>
            </div>
            <p class="footer">If you did not request this link, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; {{current_year}} {{company_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>',
    '{"user_name": "string", "magic_link_url": "string", "expiry_minutes": "number", "company_name": "string", "logo_url": "string", "current_year": "string"}'
);

-- Invitation Email
INSERT INTO email_templates (id, tenant_id, type, locale, subject, body_html, variables) VALUES (
    'e0000000-0000-0000-0000-000000000005',
    'a0000000-0000-0000-0000-000000000001',
    'invitation',
    'en',
    'You have been invited to join {{organization_name}} on {{company_name}}',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invitation</title>
    <style>
        body { margin: 0; padding: 0; font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: #F9FAFB; }
        .container { max-width: 560px; margin: 0 auto; padding: 40px 20px; }
        .card { background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 32px; }
        .logo img { height: 40px; }
        h1 { color: #1F2937; font-size: 24px; font-weight: 600; margin: 0 0 16px 0; text-align: center; }
        p { color: #6B7280; font-size: 16px; line-height: 24px; margin: 0 0 24px 0; text-align: center; }
        .button { display: block; width: 100%; padding: 14px 24px; background-color: #6366F1; color: #FFFFFF; text-decoration: none; border-radius: 8px; font-size: 16px; font-weight: 600; text-align: center; box-sizing: border-box; }
        .button:hover { background-color: #4F46E5; }
        .footer { text-align: center; margin-top: 32px; color: #9CA3AF; font-size: 14px; line-height: 20px; }
        .inviter-info { background-color: #F3F4F6; border-radius: 8px; padding: 16px; margin: 24px 0; text-align: center; }
        .inviter-info .name { color: #1F2937; font-weight: 600; font-size: 16px; }
        .inviter-info .role { color: #6B7280; font-size: 14px; }
        .link-text { color: #6B7280; font-size: 14px; word-break: break-all; text-align: center; margin-top: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">
                <img src="{{logo_url}}" alt="{{company_name}}" />
            </div>
            <h1>You are invited!</h1>
            <p>{{inviter_name}} has invited you to join <strong>{{organization_name}}</strong> on {{company_name}}.</p>
            <div class="inviter-info">
                <div class="name">{{inviter_name}}</div>
                <div class="role">invited you as {{role}}</div>
            </div>
            <a href="{{invitation_url}}" class="button">Accept Invitation</a>
            <p class="link-text">Or copy and paste this link: {{invitation_url}}</p>
            <p class="footer">This invitation will expire in {{expiry_days}} days. If you were not expecting this invitation, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; {{current_year}} {{company_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>',
    '{"inviter_name": "string", "organization_name": "string", "invitation_url": "string", "role": "string", "expiry_days": "number", "company_name": "string", "logo_url": "string", "current_year": "string"}'
);

-- MFA Code Email
INSERT INTO email_templates (id, tenant_id, type, locale, subject, body_html, variables) VALUES (
    'e0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000001',
    'mfa',
    'en',
    'Your verification code for {{company_name}}',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verification Code</title>
    <style>
        body { margin: 0; padding: 0; font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: #F9FAFB; }
        .container { max-width: 560px; margin: 0 auto; padding: 40px 20px; }
        .card { background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 32px; }
        .logo img { height: 40px; }
        h1 { color: #1F2937; font-size: 24px; font-weight: 600; margin: 0 0 16px 0; text-align: center; }
        p { color: #6B7280; font-size: 16px; line-height: 24px; margin: 0 0 24px 0; text-align: center; }
        .code-container { text-align: center; margin: 32px 0; }
        .code { display: inline-block; font-size: 36px; font-weight: 700; letter-spacing: 8px; color: #1F2937; background-color: #F3F4F6; padding: 16px 32px; border-radius: 12px; font-family: "SF Mono", "Fira Code", "Fira Mono", "Roboto Mono", monospace; }
        .footer { text-align: center; margin-top: 32px; color: #9CA3AF; font-size: 14px; line-height: 20px; }
        .warning { background-color: #FEF3C7; border: 1px solid #FCD34D; border-radius: 8px; padding: 12px 16px; margin: 24px 0; }
        .warning p { color: #92400E; font-size: 14px; margin: 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="logo">
                <img src="{{logo_url}}" alt="{{company_name}}" />
            </div>
            <h1>Verification Code</h1>
            <p>Hi {{user_name}},</p>
            <p>Use the following code to complete your sign-in. This code is valid for {{expiry_minutes}} minutes.</p>
            <div class="code-container">
                <span class="code">{{code}}</span>
            </div>
            <div class="warning">
                <p>Never share this code with anyone. {{company_name}} will never ask you for this code via phone or chat.</p>
            </div>
            <p class="footer">If you did not attempt to sign in, please change your password immediately and contact support.</p>
        </div>
        <div class="footer">
            <p>&copy; {{current_year}} {{company_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>',
    '{"user_name": "string", "code": "string", "expiry_minutes": "number", "company_name": "string", "logo_url": "string", "current_year": "string"}'
);

-- ============================================================
-- Default branding config
-- ============================================================
INSERT INTO branding_configs (id, tenant_id, primary_color, secondary_color, background_color, text_color, error_color, success_color, font_family, border_radius, layout) VALUES (
    'f0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    '#6366F1',
    '#8B5CF6',
    '#FFFFFF',
    '#1F2937',
    '#EF4444',
    '#10B981',
    'Inter',
    8,
    'centered'
);

-- ============================================================
-- Initial JWKS key pair placeholder
-- NOTE: In production, real RSA keys must be generated by the
-- application on first boot. These are placeholder values.
-- The private_key_encrypted field would normally contain an
-- AES-256-GCM encrypted PKCS#8 DER-encoded RSA private key.
-- ============================================================
INSERT INTO jwks_keys (id, kid, algorithm, public_key, private_key_encrypted, is_current) VALUES (
    'aa000000-0000-0000-0000-000000000001',
    'cpi-auth-initial-key-001',
    'RS256',
    '-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0Z3VS5JJcds3xfn/ygWe
PLACEHOLDER_PUBLIC_KEY_REPLACE_ON_FIRST_BOOT_WITH_REAL_RSA2048
RealKeyWillBeGeneratedByTheApplicationOnFirstStartup00000000
ThisIsAPlaceholderAndMustBeReplacedBeforeProductionUse00000
AQIDAQAB
-----END PUBLIC KEY-----',
    -- Placeholder encrypted private key (must be replaced on first boot)
    -- In production, the app generates a real RSA-2048 keypair and encrypts
    -- the private key with the ENCRYPTION_KEY environment variable using AES-256-GCM
    E'\\x504C414345484F4C4445525F454E435259505445445F505249564154455F4B45595F5245504C4143455F4F4E5F46495253545F424F4F54',
    TRUE
);

-- ============================================================
-- Seed audit log entry for the initial setup
-- ============================================================
INSERT INTO audit_logs (tenant_id, actor_id, actor_type, action, target_type, target_id, metadata) VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'c0000000-0000-0000-0000-000000000001',
    'system',
    'system.seed_data_applied',
    'tenant',
    'a0000000-0000-0000-0000-000000000001',
    '{"version": "002", "description": "Initial seed data applied during setup"}'
);
