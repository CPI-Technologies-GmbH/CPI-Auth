# AI Agent Integration Guide

This page provides a structured prompt and reference for AI coding agents (Claude, GPT, Copilot, Cursor, etc.) to integrate CPI Auth into any application using only the CLI — no UI required.

## Agent Prompt

Copy and paste the following into your AI agent's context when you want it to integrate CPI Auth into your project:

---

### Context

You are integrating **CPI Auth** (an OAuth 2.0 / OpenID Connect identity platform) into an application. CPI Auth is deployed and accessible. You have CLI access on this computer.

**CPI Auth server:** `https://auth.cpi.dev` (replace with your instance URL)

### Step 1: Authenticate with the CLI

```bash
cpi-auth login \
  --server https://auth.cpi.dev \
  --email admin@cpi-auth.local \
  --password <ADMIN_PASSWORD>
```

The token is saved locally to `.cpi-auth-token` and used for all subsequent commands.

### Step 2: Create the application in one command

```bash
cpi-auth setup \
  --server https://auth.cpi.dev \
  --app-name "<YOUR_APP_NAME>" \
  --app-type <spa|web|native|m2m> \
  --redirect-uri "<YOUR_CALLBACK_URL>" \
  --allowed-origin "<YOUR_APP_ORIGIN>" \
  --grant-types "authorization_code,refresh_token" \
  --create-user "<FIRST_USER_EMAIL>" \
  --user-password "<SECURE_PASSWORD>" \
  --output env
```

**Application types:**

| Type | Use case | Has client secret? |
|------|----------|-------------------|
| `spa` | React, Vue, Angular (browser) | No |
| `web` | Node.js, Django, Rails (server) | Yes |
| `native` | iOS, Android, Desktop | No |
| `m2m` | Backend services, APIs | Yes |

**Output variables** (copy to your `.env`):

```bash
CPI_AUTH_CLIENT_ID=<generated>
CPI_AUTH_CLIENT_SECRET=<generated>      # only for web/m2m
CPI_AUTH_ISSUER=https://auth.cpi.dev
CPI_AUTH_REDIRECT_URI=<your-callback>
CPI_AUTH_AUTHORIZATION_ENDPOINT=https://auth.cpi.dev/oauth/authorize
CPI_AUTH_TOKEN_ENDPOINT=https://auth.cpi.dev/oauth/token
CPI_AUTH_USERINFO_ENDPOINT=https://auth.cpi.dev/oauth/userinfo
CPI_AUTH_JWKS_URI=https://auth.cpi.dev/.well-known/jwks.json
```

### Step 3: Integrate into your application

#### React / Next.js (SPA)

```bash
npm install oidc-client-ts react-oidc-context
```

```tsx
import { AuthProvider, useAuth } from 'react-oidc-context';

const oidcConfig = {
  authority: process.env.CPI_AUTH_ISSUER,
  client_id: process.env.CPI_AUTH_CLIENT_ID,
  redirect_uri: process.env.CPI_AUTH_REDIRECT_URI,
  scope: 'openid profile email',
  response_type: 'code',
};

function App() {
  return (
    <AuthProvider {...oidcConfig}>
      <YourApp />
    </AuthProvider>
  );
}

function YourApp() {
  const auth = useAuth();

  if (auth.isLoading) return <div>Loading...</div>;
  if (auth.error) return <div>Error: {auth.error.message}</div>;

  if (!auth.isAuthenticated) {
    return <button onClick={() => auth.signinRedirect()}>Sign in</button>;
  }

  return (
    <div>
      <p>Welcome, {auth.user?.profile.name}</p>
      <p>Email: {auth.user?.profile.email}</p>
      <button onClick={() => auth.signoutRedirect()}>Sign out</button>
    </div>
  );
}
```

#### Node.js / Express (Server)

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

// Login route
app.get('/login', (req, res) => {
  const url = client.authorizationUrl({ scope: 'openid profile email' });
  res.redirect(url);
});

