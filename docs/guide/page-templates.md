# Page Templates

CPI Auth provides a template system for customizing the HTML pages used in authentication flows. Every page a user sees -- login, signup, password reset, MFA challenge -- can be fully customized per tenant.

## Template System Overview

Page templates are HTML documents with CSS that support template variables for dynamic content. Each tenant gets a set of default templates that can be duplicated and customized.

Templates are stored in the database and rendered by the Login UI and Account UI services.

## Default Templates

CPI Auth creates the following default templates for each tenant:

| Page Type | Description |
|-----------|-------------|
| `login` | Email/password login form with social login buttons |
| `signup` | User registration form with custom fields |
| `verification` | Email verification page |
| `password_reset` | Password reset form |
| `mfa_challenge` | MFA code entry page |
| `error` | Error display page |
| `consent` | OAuth consent/authorization page |
| `profile` | User profile editing page |

### Default Templates Are Read-Only

Default templates (where `is_default: true`) cannot be edited directly. To customize a template:

1. Duplicate the default template
2. Edit the duplicate
3. Set the duplicate as active

This ensures you always have the original to fall back on.

## Template Model

```json
{
  "id": "t1a2b3c4-d5e6-7890-tuvw-x12345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "page_type": "login",
  "name": "Custom Login Page",
  "html_content": "<!DOCTYPE html>...",
  "css_content": "body { ... }",
  "is_active": true,
  "is_default": false,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-02-01T08:00:00Z"
}
```

## Template Variables

Use double curly braces to insert dynamic values into your templates.

### User Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;user.name&#125;&#125;` | User's display name |
| `&#123;&#123;user.email&#125;&#125;` | User's email address |
| `&#123;&#123;user.initials&#125;&#125;` | User's initials (derived from name) |
| `&#123;&#123;user.avatar_url&#125;&#125;` | User's avatar URL |
| `&#123;&#123;user.locale&#125;&#125;` | User's preferred locale |

### Application Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;application.name&#125;&#125;` | Name of the OAuth application |
| `&#123;&#123;application.logo_url&#125;&#125;` | Application's logo URL |

### Tenant Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;tenant.name&#125;&#125;` | Tenant display name |
| `&#123;&#123;tenant.logo_url&#125;&#125;` | Tenant logo URL (from branding) |
| `&#123;&#123;tenant.primary_color&#125;&#125;` | Tenant primary color |

### Flow Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;code&#125;&#125;` | Verification or MFA code |
| `&#123;&#123;link&#125;&#125;` | Action link (verification, password reset) |
| `&#123;&#123;error&#125;&#125;` | Error message to display |

### Custom Field Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;custom_fields&#125;&#125;` | Renders registration-visible custom fields |
| `&#123;&#123;profile_fields&#125;&#125;` | Renders profile-visible custom fields |

## Language Strings

Templates support localized strings using the `&#123;&#123;t.key&#125;&#125;` syntax. When the template is rendered, CPI Auth looks up the key in the `language_strings` table for the user's locale.

```html v-pre
<h1>&#123;&#123;t.login_title&#125;&#125;</h1>
<p>&#123;&#123;t.login_subtitle&#125;&#125;</p>
<label>&#123;&#123;t.email_label&#125;&#125;</label>
<input type="email" placeholder="&#123;&#123;t.email_placeholder&#125;&#125;">
<label>&#123;&#123;t.password_label&#125;&#125;</label>
<input type="password" placeholder="&#123;&#123;t.password_placeholder&#125;&#125;">
<button type="submit">&#123;&#123;t.login_button&#125;&#125;</button>
<a href="/signup">&#123;&#123;t.signup_link&#125;&#125;</a>
```

### Managing Language Strings

