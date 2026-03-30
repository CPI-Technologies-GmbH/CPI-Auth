# Authentication Flows

CPI Auth supports multiple authentication flows for different application types and use cases. This page provides step-by-step documentation for each flow.

## Authorization Code Flow with PKCE

This is the recommended flow for all user-facing applications (SPAs, native apps, and web apps). PKCE (Proof Key for Code Exchange) prevents authorization code interception attacks.

### Step 1: Generate PKCE Parameters

```javascript
// Generate a random code verifier (43-128 characters)
function generateCodeVerifier() {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

// Derive the code challenge using SHA-256
async function generateCodeChallenge(verifier) {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

const codeVerifier = generateCodeVerifier();
const codeChallenge = await generateCodeChallenge(codeVerifier);
```

### Step 2: Redirect to Authorization Endpoint

```
GET http://localhost:5050/oauth/authorize
  ?client_id=YOUR_CLIENT_ID
  &redirect_uri=http://localhost:3000/callback
  &response_type=code
  &scope=openid profile email
  &state=RANDOM_STATE
  &code_challenge=CODE_CHALLENGE
  &code_challenge_method=S256
  &nonce=RANDOM_NONCE
```

The user is redirected to the Login UI where they authenticate. After successful authentication, CPI Auth redirects back to your `redirect_uri` with an authorization code.

### Step 3: Exchange Code for Tokens

```bash
curl -X POST http://localhost:5050/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "authorization_code",
    "code": "AUTH_CODE_FROM_REDIRECT",
    "redirect_uri": "http://localhost:3000/callback",
    "client_id": "YOUR_CLIENT_ID",
    "code_verifier": "ORIGINAL_CODE_VERIFIER"
  }'
```

Response:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "af_rt_xK9mQ2rT5wZ8...",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "scope": "openid profile email"
}
```

## Client Credentials Flow (M2M)

For machine-to-machine communication where no user is involved.

```bash
curl -X POST http://localhost:5050/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "YOUR_M2M_CLIENT_ID",
    "client_secret": "YOUR_CLIENT_SECRET",
    "scope": "read:data write:data"
  }'
```

Response:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read:data write:data"
}
```

No refresh token is returned for client credentials grants. Request a new token when the current one expires.

## Refresh Token Rotation

When exchanging a refresh token, CPI Auth issues a new refresh token and invalidates the old one. If a previously used refresh token is presented (indicating potential theft), the entire token family is revoked.

```bash
curl -X POST http://localhost:5050/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "refresh_token",
    "refresh_token": "af_rt_xK9mQ2rT5wZ8...",
    "client_id": "YOUR_CLIENT_ID"
  }'
```

Response:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...(new)",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "af_rt_newToken123...(rotated)",
  "id_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

### Token Family Revocation

Refresh tokens belong to a "family" that traces back to the original authorization. If a refresh token that has already been used is presented again, CPI Auth revokes every token in that family. This protects against token replay attacks.

## Token Revocation

Explicitly revoke an access or refresh token:

```bash
curl -X POST http://localhost:5050/oauth/revoke \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJSUzI1NiIs...",
    "token_type_hint": "access_token"
  }'
```

## Token Structure

### Access Token (JWT)

CPI Auth issues JWTs signed with RS256 (or ES256 if configured). The access token contains:

```json
{
  "sub": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "iss": "http://localhost:5050",
  "aud": "YOUR_CLIENT_ID",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "permissions": ["users:read", "reports:read"],
  "scope": "openid profile email",
  "exp": 1700000000,
  "iat": 1699996400,
  "jti": "unique-token-id"
}
```

| Claim | Description |
|-------|-------------|
| `sub` | User ID (UUID) |
| `iss` | Issuer URL (CPI Auth backend) |
| `aud` | Audience (client ID of the application) |
| `tenant_id` | Tenant the user belongs to |
| `permissions` | Effective permissions after whitelist intersection |
| `scope` | Granted OAuth scopes |
| `exp` | Expiration timestamp |
| `iat` | Issued-at timestamp |
| `jti` | Unique token identifier |

### ID Token