// Callback route
app.get('/callback', async (req, res) => {
  const params = client.callbackParams(req);
  const tokenSet = await client.callback(process.env.CPI_AUTH_REDIRECT_URI, params);
  req.session.tokens = tokenSet;
  res.redirect('/');
});
```

#### Python / Django / FastAPI

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

#### Any language (raw OAuth2 PKCE flow)

1. **Discover endpoints:**
   `GET https://auth.cpi.dev/.well-known/openid-configuration`

2. **Generate PKCE challenge:**
   Create random `code_verifier` (43-128 chars), hash with SHA256 to get `code_challenge`

3. **Redirect to authorize:**
   ```
   https://auth.cpi.dev/oauth/authorize?
     client_id=<CLIENT_ID>&
     redirect_uri=<CALLBACK_URL>&
     response_type=code&
     scope=openid+profile+email&
     code_challenge=<CODE_CHALLENGE>&
     code_challenge_method=S256&
     state=<RANDOM_STATE>
   ```

4. **Exchange code for tokens:**
   ```
   POST https://auth.cpi.dev/oauth/token
   Content-Type: application/x-www-form-urlencoded

   grant_type=authorization_code&
   code=<AUTH_CODE>&
   redirect_uri=<CALLBACK_URL>&
   client_id=<CLIENT_ID>&
   code_verifier=<CODE_VERIFIER>
   ```

5. **Get user info:**
   ```
   GET https://auth.cpi.dev/oauth/userinfo
   Authorization: Bearer <ACCESS_TOKEN>
   ```

6. **Validate JWT tokens:**
   Fetch public keys from `https://auth.cpi.dev/.well-known/jwks.json` and verify RS256 signature.

### Step 4: Additional CLI commands

```bash
# User management
cpi-auth users create --email user@example.com --password Pass123 --name "John"
cpi-auth users list --search "john"
cpi-auth users block <USER_ID>

# Role management
cpi-auth roles create --name editor --permissions "posts:read,posts:write"
cpi-auth roles list

# Application management
cpi-auth apps list
cpi-auth apps rotate-secret <APP_ID>

# Permissions
cpi-auth roles create-permission --name "billing:manage"
cpi-auth roles permissions

# Language strings (for login page customization)
cpi-auth strings add "login.welcome" "Welcome to Our App" --locale en
cpi-auth strings add "login.welcome" "Willkommen" --locale de
```

## JWT Token Structure

Access tokens are RS256-signed JWTs:

```json
{
  "iss": "https://auth.cpi.dev",
  "sub": "user-uuid",
  "aud": ["client-id"],
  "exp": 1234567890,
  "iat": 1234567800,
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
| `iss` | Issuer URL |
| `aud` | Client ID of the application |
| `tenant_id` | Tenant this user belongs to |
| `permissions` | RBAC permissions (intersection of user and app permissions) |
| `act` | Present only during admin impersonation; contains the admin's user ID |

## API Endpoints Reference

| Category | Method | Path | Description |
|----------|--------|------|-------------|
| **Discovery** | GET | `/.well-known/openid-configuration` | OIDC discovery document |
| | GET | `/.well-known/jwks.json` | JWT signing public keys |
| **OAuth** | POST | `/oauth/authorize` | Authorization endpoint |
| | POST | `/oauth/token` | Token exchange |
| | POST | `/oauth/revoke` | Token revocation |
| | GET | `/oauth/userinfo` | User profile endpoint |
| **Auth** | POST | `/api/v1/auth/login` | Login (email + password) |
| | POST | `/api/v1/auth/register` | Register new user |
| **Admin** | POST | `/admin/auth/login` | Admin login |
| | GET | `/admin/auth/me` | Current admin user |
| | GET | `/admin/users` | List users |
| | POST | `/admin/users` | Create user |
| | GET | `/admin/applications` | List applications |
| | POST | `/admin/applications` | Create application |
| | GET | `/admin/roles` | List roles |
| | POST | `/admin/roles` | Create role |
| | GET | `/admin/permissions` | List permissions |
| | POST | `/admin/permissions` | Create permission |

Full API reference: [API Documentation](/api/authentication)