```bash
# Set a language string
curl -X PUT http://localhost:5050/api/v1/language-strings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "string_key": "login_title",
    "locale": "en",
    "value": "Welcome Back"
  }'

# Set the German translation
curl -X PUT http://localhost:5050/api/v1/language-strings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "string_key": "login_title",
    "locale": "de",
    "value": "Willkommen zurueck"
  }'

# List all strings for a locale
curl "http://localhost:5050/api/v1/language-strings?locale=en" \
  -H "Authorization: Bearer $TOKEN"

# Delete a string
curl -X DELETE http://localhost:5050/api/v1/language-strings/login_title/en \
  -H "Authorization: Bearer $TOKEN"
```

### Locale Resolution

When rendering a template, the locale is resolved in this order:

1. User's `locale` field (if authenticated)
2. `Accept-Language` header from the browser
3. Tenant's default locale
4. Falls back to `en`

If a string is not found for the requested locale, the English (`en`) value is used as a fallback.

## Managing Templates

### List Templates

```bash
curl http://localhost:5050/api/v1/page-templates \
  -H "Authorization: Bearer $TOKEN"
```

### Get a Template

```bash
curl http://localhost:5050/api/v1/page-templates/{template_id} \
  -H "Authorization: Bearer $TOKEN"
```

### Duplicate a Template

Since default templates are read-only, duplicate them to create an editable copy:

```bash
curl -X POST http://localhost:5050/api/v1/page-templates/{template_id}/duplicate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Custom Login"
  }'
```

### Create a Template

```bash v-pre
curl -X POST http://localhost:5050/api/v1/page-templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "page_type": "login",
    "name": "Minimal Login",
    "html_content": "<!DOCTYPE html><html><head><title>&#123;&#123;t.login_title&#125;&#125;</title><style>&#123;&#123;css&#125;&#125;</style></head><body><div class=\"container\"><img src=\"&#123;&#123;tenant.logo_url&#125;&#125;\" alt=\"Logo\"><h1>&#123;&#123;t.login_title&#125;&#125;</h1><form method=\"POST\"><input type=\"email\" name=\"email\" placeholder=\"&#123;&#123;t.email_placeholder&#125;&#125;\" required><input type=\"password\" name=\"password\" placeholder=\"&#123;&#123;t.password_placeholder&#125;&#125;\" required><button type=\"submit\">&#123;&#123;t.login_button&#125;&#125;</button></form></div></body></html>",
    "css_content": ".container { max-width: 400px; margin: 100px auto; text-align: center; font-family: system-ui; }",
    "is_active": false
  }'
```

### Update a Template

```bash
curl -X PATCH http://localhost:5050/api/v1/page-templates/{template_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "html_content": "<!DOCTYPE html>...(updated)...",
    "is_active": true
  }'
```

### Delete a Template

```bash
curl -X DELETE http://localhost:5050/api/v1/page-templates/{template_id} \
  -H "Authorization: Bearer $TOKEN"
```

::: warning
Default templates cannot be deleted. Only custom (duplicated or created) templates can be removed.
:::

## Custom Page Type

In addition to the standard page types, you can create templates with a `custom` page type for arbitrary pages:

```bash
curl -X POST http://localhost:5050/api/v1/page-templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "page_type": "custom",
    "name": "Terms of Service",
    "html_content": "<!DOCTYPE html><html>...</html>",
    "css_content": "..."
  }'
```

## Preview with Sample Data

The Admin UI provides a live preview feature that renders templates with sample data. This lets you see how variables are resolved without going through a real authentication flow.

Sample data used for preview:

```json
{
  "user": {
    "name": "Jane Smith",
    "email": "jane@example.com",
    "initials": "JS"
  },
  "application": {
    "name": "My App"
  },
  "tenant": {
    "name": "Acme Corp",
    "logo_url": "https://example.com/logo.png",
    "primary_color": "#4F46E5"
  },
  "code": "123456",
  "link": "https://example.com/verify?token=sample",
  "error": ""
}
```

## Next Steps

- [Custom Fields](./custom-fields) -- Define fields that appear in templates
- [Email Templates](./email-templates) -- Customize email content
- [Tenants](./tenants) -- Tenant branding that feeds into templates
