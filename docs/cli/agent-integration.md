# AI Agent Integration Guide

This page provides a structured prompt for AI coding agents to integrate CPI Auth into any application. Copy the prompt below into your agent's context.

---

---

## Prompt

You are integrating **CPI Auth** into an application. CPI Auth is an OAuth 2.0 / OpenID Connect identity platform.

**CPI Auth Server:** `<CPI_AUTH_SERVER_URL>`
**Admin Credentials:** `<ADMIN_EMAIL>` / `<ADMIN_PASSWORD>`

### Step 1: Install the CLI (if not already installed)

```bash
npm install -g @cpi-auth/cli
```

Verify with `cpi-auth --version`. If the npm package is not published yet, you can skip the CLI and use the REST API directly (see Step 2b).

### Step 2a: Setup via CLI (preferred)

```bash
# Login
cpi-auth login --server <CPI_AUTH_SERVER_URL> --email <ADMIN_EMAIL> --password <ADMIN_PASSWORD>

# Verify connection
cpi-auth status

# Create application + users in one command
cpi-auth setup \
  --app-name "<APP_NAME>" \
  --app-type <spa|web|native|m2m> \
  --redirect-uri "<CALLBACK_URL>" \
  --allowed-origin "<APP_ORIGIN>" \
  --grant-types "authorization_code,refresh_token" \
  --create-user "<USER_EMAIL>" \
  --user-password "<USER_PASSWORD>" \
  --output env
```

The output gives you the `.env` variables to add to your application:

```
CPI_AUTH_CLIENT_ID=<generated>
CPI_AUTH_CLIENT_SECRET=<generated>        # only for web/m2m types
CPI_AUTH_ISSUER=<CPI_AUTH_SERVER_URL>
CPI_AUTH_REDIRECT_URI=<CALLBACK_URL>
CPI_AUTH_AUTHORIZATION_ENDPOINT=<CPI_AUTH_SERVER_URL>/oauth/authorize
CPI_AUTH_TOKEN_ENDPOINT=<CPI_AUTH_SERVER_URL>/oauth/token
CPI_AUTH_USERINFO_ENDPOINT=<CPI_AUTH_SERVER_URL>/oauth/userinfo
CPI_AUTH_JWKS_URI=<CPI_AUTH_SERVER_URL>/.well-known/jwks.json
```

### Step 2b: Setup via REST API (if CLI is not available)

```bash
# 1. Login to get an admin token
TOKEN=$(curl -s -X POST <CPI_AUTH_SERVER_URL>/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}' | jq -r '.access_token')

# 2. Create an application
curl -s -X POST <CPI_AUTH_SERVER_URL>/admin/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "<APP_NAME>",
    "type": "spa",
    "redirect_uris": ["<CALLBACK_URL>"],
    "allowed_origins": ["<APP_ORIGIN>"],
    "grant_types": ["authorization_code", "refresh_token"],
    "is_active": true
  }'
# Response contains client_id (and client_secret for web/m2m types)

# 3. Create a user (optional)
curl -s -X POST <CPI_AUTH_SERVER_URL>/admin/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"<USER_EMAIL>","password":"<USER_PASSWORD>","name":"<USER_NAME>"}'
```

### Step 3: Integrate into your application

CPI Auth is a **standard OAuth 2.0 / OIDC server**. Use any OAuth library for your framework.

#### React / Next.js

```bash
npm install oidc-client-ts react-oidc-context
```

```tsx
import { AuthProvider, useAuth } from 'react-oidc-context';

const oidcConfig = {
  authority: process.env.CPI_AUTH_ISSUER,       // e.g. https://auth.example.com
  client_id: process.env.CPI_AUTH_CLIENT_ID,
  redirect_uri: process.env.CPI_AUTH_REDIRECT_URI,
  scope: 'openid profile email',
  response_type: 'code',
};

function App() {
  return <AuthProvider {...oidcConfig}><YourApp /></AuthProvider>;
}

function YourApp() {
  const auth = useAuth();
  if (auth.isLoading) return <div>Loading...</div>;
  if (!auth.isAuthenticated) return <button onClick={() => auth.signinRedirect()}>Sign in</button>;
  return <div>Welcome, {auth.user?.profile.name}</div>;
}
```

#### Node.js / Express

```bash
npm install openid-client express-session
```

```js
const { Issuer } = require('openid-client');

const issuer = await Issuer.discover(process.env.CPI_AUTH_ISSUER);
const client = new issuer.Client({
  client_id: process.env.CPI_AUTH_CLIENT_ID,
  client_secret: process.env.CPI_AUTH_CLIENT_SECRET,
  redirect_uris: [process.env.CPI_AUTH_REDIRECT_URI],
  response_types: ['code'],
});

// Redirect to login
app.get('/login', (req, res) => {
  res.redirect(client.authorizationUrl({ scope: 'openid profile email' }));
});

// Handle callback
app.get('/callback', async (req, res) => {
  const tokenSet = await client.callback(process.env.CPI_AUTH_REDIRECT_URI, client.callbackParams(req));
  req.session.tokens = tokenSet;
  res.redirect('/');
});
```

