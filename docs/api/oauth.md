# OAuth 2.0 & OpenID Connect

CPI Auth implements OAuth 2.0 and OpenID Connect (OIDC) with PKCE support. This page covers the full authorization flow, token management, and discovery endpoints.

## Base URL

```
http://localhost:5054
```

---

## OIDC Discovery

### GET /.well-known/openid-configuration

Returns the OpenID Connect discovery document with all supported endpoints and capabilities.

```bash
curl http://localhost:5054/.well-known/openid-configuration
```

**Response 200 OK:**

```json
{
  "issuer": "http://localhost:5054",
  "authorization_endpoint": "http://localhost:5054/oauth/authorize",
  "token_endpoint": "http://localhost:5054/oauth/token",
  "userinfo_endpoint": "http://localhost:5054/oauth/userinfo",
  "jwks_uri": "http://localhost:5054/.well-known/jwks.json",
  "revocation_endpoint": "http://localhost:5054/oauth/revoke",
  "response_types_supported": ["code"],
  "grant_types_supported": [
    "authorization_code",
    "refresh_token",
    "client_credentials"
  ],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "profile", "email", "offline_access"],
  "token_endpoint_auth_methods_supported": [
    "client_secret_post",
    "client_secret_basic",
    "none"
  ],
  "code_challenge_methods_supported": ["S256"]
}
```

---

## JWKS

### GET /.well-known/jwks.json

Returns the JSON Web Key Set used to verify JWT signatures.

```bash
curl http://localhost:5054/.well-known/jwks.json
```

**Response 200 OK:**

```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "cpi-auth-key-1",
      "use": "sig",
      "alg": "RS256",
      "n": "0vx7agoebGcQSuu...",
      "e": "AQAB"
    }
  ]
}
```

---

## Authorization

### POST /oauth/authorize

Initiates the authorization code flow. For SPAs and native apps, PKCE is required.

**Request:** Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `client_id` | Yes | Application client ID |
| `redirect_uri` | Yes | Must match a registered redirect URI |
| `response_type` | Yes | Must be `code` |
| `scope` | Yes | Space-separated scopes (e.g., `openid profile email`) |
| `state` | Recommended | CSRF protection random string |
| `code_challenge` | Yes (PKCE) | Base64url-encoded SHA256 hash of code_verifier |
| `code_challenge_method` | Yes (PKCE) | Must be `S256` |

### curl Example

```bash
curl -X POST http://localhost:5054/oauth/authorize \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "app_abc123",
    "redirect_uri": "http://localhost:3000/callback",
    "response_type": "code",
    "scope": "openid profile email",
    "state": "random-csrf-state-value",
    "code_challenge": "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
    "code_challenge_method": "S256"
  }'
```

**Response 200 OK:**

```json
{
  "code": "auth_code_abc123xyz",
  "state": "random-csrf-state-value",
  "redirect_uri": "http://localhost:3000/callback?code=auth_code_abc123xyz&state=random-csrf-state-value"
}
```

---

## Token Exchange

### POST /oauth/token

Exchange an authorization code for tokens, or refresh an existing token.

### Authorization Code Grant

```bash
curl -X POST http://localhost:5054/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "authorization_code",
    "code": "auth_code_abc123xyz",
    "redirect_uri": "http://localhost:3000/callback",
    "client_id": "app_abc123",
    "code_verifier": "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
  }'
```

**Response 200 OK:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_value",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "openid profile email"
}
```

### Refresh Token Grant

```bash
curl -X POST http://localhost:5054/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "refresh_token",
    "refresh_token": "refresh_token_value",
    "client_id": "app_abc123"
  }'
```

### Client Credentials Grant (M2M)

```bash
curl -X POST http://localhost:5054/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "m2m_app_id",
    "client_secret": "m2m_app_secret",
    "scope": "users:read users:write"
  }'
```

---

## Token Revocation

### POST /oauth/revoke

Revoke an access token or refresh token.

```bash
curl -X POST http://localhost:5054/oauth/revoke \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJSUzI1NiIs...",
    "token_type_hint": "access_token"
  }'
```

**Response 200 OK:**

```json
{
  "revoked": true
}
```

---

## Userinfo

### GET /oauth/userinfo

Returns claims about the authenticated user. Requires a valid access token with `openid` scope.

```bash
curl http://localhost:5054/oauth/userinfo \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIs..."
```

**Response 200 OK:**

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "email_verified": true,
  "name": "Jane Doe",
  "given_name": "Jane",
  "family_name": "Doe",
  "locale": "en",
  "updated_at": 1711612800
}
```

---

## JWT Token Claims

### Access Token

```json
{
  "iss": "http://localhost:5054",
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "aud": "app_abc123",
  "exp": 1711616400,
  "iat": 1711612800,
  "scope": "openid profile email",
  "tenant_id": "tenant-uuid",
  "permissions": ["users:read", "posts:write"]
}
```

### ID Token

```json
{
  "iss": "http://localhost:5054",
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "aud": "app_abc123",
  "exp": 1711616400,
  "iat": 1711612800,
  "nonce": "random-nonce-value",
  "email": "user@example.com",
  "email_verified": true,
  "name": "Jane Doe"
}
```

**Permission model:** Token permissions are the intersection of the user's effective permissions and the application's permission whitelist. If the application has an empty whitelist, all user permissions are included.

---

## Full PKCE Flow Example

Here is a complete end-to-end PKCE authorization code flow using curl and standard CLI tools.

### Step 1: Generate PKCE Values

```bash
# Generate code_verifier (43-128 characters, URL-safe)
CODE_VERIFIER=$(openssl rand -base64 32 | tr -d '=/+' | head -c 43)

# Generate code_challenge (S256)
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -sha256 -binary | openssl base64 | tr '+/' '-_' | tr -d '=')

echo "code_verifier: $CODE_VERIFIER"
echo "code_challenge: $CODE_CHALLENGE"
```

### Step 2: Request Authorization Code

```bash
STATE=$(openssl rand -hex 16)

curl -X POST http://localhost:5054/oauth/authorize \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"app_abc123\",
    \"redirect_uri\": \"http://localhost:3000/callback\",
    \"response_type\": \"code\",
    \"scope\": \"openid profile email offline_access\",
    \"state\": \"$STATE\",
    \"code_challenge\": \"$CODE_CHALLENGE\",
    \"code_challenge_method\": \"S256\"
  }"
```

### Step 3: Exchange Code for Tokens

```bash
AUTH_CODE="auth_code_from_step_2"

curl -X POST http://localhost:5054/oauth/token \
  -H "Content-Type: application/json" \
  -d "{
    \"grant_type\": \"authorization_code\",
    \"code\": \"$AUTH_CODE\",
    \"redirect_uri\": \"http://localhost:3000/callback\",
    \"client_id\": \"app_abc123\",
    \"code_verifier\": \"$CODE_VERIFIER\"
  }"
```

### Step 4: Use the Access Token

```bash
ACCESS_TOKEN="token_from_step_3"

curl http://localhost:5054/oauth/userinfo \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Step 5: Refresh When Expired

```bash
REFRESH_TOKEN="refresh_token_from_step_3"

curl -X POST http://localhost:5054/oauth/token \
  -H "Content-Type: application/json" \
  -d "{
    \"grant_type\": \"refresh_token\",
    \"refresh_token\": \"$REFRESH_TOKEN\",
    \"client_id\": \"app_abc123\"
  }"
```
