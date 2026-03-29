# Language Strings API

Language strings provide localized text for page templates. CPI Auth supports multiple locales and uses a key-value system where keys follow a `page.element` naming convention.

## Base URL

```
http://localhost:5054/admin/language-strings
```

## Supported Locales

| Code | Language |
|------|----------|
| `en` | English |
| `de` | German (Deutsch) |
| `fr` | French (Francais) |
| `es` | Spanish (Espanol) |

---

## List Language Strings

### GET /admin/language-strings

Retrieve all language strings, optionally filtered by locale.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `locale` | string | | Filter by locale (`en`, `de`, `fr`, `es`) |
| `search` | string | | Search by key or value |
| `page` | integer | 1 | Page number |
| `per_page` | integer | 50 | Items per page |

```bash
curl "http://localhost:5054/admin/language-strings?locale=en" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "string_key": "login.title",
      "locale": "en",
      "value": "Welcome back",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "string_key": "login.email_placeholder",
      "locale": "en",
      "value": "Email address",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "string_key": "login.password_placeholder",
      "locale": "en",
      "value": "Password",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "string_key": "login.submit_button",
      "locale": "en",
      "value": "Sign in",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "string_key": "login.signup_link",
      "locale": "en",
      "value": "Don't have an account? Sign up",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "total": 5
}
```

### Fetching All Locales for a Key

```bash
curl "http://localhost:5054/admin/language-strings?search=login.title" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

```json
{
  "data": [
    { "string_key": "login.title", "locale": "en", "value": "Welcome back" },
    { "string_key": "login.title", "locale": "de", "value": "Willkommen zurueck" },
    { "string_key": "login.title", "locale": "fr", "value": "Bon retour" },
    { "string_key": "login.title", "locale": "es", "value": "Bienvenido de nuevo" }
  ],
  "total": 4
}
```

---

## Upsert Language String

### PUT /admin/language-strings

Create a new language string or update an existing one. The combination of `string_key` and `locale` is the unique identifier.

```bash
curl -X PUT http://localhost:5054/admin/language-strings \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "string_key": "login.title",
    "locale": "en",
    "value": "Sign in to your account"
  }'
```

**Response 200 OK:**

```json
{
  "string_key": "login.title",
  "locale": "en",
  "value": "Sign in to your account",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

### Batch Upsert Example

To set translations for multiple locales, make multiple requests:

```bash
# English
curl -X PUT http://localhost:5054/admin/language-strings \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{"string_key": "signup.title", "locale": "en", "value": "Create your account"}'

# German
curl -X PUT http://localhost:5054/admin/language-strings \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{"string_key": "signup.title", "locale": "de", "value": "Konto erstellen"}'

# French
curl -X PUT http://localhost:5054/admin/language-strings \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{"string_key": "signup.title", "locale": "fr", "value": "Creez votre compte"}'

# Spanish
curl -X PUT http://localhost:5054/admin/language-strings \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{"string_key": "signup.title", "locale": "es", "value": "Crea tu cuenta"}'
```

---

## Delete Language String

### DELETE /admin/language-strings/:key/:locale

Remove a specific language string for a given locale.

```bash
curl -X DELETE http://localhost:5054/admin/language-strings/signup.title/fr \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

---

## Key Naming Conventions

Language string keys follow a `page.element` pattern for organization:

| Pattern | Examples |
|---------|----------|
| `login.*` | `login.title`, `login.submit_button`, `login.email_placeholder` |
| `signup.*` | `signup.title`, `signup.terms_checkbox`, `signup.submit_button` |
| `verification.*` | `verification.title`, `verification.message`, `verification.resend` |
| `password_reset.*` | `password_reset.title`, `password_reset.instructions` |
| `mfa.*` | `mfa.title`, `mfa.code_placeholder`, `mfa.verify_button` |
| `error.*` | `error.title`, `error.generic_message`, `error.back_button` |
| `consent.*` | `consent.title`, `consent.approve_button`, `consent.deny_button` |
| `profile.*` | `profile.title`, `profile.save_button`, `profile.change_password` |
| `common.*` | `common.or`, `common.loading`, `common.cancel`, `common.submit` |

### Default Keys Reference

| Key | English Default |
|-----|----------------|
| `login.title` | Welcome back |
| `login.email_placeholder` | Email address |
| `login.password_placeholder` | Password |
| `login.submit_button` | Sign in |
| `login.forgot_password` | Forgot password? |
| `login.signup_link` | Don't have an account? Sign up |
| `signup.title` | Create your account |
| `signup.submit_button` | Sign up |
| `signup.login_link` | Already have an account? Sign in |
| `signup.terms_checkbox` | I agree to the Terms of Service |
| `verification.title` | Verify your email |
| `verification.message` | We sent a verification code to your email |
| `verification.resend` | Resend code |
| `password_reset.title` | Reset your password |
| `password_reset.instructions` | Enter your email to receive a reset link |
| `password_reset.submit_button` | Send reset link |
| `mfa.title` | Two-factor authentication |
| `mfa.code_placeholder` | Enter your code |
| `mfa.verify_button` | Verify |
| `error.title` | Something went wrong |
| `error.back_button` | Go back |

---

## Using in Templates

Language strings are accessed in page templates using the `&#123;&#123;t.key&#125;&#125;` syntax, where dots in the key are replaced with underscores:

```html v-pre
<!-- In a page template -->
<h1>&#123;&#123;t.login_title&#125;&#125;</h1>
<form>
  <input type="email" placeholder="&#123;&#123;t.login_email_placeholder&#125;&#125;">
  <input type="password" placeholder="&#123;&#123;t.login_password_placeholder&#125;&#125;">
  <button type="submit">&#123;&#123;t.login_submit_button&#125;&#125;</button>
</form>
<a href="/forgot-password">&#123;&#123;t.login_forgot_password&#125;&#125;</a>
```

The correct locale is resolved automatically based on the user's locale preference, browser `Accept-Language` header, or the tenant's default locale.
