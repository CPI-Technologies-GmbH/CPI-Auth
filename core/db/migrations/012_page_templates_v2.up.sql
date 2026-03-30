-- Page Templates v2: default templates, language strings, custom pages

-- 1. Add is_default flag and allow multiple templates per page_type
ALTER TABLE page_templates ADD COLUMN IF NOT EXISTS is_default BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE page_templates DROP CONSTRAINT IF EXISTS page_templates_tenant_id_page_type_key;

-- 2. Language strings table for template i18n
CREATE TABLE IF NOT EXISTS template_language_strings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    string_key VARCHAR(255) NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    value TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, string_key, locale)
);
CREATE INDEX IF NOT EXISTS idx_tls_tenant ON template_language_strings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tls_key ON template_language_strings(tenant_id, string_key);

-- 3. Seed default templates for the default tenant
INSERT INTO page_templates (id, tenant_id, page_type, name, html_content, css_content, is_active, is_default, created_at, updated_at)
VALUES
-- Login
('f0000001-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'login', 'Default Login',
'<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Sign In - {{application.name}}</title>
</head>
<body>
  <div class="card">
    <div class="logo">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
    </div>
    <h1>{{t.login.title}}</h1>
    <p class="subtitle">{{t.login.subtitle}}</p>
    <form>
      <div class="field">
        <label>{{t.login.email}}</label>
        <input type="email" placeholder="you@example.com" />
      </div>
      <div class="field">
        <label>{{t.login.password}}</label>
        <input type="password" placeholder="Enter your password" />
      </div>
      <button type="submit" class="btn-primary">{{t.login.submit}}</button>
      <a href="#" class="link">{{t.login.forgot_password}}</a>
    </form>
    <div class="divider"><span>or</span></div>
    <div class="social-buttons">
      <button class="btn-social">Continue with Google</button>
      <button class="btn-social">Continue with GitHub</button>
    </div>
    <p class="footer-text">{{t.login.no_account}} <a href="#">{{t.login.sign_up}}</a></p>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 420px; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
.logo { width: 48px; height: 48px; margin: 0 auto 1.5rem; color: #818cf8; }
h1 { text-align: center; font-size: 1.5rem; font-weight: 700; margin-bottom: .25rem; }
.subtitle { text-align: center; color: #94a3b8; font-size: .875rem; margin-bottom: 2rem; }
.field { margin-bottom: 1.25rem; }
.field label { display: block; font-size: .8125rem; font-weight: 500; margin-bottom: .5rem; color: #cbd5e1; }
.field input { width: 100%; padding: .75rem 1rem; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #e2e8f0; font-size: .875rem; outline: none; transition: border-color .2s; }
.field input:focus { border-color: #818cf8; }
.btn-primary { width: 100%; padding: .75rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; cursor: pointer; transition: background .2s; }
.btn-primary:hover { background: #4f46e5; }
.link { display: block; text-align: right; color: #818cf8; font-size: .8125rem; text-decoration: none; margin-top: .75rem; }
.divider { display: flex; align-items: center; gap: 1rem; margin: 1.5rem 0; color: #475569; font-size: .8125rem; }
.divider::before, .divider::after { content: ""; flex: 1; height: 1px; background: #334155; }
.social-buttons { display: flex; flex-direction: column; gap: .5rem; }
.btn-social { width: 100%; padding: .625rem; background: transparent; border: 1px solid #334155; border-radius: 8px; color: #cbd5e1; font-size: .8125rem; cursor: pointer; transition: background .2s; }
.btn-social:hover { background: #334155; }
.footer-text { text-align: center; color: #94a3b8; font-size: .8125rem; margin-top: 1.5rem; }
.footer-text a { color: #818cf8; text-decoration: none; }',
true, true, NOW(), NOW()),

-- Sign Up
('f0000001-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 'signup', 'Default Sign Up',
'<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Sign Up - {{application.name}}</title></head>
<body>
  <div class="card">
    <h1>{{t.signup.title}}</h1>
    <p class="subtitle">{{t.signup.subtitle}}</p>
    <form>
      <div class="field"><label>{{t.signup.name}}</label><input type="text" placeholder="John Doe" /></div>
      <div class="field"><label>{{t.signup.email}}</label><input type="email" placeholder="you@example.com" /></div>
      <div class="field"><label>{{t.signup.password}}</label><input type="password" placeholder="Min. 8 characters" /></div>
      <button type="submit" class="btn-primary">{{t.signup.submit}}</button>
    </form>
    <p class="footer-text">{{t.signup.have_account}} <a href="#">{{t.signup.sign_in}}</a></p>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 420px; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
h1 { text-align: center; font-size: 1.5rem; font-weight: 700; margin-bottom: .25rem; }
.subtitle { text-align: center; color: #94a3b8; font-size: .875rem; margin-bottom: 2rem; }
.field { margin-bottom: 1.25rem; }
.field label { display: block; font-size: .8125rem; font-weight: 500; margin-bottom: .5rem; color: #cbd5e1; }
.field input { width: 100%; padding: .75rem 1rem; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #e2e8f0; font-size: .875rem; outline: none; }
.field input:focus { border-color: #818cf8; }
.btn-primary { width: 100%; padding: .75rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; cursor: pointer; }
.btn-primary:hover { background: #4f46e5; }
.footer-text { text-align: center; color: #94a3b8; font-size: .8125rem; margin-top: 1.5rem; }
.footer-text a { color: #818cf8; text-decoration: none; }',
true, true, NOW(), NOW()),

-- Email Verification
('f0000001-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001', 'verification', 'Default Verification',
'<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Verify Email</title></head>
<body>
  <div class="card">
    <div class="icon">&#9993;</div>
    <h1>{{t.verify.title}}</h1>
    <p class="subtitle">{{t.verify.subtitle}} <strong>{{user.email}}</strong></p>
    <div class="code-box">{{code}}</div>
    <p class="hint">{{t.verify.hint}}</p>
    <a href="{{link}}" class="btn-primary">{{t.verify.button}}</a>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 420px; text-align: center; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
.icon { font-size: 3rem; margin-bottom: 1rem; }
h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: .5rem; }
.subtitle { color: #94a3b8; font-size: .875rem; margin-bottom: 1.5rem; }
.code-box { background: #0f172a; border: 2px dashed #334155; border-radius: 8px; padding: 1rem; font-size: 2rem; font-weight: 700; letter-spacing: .5rem; font-family: monospace; color: #818cf8; margin-bottom: 1rem; }
.hint { color: #64748b; font-size: .75rem; margin-bottom: 1.5rem; }
.btn-primary { display: inline-block; padding: .75rem 2rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; text-decoration: none; }',
true, true, NOW(), NOW()),

-- Password Reset
('f0000001-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001', 'password_reset', 'Default Password Reset',
'<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Reset Password</title></head>
<body>
  <div class="card">
    <h1>{{t.reset.title}}</h1>
    <p class="subtitle">{{t.reset.subtitle}}</p>
    <form>
      <div class="field"><label>{{t.reset.new_password}}</label><input type="password" placeholder="New password" /></div>
      <div class="field"><label>{{t.reset.confirm}}</label><input type="password" placeholder="Confirm password" /></div>
      <button type="submit" class="btn-primary">{{t.reset.submit}}</button>
    </form>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 420px; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
h1 { text-align: center; font-size: 1.5rem; font-weight: 700; margin-bottom: .25rem; }
.subtitle { text-align: center; color: #94a3b8; font-size: .875rem; margin-bottom: 2rem; }
.field { margin-bottom: 1.25rem; }
.field label { display: block; font-size: .8125rem; font-weight: 500; margin-bottom: .5rem; color: #cbd5e1; }
.field input { width: 100%; padding: .75rem 1rem; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #e2e8f0; font-size: .875rem; outline: none; }
.field input:focus { border-color: #818cf8; }
.btn-primary { width: 100%; padding: .75rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; cursor: pointer; }',
true, true, NOW(), NOW()),

-- MFA Challenge
('f0000001-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001', 'mfa_challenge', 'Default MFA Challenge',
'<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>MFA Verification</title></head>
<body>
  <div class="card">
    <div class="icon">&#128274;</div>
    <h1>{{t.mfa.title}}</h1>
    <p class="subtitle">{{t.mfa.subtitle}}</p>
    <form>
      <div class="field"><input type="text" placeholder="000000" class="code-input" maxlength="6" /></div>
      <button type="submit" class="btn-primary">{{t.mfa.submit}}</button>
    </form>
    <a href="#" class="link">{{t.mfa.use_recovery}}</a>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 420px; text-align: center; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
.icon { font-size: 3rem; margin-bottom: 1rem; }
h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: .5rem; }
.subtitle { color: #94a3b8; font-size: .875rem; margin-bottom: 2rem; }
.field { margin-bottom: 1.5rem; }
.code-input { width: 100%; padding: 1rem; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #818cf8; font-size: 2rem; text-align: center; letter-spacing: 1rem; font-family: monospace; outline: none; }
.code-input:focus { border-color: #818cf8; }
.btn-primary { width: 100%; padding: .75rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; cursor: pointer; }
.link { display: block; color: #818cf8; font-size: .8125rem; text-decoration: none; margin-top: 1rem; }',
true, true, NOW(), NOW()),

-- Error Page
('f0000001-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001', 'error', 'Default Error Page',
'<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Error</title></head>
<body>
  <div class="card">
    <div class="icon">&#9888;</div>
    <h1>{{t.error.title}}</h1>
    <p class="error-msg">{{error}}</p>
    <p class="subtitle">{{t.error.subtitle}}</p>
    <a href="#" class="btn-primary">{{t.error.back}}</a>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 420px; text-align: center; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
.icon { font-size: 3rem; margin-bottom: 1rem; color: #f87171; }
h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 1rem; }
.error-msg { background: #450a0a; border: 1px solid #991b1b; border-radius: 8px; padding: .75rem 1rem; color: #fca5a5; font-size: .875rem; margin-bottom: 1rem; font-family: monospace; }
.subtitle { color: #94a3b8; font-size: .875rem; margin-bottom: 1.5rem; }
.btn-primary { display: inline-block; padding: .75rem 2rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; text-decoration: none; }',
true, true, NOW(), NOW()),

-- OAuth Consent
('f0000001-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001', 'consent', 'Default Consent',
'<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Authorize - {{application.name}}</title></head>
<body>
  <div class="card">
    <h1>{{t.consent.title}}</h1>
    <p class="subtitle"><strong>{{application.name}}</strong> {{t.consent.wants_access}}</p>
    <div class="scope-list">
      <div class="scope-item"><span class="check">&#10003;</span> {{t.consent.scope_profile}}</div>
      <div class="scope-item"><span class="check">&#10003;</span> {{t.consent.scope_email}}</div>
    </div>
    <p class="user-info">{{t.consent.signed_in}} <strong>{{user.email}}</strong></p>
    <div class="actions">
      <button class="btn-secondary">{{t.consent.deny}}</button>
      <button class="btn-primary">{{t.consent.allow}}</button>
    </div>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 420px; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
h1 { text-align: center; font-size: 1.5rem; font-weight: 700; margin-bottom: .5rem; }
.subtitle { text-align: center; color: #94a3b8; font-size: .875rem; margin-bottom: 1.5rem; }
.scope-list { background: #0f172a; border-radius: 8px; padding: 1rem; margin-bottom: 1.5rem; }
.scope-item { display: flex; align-items: center; gap: .5rem; padding: .5rem 0; font-size: .875rem; }
.check { color: #34d399; font-weight: bold; }
.user-info { text-align: center; color: #94a3b8; font-size: .8125rem; margin-bottom: 1.5rem; }
.actions { display: flex; gap: .75rem; }
.btn-secondary { flex: 1; padding: .75rem; background: transparent; border: 1px solid #334155; border-radius: 8px; color: #cbd5e1; font-size: .875rem; cursor: pointer; }
.btn-primary { flex: 1; padding: .75rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; cursor: pointer; }',
true, true, NOW(), NOW())

ON CONFLICT (id) DO NOTHING;

-- 4. Seed default language strings
INSERT INTO template_language_strings (tenant_id, string_key, locale, value) VALUES
-- Login strings
('a0000000-0000-0000-0000-000000000001', 'login.title', 'en', 'Welcome back'),
('a0000000-0000-0000-0000-000000000001', 'login.title', 'de', 'Willkommen zurueck'),
('a0000000-0000-0000-0000-000000000001', 'login.subtitle', 'en', 'Sign in to your account'),
('a0000000-0000-0000-0000-000000000001', 'login.subtitle', 'de', 'Melden Sie sich an'),
('a0000000-0000-0000-0000-000000000001', 'login.email', 'en', 'Email address'),
('a0000000-0000-0000-0000-000000000001', 'login.email', 'de', 'E-Mail-Adresse'),
('a0000000-0000-0000-0000-000000000001', 'login.password', 'en', 'Password'),
('a0000000-0000-0000-0000-000000000001', 'login.password', 'de', 'Passwort'),
('a0000000-0000-0000-0000-000000000001', 'login.submit', 'en', 'Sign in'),
('a0000000-0000-0000-0000-000000000001', 'login.submit', 'de', 'Anmelden'),
('a0000000-0000-0000-0000-000000000001', 'login.forgot_password', 'en', 'Forgot password?'),
('a0000000-0000-0000-0000-000000000001', 'login.forgot_password', 'de', 'Passwort vergessen?'),
('a0000000-0000-0000-0000-000000000001', 'login.no_account', 'en', 'Don''t have an account?'),
('a0000000-0000-0000-0000-000000000001', 'login.no_account', 'de', 'Noch kein Konto?'),
('a0000000-0000-0000-0000-000000000001', 'login.sign_up', 'en', 'Sign up'),
('a0000000-0000-0000-0000-000000000001', 'login.sign_up', 'de', 'Registrieren'),
-- Signup strings
('a0000000-0000-0000-0000-000000000001', 'signup.title', 'en', 'Create your account'),
('a0000000-0000-0000-0000-000000000001', 'signup.title', 'de', 'Konto erstellen'),
('a0000000-0000-0000-0000-000000000001', 'signup.subtitle', 'en', 'Get started in seconds'),
('a0000000-0000-0000-0000-000000000001', 'signup.subtitle', 'de', 'Starten Sie in Sekunden'),
('a0000000-0000-0000-0000-000000000001', 'signup.name', 'en', 'Full name'),
('a0000000-0000-0000-0000-000000000001', 'signup.email', 'en', 'Email address'),
('a0000000-0000-0000-0000-000000000001', 'signup.password', 'en', 'Password'),
('a0000000-0000-0000-0000-000000000001', 'signup.submit', 'en', 'Create account'),
('a0000000-0000-0000-0000-000000000001', 'signup.have_account', 'en', 'Already have an account?'),
('a0000000-0000-0000-0000-000000000001', 'signup.sign_in', 'en', 'Sign in'),
-- Verification strings
('a0000000-0000-0000-0000-000000000001', 'verify.title', 'en', 'Check your email'),
('a0000000-0000-0000-0000-000000000001', 'verify.subtitle', 'en', 'We sent a verification code to'),
('a0000000-0000-0000-0000-000000000001', 'verify.hint', 'en', 'The code expires in 10 minutes'),
('a0000000-0000-0000-0000-000000000001', 'verify.button', 'en', 'Verify Email'),
-- Password Reset strings
('a0000000-0000-0000-0000-000000000001', 'reset.title', 'en', 'Reset your password'),
('a0000000-0000-0000-0000-000000000001', 'reset.subtitle', 'en', 'Enter your new password below'),
('a0000000-0000-0000-0000-000000000001', 'reset.new_password', 'en', 'New password'),
('a0000000-0000-0000-0000-000000000001', 'reset.confirm', 'en', 'Confirm password'),
('a0000000-0000-0000-0000-000000000001', 'reset.submit', 'en', 'Reset password'),
-- MFA strings
('a0000000-0000-0000-0000-000000000001', 'mfa.title', 'en', 'Two-factor authentication'),
('a0000000-0000-0000-0000-000000000001', 'mfa.subtitle', 'en', 'Enter the 6-digit code from your authenticator app'),
('a0000000-0000-0000-0000-000000000001', 'mfa.submit', 'en', 'Verify'),
('a0000000-0000-0000-0000-000000000001', 'mfa.use_recovery', 'en', 'Use a recovery code instead'),
-- Error strings
('a0000000-0000-0000-0000-000000000001', 'error.title', 'en', 'Something went wrong'),
('a0000000-0000-0000-0000-000000000001', 'error.subtitle', 'en', 'Please try again or contact support'),
('a0000000-0000-0000-0000-000000000001', 'error.back', 'en', 'Go back'),
-- Consent strings
('a0000000-0000-0000-0000-000000000001', 'consent.title', 'en', 'Authorize Application'),
('a0000000-0000-0000-0000-000000000001', 'consent.wants_access', 'en', 'wants to access your account'),
('a0000000-0000-0000-0000-000000000001', 'consent.scope_profile', 'en', 'View your profile information'),
('a0000000-0000-0000-0000-000000000001', 'consent.scope_email', 'en', 'View your email address'),
('a0000000-0000-0000-0000-000000000001', 'consent.signed_in', 'en', 'Signed in as'),
('a0000000-0000-0000-0000-000000000001', 'consent.deny', 'en', 'Deny'),
('a0000000-0000-0000-0000-000000000001', 'consent.allow', 'en', 'Allow')
ON CONFLICT (tenant_id, string_key, locale) DO NOTHING;
