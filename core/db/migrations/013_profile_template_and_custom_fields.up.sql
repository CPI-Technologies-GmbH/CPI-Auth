-- Add profile page type and update signup template with custom fields support

-- Update signup template to include custom fields placeholder
UPDATE page_templates
SET html_content = '<!DOCTYPE html>
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
      {{custom_fields}}
      <button type="submit" class="btn-primary">{{t.signup.submit}}</button>
    </form>
    <p class="footer-text">{{t.signup.have_account}} <a href="#">{{t.signup.sign_in}}</a></p>
  </div>
</body>
</html>'
WHERE id = 'f0000001-0000-0000-0000-000000000002';

-- Add profile template
INSERT INTO page_templates (id, tenant_id, page_type, name, html_content, css_content, is_active, is_default, created_at, updated_at)
VALUES (
'f0000001-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001', 'profile', 'Default Profile',
'<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Profile - {{application.name}}</title></head>
<body>
  <div class="card">
    <div class="avatar">{{user.initials}}</div>
    <h1>{{t.profile.title}}</h1>
    <p class="subtitle">{{t.profile.subtitle}}</p>
    <form>
      <div class="field">
        <label>{{t.profile.name}}</label>
        <input type="text" value="{{user.name}}" />
      </div>
      <div class="field">
        <label>{{t.profile.email}}</label>
        <input type="email" value="{{user.email}}" disabled />
        <span class="hint">{{t.profile.email_hint}}</span>
      </div>
      {{profile_fields}}
      <div class="section">
        <h2>{{t.profile.security}}</h2>
        <a href="#" class="btn-secondary">{{t.profile.change_password}}</a>
        <a href="#" class="btn-secondary">{{t.profile.manage_mfa}}</a>
      </div>
      <button type="submit" class="btn-primary">{{t.profile.save}}</button>
    </form>
  </div>
</body>
</html>',
'* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.card { background: #1e293b; border-radius: 16px; padding: 2.5rem; width: 100%; max-width: 480px; box-shadow: 0 25px 50px -12px rgba(0,0,0,.5); }
.avatar { width: 64px; height: 64px; border-radius: 50%; background: #6366f1; display: flex; align-items: center; justify-content: center; font-size: 1.5rem; font-weight: 700; margin: 0 auto 1rem; }
h1 { text-align: center; font-size: 1.5rem; font-weight: 700; margin-bottom: .25rem; }
h2 { font-size: 1rem; font-weight: 600; margin-bottom: .75rem; color: #cbd5e1; }
.subtitle { text-align: center; color: #94a3b8; font-size: .875rem; margin-bottom: 2rem; }
.field { margin-bottom: 1.25rem; }
.field label { display: block; font-size: .8125rem; font-weight: 500; margin-bottom: .5rem; color: #cbd5e1; }
.field input, .field select, .field textarea { width: 100%; padding: .75rem 1rem; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #e2e8f0; font-size: .875rem; outline: none; }
.field input:focus, .field select:focus, .field textarea:focus { border-color: #818cf8; }
.field input:disabled { opacity: .5; cursor: not-allowed; }
.field textarea { min-height: 80px; resize: vertical; }
.hint { display: block; font-size: .75rem; color: #64748b; margin-top: .25rem; }
.section { border-top: 1px solid #334155; padding-top: 1.5rem; margin-top: 1.5rem; margin-bottom: 1.5rem; }
.btn-primary { width: 100%; padding: .75rem; background: #6366f1; color: white; border: none; border-radius: 8px; font-size: .875rem; font-weight: 600; cursor: pointer; margin-top: .5rem; }
.btn-primary:hover { background: #4f46e5; }
.btn-secondary { display: block; width: 100%; padding: .625rem 1rem; background: transparent; border: 1px solid #334155; border-radius: 8px; color: #cbd5e1; font-size: .8125rem; cursor: pointer; margin-bottom: .5rem; text-align: center; text-decoration: none; }
.btn-secondary:hover { background: #334155; }
.custom-field { margin-bottom: 1.25rem; }
.custom-field label { display: block; font-size: .8125rem; font-weight: 500; margin-bottom: .5rem; color: #cbd5e1; }
.custom-field .required { color: #f87171; }',
true, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Add profile language strings
INSERT INTO template_language_strings (tenant_id, string_key, locale, value) VALUES
('a0000000-0000-0000-0000-000000000001', 'profile.title', 'en', 'Your Profile'),
('a0000000-0000-0000-0000-000000000001', 'profile.title', 'de', 'Ihr Profil'),
('a0000000-0000-0000-0000-000000000001', 'profile.subtitle', 'en', 'Manage your account settings'),
('a0000000-0000-0000-0000-000000000001', 'profile.subtitle', 'de', 'Kontoeinstellungen verwalten'),
('a0000000-0000-0000-0000-000000000001', 'profile.name', 'en', 'Full name'),
('a0000000-0000-0000-0000-000000000001', 'profile.email', 'en', 'Email address'),
('a0000000-0000-0000-0000-000000000001', 'profile.email_hint', 'en', 'Email cannot be changed here'),
('a0000000-0000-0000-0000-000000000001', 'profile.security', 'en', 'Security'),
('a0000000-0000-0000-0000-000000000001', 'profile.change_password', 'en', 'Change password'),
('a0000000-0000-0000-0000-000000000001', 'profile.manage_mfa', 'en', 'Manage two-factor authentication'),
('a0000000-0000-0000-0000-000000000001', 'profile.save', 'en', 'Save changes'),
-- Add signup custom fields string
('a0000000-0000-0000-0000-000000000001', 'signup.custom_fields_heading', 'en', 'Additional Information'),
('a0000000-0000-0000-0000-000000000001', 'signup.custom_fields_heading', 'de', 'Zusaetzliche Informationen')
ON CONFLICT (tenant_id, string_key, locale) DO NOTHING;
