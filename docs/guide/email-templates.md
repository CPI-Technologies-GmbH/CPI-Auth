# Email Templates

CPI Auth sends transactional emails for verification, password resets, MFA codes, and more. Every email can be customized per tenant and per locale using HTML or MJML templates.

## Email Template Types

| Type | Trigger | Description |
|------|---------|-------------|
| `verification` | User registration | Email address verification link/code |
| `password_reset` | Password reset request | Password reset link |
| `mfa` | MFA challenge | One-time MFA code via email |
| `welcome` | After email verification | Welcome message to new users |
| `invitation` | Admin invites a user | Account invitation with setup link |
| `magic_link` | Passwordless login | Magic link for passwordless authentication |

## Email Template Model

```json
{
  "id": "e1t2m3l4-a5b6-7890-cdef-012345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "verification",
  "locale": "en",
  "subject": "Verify your email address",
  "body_mjml": "<mjml>...</mjml>",
  "body_html": "<html>...</html>",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

| Field | Description |
|-------|-------------|
| `type` | One of the template types above |
| `locale` | Language code (e.g., `en`, `de`, `fr`, `es`) |
| `subject` | Email subject line (supports template variables) |
| `body_mjml` | MJML source (optional; compiled to HTML on save) |
| `body_html` | Final HTML body (used for sending) |

## MJML Support

[MJML](https://mjml.io/) is a markup language for creating responsive email templates. CPI Auth accepts MJML in the `body_mjml` field and compiles it to HTML automatically.

### Example MJML Template

```xml v-pre
<mjml>
  <mj-head>
    <mj-attributes>
      <mj-all font-family="system-ui, sans-serif" />
      <mj-text font-size="16px" color="#333333" />
    </mj-attributes>
  </mj-head>
  <mj-body background-color="#f4f4f4">
    <mj-section background-color="#ffffff" padding="40px">
      <mj-column>
        <mj-image src="&#123;&#123;tenant.logo_url&#125;&#125;" width="150px" />
        <mj-text font-size="24px" font-weight="bold" padding-top="20px">
          Verify Your Email
        </mj-text>
        <mj-text>
          Hi &#123;&#123;user.name&#125;&#125;,
        </mj-text>
        <mj-text>
          Please verify your email address by clicking the button below
          or entering this code: <strong>&#123;&#123;code&#125;&#125;</strong>
        </mj-text>
        <mj-button background-color="&#123;&#123;tenant.primary_color&#125;&#125;" href="&#123;&#123;link&#125;&#125;">
          Verify Email
        </mj-button>
        <mj-text font-size="12px" color="#999999" padding-top="20px">
          If you didn't create an account with &#123;&#123;tenant.name&#125;&#125;, you can ignore this email.
        </mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
```

If you prefer to write raw HTML, leave `body_mjml` empty and provide only `body_html`.

## Template Variables

Email templates support the same variables as page templates:

| Variable | Description |
|----------|-------------|
| `&#123;&#123;user.name&#125;&#125;` | Recipient's display name |
| `&#123;&#123;user.email&#125;&#125;` | Recipient's email address |
| `&#123;&#123;tenant.name&#125;&#125;` | Tenant display name |
| `&#123;&#123;tenant.logo_url&#125;&#125;` | Tenant logo URL |
| `&#123;&#123;tenant.primary_color&#125;&#125;` | Tenant accent color |
| `&#123;&#123;application.name&#125;&#125;` | Name of the application (if applicable) |
| `&#123;&#123;code&#125;&#125;` | Verification or MFA code |
| `&#123;&#123;link&#125;&#125;` | Action link (verification, password reset, magic link) |

## Per-Locale Templates

Create templates for different languages. CPI Auth selects the template based on the user's `locale` field.

```bash v-pre
# English verification email
curl -X POST http://localhost:5050/api/v1/email-templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "verification",
    "locale": "en",
    "subject": "Verify your email address",
    "body_html": "<html><body><h1>Hi &#123;&#123;user.name&#125;&#125;,</h1><p>Your code is: <strong>&#123;&#123;code&#125;&#125;</strong></p><p><a href=\"&#123;&#123;link&#125;&#125;\">Verify Email</a></p></body></html>"
  }'

# German verification email
curl -X POST http://localhost:5050/api/v1/email-templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "verification",
    "locale": "de",
    "subject": "Bestaetigen Sie Ihre E-Mail-Adresse",
    "body_html": "<html><body><h1>Hallo &#123;&#123;user.name&#125;&#125;,</h1><p>Ihr Code lautet: <strong>&#123;&#123;code&#125;&#125;</strong></p><p><a href=\"&#123;&#123;link&#125;&#125;\">E-Mail bestaetigen</a></p></body></html>"
  }'
```

### Locale Fallback

If no template exists for the user's locale, CPI Auth falls back to `en`. If no English template exists, the system default is used.

## Managing Email Templates

### List Templates

```bash
curl http://localhost:5050/api/v1/email-templates?page=1&per_page=20 \
  -H "Authorization: Bearer $TOKEN"
```

### Get a Template

```bash
curl http://localhost:5050/api/v1/email-templates/{template_id} \
  -H "Authorization: Bearer $TOKEN"
```

### Update a Template

```bash v-pre
curl -X PATCH http://localhost:5050/api/v1/email-templates/{template_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Please verify your email for &#123;&#123;tenant.name&#125;&#125;",
    "body_mjml": "<mjml>...(updated)...</mjml>"
  }'
```

### Delete a Template

```bash
curl -X DELETE http://localhost:5050/api/v1/email-templates/{template_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Testing Emails

Send a test email to verify your template renders correctly:

```bash
curl -X POST http://localhost:5050/api/v1/email-templates/{template_id}/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "test@example.com"
  }'
```

The test email is rendered with sample data (similar to page template preview) and sent to the specified address. In development, all emails are captured by MailHog at [http://localhost:5059](http://localhost:5059).

## SMTP Testing

Test your SMTP configuration without sending a real email:

```bash
curl -X POST http://localhost:5050/api/v1/settings/smtp/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "admin@example.com"
  }'
```

## Email Delivery

CPI Auth sends emails via SMTP. For production, configure a transactional email service:

| Provider | SMTP Host | Port |
|----------|-----------|------|
| SendGrid | `smtp.sendgrid.net` | 587 |
| Amazon SES | `email-smtp.{region}.amazonaws.com` | 587 |
| Postmark | `smtp.postmarkapp.com` | 587 |
| Mailgun | `smtp.mailgun.org` | 587 |

Configure the SMTP settings in `config.yaml` or via environment variables:

```yaml
smtp:
  host: smtp.sendgrid.net
  port: 587
  from: "noreply@yourdomain.com"
  username: "apikey"
  password: "SG.your-api-key"
```

## Next Steps

- [Page Templates](./page-templates) -- Customize auth pages
- [Configuration](./configuration) -- SMTP configuration reference
- [Custom Fields](./custom-fields) -- Include custom data in emails
