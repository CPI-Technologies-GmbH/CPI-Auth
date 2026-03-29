# Page Templates API

Page templates control the HTML/CSS rendering of CPI Auth's user-facing pages (login, signup, password reset, etc.). Templates support dynamic variables and language strings for localization.

## Base URL

```
http://localhost:5054/admin/page-templates
```

## Key Rules

- **Default templates** are read-only. You cannot update or delete them (returns `403`).
- To customize a default template, **duplicate** it first.
- Templates use Handlebars-style variables: `&#123;&#123;variable&#125;&#125;`.
- Language strings are accessed with `&#123;&#123;t.key&#125;&#125;`.

---

## List Templates

### GET /admin/page-templates

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `per_page` | integer | 20 | Items per page |
| `type` | string | | Filter by page type |
| `search` | string | | Search by name |

```bash
curl "http://localhost:5054/admin/page-templates?type=login" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "tpl-uuid-1",
      "name": "Default Login",
      "type": "login",
      "is_default": true,
      "is_active": true,
      "html": "<!DOCTYPE html>...",
      "css": "body { ... }",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "tpl-uuid-2",
      "name": "Custom Login Dark",
      "type": "login",
      "is_default": false,
      "is_active": true,
      "html": "<!DOCTYPE html>...",
      "css": "body { background: #1a1a2e; }",
      "created_at": "2026-02-10T08:00:00Z",
      "updated_at": "2026-03-28T12:00:00Z"
    }
  ],
  "total": 2
}
```

---

## Create Template

### POST /admin/page-templates

```bash v-pre
curl -X POST http://localhost:5054/admin/page-templates \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "Branded Login",
    "type": "login",
    "html": "<!DOCTYPE html><html><head><style>&#123;&#123;css&#125;&#125;</style></head><body><div class=\"container\"><img src=\"&#123;&#123;branding.logo_url&#125;&#125;\" alt=\"Logo\"><h1>&#123;&#123;t.login_title&#125;&#125;</h1><form method=\"POST\"><input type=\"email\" name=\"email\" placeholder=\"&#123;&#123;t.email_placeholder&#125;&#125;\"><input type=\"password\" name=\"password\" placeholder=\"&#123;&#123;t.password_placeholder&#125;&#125;\"><button type=\"submit\">&#123;&#123;t.login_button&#125;&#125;</button></form><p>&#123;&#123;t.signup_link&#125;&#125;</p></div></body></html>",
    "css": ".container { max-width: 400px; margin: 100px auto; padding: 2rem; font-family: var(--af-font-family); } button { background: var(--af-color-primary); color: white; border: none; padding: 12px 24px; border-radius: var(--af-border-radius); cursor: pointer; width: 100%; }"
  }'
```

**Response 201 Created:**

```json
{
  "id": "tpl-uuid-3",
  "name": "Branded Login",
  "type": "login",
  "is_default": false,
  "is_active": true,
  "html": "<!DOCTYPE html>...",
  "css": ".container { ... }",
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

---

## Get Template

### GET /admin/page-templates/:id

```bash
curl http://localhost:5054/admin/page-templates/tpl-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

---

## Update Template

### PATCH /admin/page-templates/:id

Only custom templates can be updated. Default templates return `403`.

```bash
curl -X PATCH http://localhost:5054/admin/page-templates/tpl-uuid-3 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "Branded Login v2",
    "css": ".container { max-width: 450px; margin: 80px auto; }"
  }'
```

### Error: Updating Default Template

```json
{
  "error": "forbidden",
  "error_description": "Default templates are read-only. Duplicate the template to create an editable copy."
}
```

---

## Delete Template

### DELETE /admin/page-templates/:id

Only custom templates can be deleted. Default templates return `403`.

```bash
curl -X DELETE http://localhost:5054/admin/page-templates/tpl-uuid-3 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

---

## Duplicate Template

### POST /admin/page-templates/:id/duplicate

Create an editable copy of any template (including defaults).

```bash
curl -X POST http://localhost:5054/admin/page-templates/tpl-uuid-1/duplicate \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "My Custom Login"
  }'
```

**Response 201 Created:**

```json
{
  "id": "tpl-uuid-4",
  "name": "My Custom Login",
  "type": "login",
  "is_default": false,
  "is_active": false,
  "html": "<!DOCTYPE html>...",
  "css": "body { ... }",
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

---

## Page Types

| Type | Description |
|------|-------------|
| `login` | User login page |
| `signup` | New user registration |
| `verification` | Email verification page |
| `password_reset` | Password reset request and form |
| `mfa_challenge` | Multi-factor authentication prompt |
| `error` | Error display page |
| `consent` | OAuth consent screen |
| `profile` | User profile management |
| `custom` | Custom page for any purpose |

---

## Template Variables Reference

Variables are injected at render time and accessed with `&#123;&#123;variable&#125;&#125;` syntax.

### User Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;user.name&#125;&#125;` | User's full name |
| `&#123;&#123;user.email&#125;&#125;` | User's email address |
| `&#123;&#123;user.locale&#125;&#125;` | User's preferred locale |
| `&#123;&#123;user.email_verified&#125;&#125;` | Whether email is verified |

### Branding Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;branding.logo_url&#125;&#125;` | Logo URL (light mode) |
| `&#123;&#123;branding.logo_dark_url&#125;&#125;` | Logo URL (dark mode) |
| `&#123;&#123;branding.primary_color&#125;&#125;` | Primary brand color |
| `&#123;&#123;branding.secondary_color&#125;&#125;` | Secondary brand color |
| `&#123;&#123;branding.font_family&#125;&#125;` | Configured font family |

### Application Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;application.name&#125;&#125;` | Application display name |
| `&#123;&#123;application.logo_url&#125;&#125;` | Application logo |

### Page Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;csrf_token&#125;&#125;` | CSRF protection token |
| `&#123;&#123;error&#125;&#125;` | Error message (if any) |
| `&#123;&#123;redirect_uri&#125;&#125;` | Post-action redirect URL |
| `&#123;&#123;css&#125;&#125;` | Compiled CSS (use in `<style>` tag) |
| `&#123;&#123;custom_fields&#125;&#125;` | Rendered custom field HTML |

### Language String Variables

| Variable | Description |
|----------|-------------|
| `&#123;&#123;t.login_title&#125;&#125;` | Localized login page title |
| `&#123;&#123;t.email_placeholder&#125;&#125;` | "Email" in current locale |
| `&#123;&#123;t.password_placeholder&#125;&#125;` | "Password" in current locale |
| `&#123;&#123;t.login_button&#125;&#125;` | Login button text |
| `&#123;&#123;t.signup_link&#125;&#125;` | "Don't have an account?" text |
| `&#123;&#123;t.forgot_password&#125;&#125;` | "Forgot password?" link text |
| `&#123;&#123;t.any_custom_key&#125;&#125;` | Any custom language string key |

See the [Language Strings API](/api/language-strings) for managing translation values.

---

## Design Token CSS Properties

Templates can use CSS custom properties set by the branding configuration:

```css
.button {
  background-color: var(--af-color-primary);
  color: var(--af-color-text);
  font-family: var(--af-font-family);
  border-radius: var(--af-border-radius);
}

.page {
  background-color: var(--af-color-background);
}

.accent {
  color: var(--af-color-secondary);
}
```

See the [Design Tokens guide](/cli/design-tokens) for the full list of available CSS properties.
