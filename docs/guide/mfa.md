# Multi-Factor Authentication

CPI Auth supports multiple MFA methods to add a second layer of security to user accounts. MFA can be optional (user-initiated) or enforced at the tenant level.

## Supported Methods

| Method | Code | Description |
|--------|------|-------------|
| TOTP | `totp` | Time-based one-time passwords (Google Authenticator, Authy, etc.) |
| SMS | `sms` | One-time codes sent via SMS |
| Email | `email` | One-time codes sent via email |
| WebAuthn | `webauthn` | Biometric authenticators and hardware security keys (FIDO2) |

## How MFA Works

MFA in CPI Auth follows a two-phase approach:

1. **Enrollment** -- The user registers an MFA method (scan QR code, verify phone, register security key)
2. **Challenge** -- During login, after the primary credential is verified, the user must complete an MFA challenge

### MFA Flow Diagram

```
User enters email + password
    |
    v
Primary authentication succeeds
    |
    v
Has MFA enrollment?
    |           |
   Yes          No
    |           |
    v           v
MFA Challenge   Is MFA required (tenant setting)?
    |               |           |
    v              Yes          No
Verify code         |           |
    |               v           v
    v           Block login   Issue tokens
Issue tokens    (force enroll)
```

## MFA Enrollment

### TOTP Enrollment

```bash
# Step 1: Begin TOTP enrollment (returns QR code data)
curl -X POST http://localhost:5050/mfa/enroll \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "totp"
  }'
```

Response:

```json
{
  "enrollment_id": "e1f2a3b4-c5d6-7890-efgh-i12345678901",
  "method": "totp",
  "totp_uri": "otpauth://totp/CPI Auth:jane@example.com?secret=BASE32SECRET&issuer=CPI Auth",
  "secret": "BASE32SECRET",
  "recovery_codes": [
    "ABCD-1234-EFGH",
    "IJKL-5678-MNOP",
    "QRST-9012-UVWX",
    "YZAB-3456-CDEF",
    "GHIJ-7890-KLMN",
    "OPQR-1234-STUV",
    "WXYZ-5678-ABCD",
    "EFGH-9012-IJKL"
  ]
}
```

Display the `totp_uri` as a QR code for the user to scan with their authenticator app. Store the `recovery_codes` and show them to the user.

```bash
# Step 2: Verify the enrollment with a code from the authenticator app
curl -X POST http://localhost:5050/mfa/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enrollment_id": "e1f2a3b4-c5d6-7890-efgh-i12345678901",
    "code": "123456"
  }'
```

### SMS Enrollment

```bash
curl -X POST http://localhost:5050/mfa/enroll \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "sms",
    "phone": "+1234567890"
  }'
```

A verification code is sent to the phone number. The user enters it to complete enrollment.

### Email Enrollment

```bash
curl -X POST http://localhost:5050/mfa/enroll \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "email"
  }'
```

A verification code is sent to the user's email address.

### WebAuthn Enrollment

WebAuthn enrollment uses the standard FIDO2 registration ceremony. See the [Auth Flows](./auth-flows#webauthn--fido2) section for the full WebAuthn registration and login flow.

## MFA Challenge Flow

During login, if the user has an active MFA enrollment, the primary authentication step returns a challenge instead of tokens:

```bash
# Primary authentication
curl -X POST http://localhost:5050/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "SecureP@ss123"
  }'
```

If MFA is required, the response indicates a challenge is needed:

```json
{
  "mfa_required": true,
  "mfa_token": "temporary-mfa-token",
  "available_methods": ["totp", "email"]
}
```

### Request a Challenge

```bash
curl -X POST http://localhost:5050/mfa/challenge \
  -H "Content-Type: application/json" \
  -d '{
    "mfa_token": "temporary-mfa-token",
    "method": "totp"
  }'
```

For `sms` and `email` methods, this triggers sending a code. For `totp`, no server action is needed -- the user generates the code locally.

### Verify the Challenge

```bash
curl -X POST http://localhost:5050/mfa/verify \
  -H "Content-Type: application/json" \
  -d '{
    "mfa_token": "temporary-mfa-token",
    "code": "123456"
  }'
```

On success, the response contains the access, refresh, and ID tokens as in a normal login.

## Recovery Codes

Recovery codes are generated during MFA enrollment. Each code can be used exactly once as a fallback when the primary MFA method is unavailable.

- 8 recovery codes are generated per enrollment
- Each code is a single-use backup
- Codes are formatted as `XXXX-XXXX-XXXX` for readability
- Once all codes are used, the user must contact an administrator or re-enroll

### Using a Recovery Code

```bash
curl -X POST http://localhost:5050/mfa/verify \
  -H "Content-Type: application/json" \
  -d '{
    "mfa_token": "temporary-mfa-token",
    "recovery_code": "ABCD-1234-EFGH"
  }'
```

## Enforcing MFA per Tenant

Tenant administrators can require MFA for all users:

```bash
curl -X PATCH http://localhost:5050/api/v1/tenants/{tenant_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "settings": {
      "mfa_required": true,
      "allowed_mfa_methods": ["totp", "webauthn"]
    }
  }'
```

When `mfa_required` is `true`:

- Users without an MFA enrollment are prompted to enroll during their next login
- Only methods listed in `allowed_mfa_methods` are available for enrollment
- Users cannot remove their last MFA enrollment

## Admin MFA Management

Administrators can view a user's MFA enrollments:

```bash
curl http://localhost:5050/api/v1/users/{user_id}/mfa \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
[
  {
    "id": "e1f2a3b4-c5d6-7890-efgh-i12345678901",
    "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "method": "totp",
    "verified": true,
    "created_at": "2025-01-20T14:00:00Z"
  }
]
```

## Next Steps

- [Auth Flows](./auth-flows) -- Full authentication flow documentation
- [WebAuthn](./auth-flows#webauthn--fido2) -- Passwordless with biometrics
- [Tenants](./tenants) -- Tenant-level security settings