The ID token is an OIDC-standard JWT containing identity claims:

```json
{
  "sub": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "iss": "http://localhost:5050",
  "aud": "YOUR_CLIENT_ID",
  "name": "Jane Smith",
  "email": "jane@example.com",
  "email_verified": true,
  "nonce": "RANDOM_NONCE",
  "exp": 1700000000,
  "iat": 1699996400
}
```

## OIDC Discovery

The discovery endpoint returns the full OpenID Connect configuration:

```bash
curl http://localhost:5050/.well-known/openid-configuration
```

```json
{
  "issuer": "http://localhost:5050",
  "authorization_endpoint": "http://localhost:5050/oauth/authorize",
  "token_endpoint": "http://localhost:5050/oauth/token",
  "userinfo_endpoint": "http://localhost:5050/oauth/userinfo",
  "jwks_uri": "http://localhost:5050/.well-known/jwks.json",
  "revocation_endpoint": "http://localhost:5050/oauth/revoke",
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "client_credentials", "refresh_token"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "profile", "email", "phone"],
  "code_challenge_methods_supported": ["S256"]
}
```

## JWKS Endpoint

Retrieve the public keys used to verify JWT signatures:

```bash
curl http://localhost:5050/.well-known/jwks.json
```

```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "key-id-1",
      "use": "sig",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

Use this endpoint to validate tokens in your backend services. Most JWT libraries can consume JWKS endpoints directly.

## UserInfo Endpoint

Retrieve the authenticated user's profile:

```bash
curl http://localhost:5050/oauth/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

## Passwordless Authentication

### Email Magic Link

```bash
# Start passwordless flow
curl -X POST http://localhost:5050/passwordless/start \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "client_id": "YOUR_CLIENT_ID",
    "redirect_uri": "http://localhost:3000/callback",
    "method": "email"
  }'

# Verify the code/link (after user clicks the email link)
curl -X POST http://localhost:5050/passwordless/verify \
  -H "Content-Type: application/json" \
  -d '{
    "code": "VERIFICATION_CODE",
    "client_id": "YOUR_CLIENT_ID",
    "redirect_uri": "http://localhost:3000/callback"
  }'
```

### SMS Magic Link

Same flow as email, but with `"method": "sms"` and a phone number instead of email.

## WebAuthn / FIDO2

### Registration

```bash
# Begin registration ceremony
curl -X POST http://localhost:5050/webauthn/register/begin \
  -H "Authorization: Bearer $TOKEN"

# Complete registration with authenticator response
curl -X POST http://localhost:5050/webauthn/register/finish \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "credential": { ... } }'
```

### Login

```bash
# Begin login ceremony
curl -X POST http://localhost:5050/webauthn/login/begin \
  -H "Content-Type: application/json" \
  -d '{ "email": "jane@example.com" }'

# Complete login with authenticator assertion
curl -X POST http://localhost:5050/webauthn/login/finish \
  -H "Content-Type: application/json" \
  -d '{ "assertion": { ... } }'
```

## SAML Federation

CPI Auth acts as a SAML Service Provider (SP).

### Metadata

```bash
curl http://localhost:5050/saml/metadata
```

Returns the SP metadata XML document to register with your Identity Provider (IdP).

### SP-Initiated SSO

```
GET http://localhost:5050/saml/sso
  ?idp=https://idp.example.com
  &relay_state=https://app.example.com/dashboard
```

### Assertion Consumer Service (ACS)

The IdP posts the SAML response to:

```
POST http://localhost:5050/saml/acs
```

After validating the assertion, CPI Auth creates or links the user and issues an OAuth token.

## Direct Login (Non-OAuth)

For simple integrations, CPI Auth provides direct login and registration endpoints:

```bash
# Login
curl -X POST http://localhost:5050/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "SecureP@ss123"
  }'

# Register
curl -X POST http://localhost:5050/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecureP@ss123",
    "name": "New User"
  }'
```

## Next Steps

- [RBAC](./rbac) -- Understand how permissions flow into tokens
- [MFA](./mfa) -- Multi-factor authentication flows
- [Applications](./applications) -- Configure token lifetimes and grant types