#### Python (FastAPI / Django)

```bash
pip install authlib httpx
```

```python
from authlib.integrations.starlette_client import OAuth

oauth = OAuth()
oauth.register(
    name='cpiauth',
    server_metadata_url=f'{CPI_AUTH_ISSUER}/.well-known/openid-configuration',
    client_id=CPI_AUTH_CLIENT_ID,
    client_secret=CPI_AUTH_CLIENT_SECRET,
    client_kwargs={'scope': 'openid profile email'},
)
```

#### Any language (raw PKCE flow)

1. **Discover:** `GET <CPI_AUTH_SERVER_URL>/.well-known/openid-configuration`

2. **Generate PKCE:** Create `code_verifier` (43-128 random chars), SHA256 hash it to `code_challenge`

3. **Authorize:** Redirect user to:
   ```
   <CPI_AUTH_SERVER_URL>/oauth/authorize?
     client_id=<CLIENT_ID>&
     redirect_uri=<CALLBACK_URL>&
     response_type=code&
     scope=openid+profile+email&
     code_challenge=<CODE_CHALLENGE>&
     code_challenge_method=S256&
     state=<RANDOM_STATE>
   ```

4. **Exchange code:** `POST <CPI_AUTH_SERVER_URL>/oauth/token`
   ```
   grant_type=authorization_code&
   code=<AUTH_CODE>&
   redirect_uri=<CALLBACK_URL>&
   client_id=<CLIENT_ID>&
   code_verifier=<CODE_VERIFIER>
   ```

5. **Get user info:** `GET <CPI_AUTH_SERVER_URL>/oauth/userinfo` with `Authorization: Bearer <access_token>`

6. **Validate JWTs:** Fetch keys from `<CPI_AUTH_SERVER_URL>/.well-known/jwks.json`, verify RS256 signature

### Step 4: Additional management (optional)

If the CLI is installed:

```bash
# Users
cpi-auth users create --email user@example.com --password SecurePass123 --name "John Doe"
cpi-auth users list --search "john"
cpi-auth users block <USER_ID>

# Roles & Permissions
cpi-auth roles create --name editor --permissions "posts:read,posts:write"
cpi-auth roles create-permission --name "billing:manage"
cpi-auth roles list

# Applications
cpi-auth apps list
cpi-auth apps rotate-secret <APP_ID>

# Context management (multiple servers)
cpi-auth config add-context staging --server https://staging-auth.example.com
cpi-auth config use-context staging
cpi-auth config list-contexts
```

If the CLI is not available, all operations are available via REST API with Bearer token authentication at `/admin/*` endpoints.

### JWT Token Structure

Access tokens are RS256-signed JWTs:

```json
{
  "iss": "<CPI_AUTH_SERVER_URL>",
  "sub": "user-uuid",
  "aud": ["client-id"],
  "exp": 1234567890,
  "tenant_id": "tenant-uuid",
  "email": "user@example.com",
  "name": "User Name",
  "scope": "openid profile email",
  "permissions": ["posts:read", "posts:write"],
  "act": { "sub": "admin-uuid" }
}
```

| Claim | Description |
|-------|-------------|
| `sub` | User ID (UUID) |
| `iss` | Issuer URL (your CPI Auth server) |
| `aud` | Client ID of the application |
| `tenant_id` | Tenant this user belongs to |
| `permissions` | RBAC permissions array |
| `act` | Only present during impersonation; contains admin's user ID |

Validate tokens by fetching public keys from `<CPI_AUTH_SERVER_URL>/.well-known/jwks.json` and verifying the RS256 signature.

### API Endpoints Reference

| Category | Method | Endpoint | Auth |
|----------|--------|----------|------|
| Discovery | GET | `/.well-known/openid-configuration` | Public |
| JWKS | GET | `/.well-known/jwks.json` | Public |
| Authorize | GET | `/oauth/authorize` | Public (browser redirect) |
| Token | POST | `/oauth/token` | Public |
| Revoke | POST | `/oauth/revoke` | Public |
| Userinfo | GET | `/oauth/userinfo` | Bearer token |
| Login | POST | `/api/v1/auth/login` | Public |
| Register | POST | `/api/v1/auth/register` | Public |
| Logout | GET/POST | `/api/v1/auth/logout` | Session cookie |
| Admin Login | POST | `/admin/auth/login` | Public |
| Admin Me | GET | `/admin/auth/me` | Bearer token |
| Users | GET/POST | `/admin/users` | Bearer token |
| Applications | GET/POST | `/admin/applications` | Bearer token |
| Roles | GET/POST | `/admin/roles` | Bearer token |
| Permissions | GET/POST | `/admin/permissions` | Bearer token |

### Session Behavior

- After login, CPI Auth sets an `HttpOnly` session cookie (`__cpi_auth_session`)
- Default session: 24 hours. With "Remember me": 30 days
- Subsequent OAuth authorizations are auto-approved if the session is valid (SSO)
- Logout via `GET/POST /api/v1/auth/logout` clears the session cookie
