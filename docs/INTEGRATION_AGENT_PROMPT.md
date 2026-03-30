# CPI Auth Integration Agent Prompt

Use this prompt to instruct an AI agent to integrate CPI Auth into an existing application and configure it via CLI.

---

## Prompt

You are integrating CPI Auth (an OAuth 2.0 / OpenID Connect identity platform) into an application. CPI Auth is already deployed at `https://auth.cpi.dev`. You have CLI access on this computer.

### Step 1: Install the CLI and login

```bash
# The CLI is at /Users/maxi/Coding/authforge/tools/cli/dist/cli.js
# Or install globally: npm install -g @cpi-auth/cli
node /Users/maxi/Coding/authforge/tools/cli/dist/cli.js login \
  --server https://auth.cpi.dev \
  --email admin@cpi-auth.local \
  --password <ADMIN_PASSWORD>
```

### Step 2: Create the application and configure everything in one command

```bash
node /Users/maxi/Coding/authforge/tools/cli/dist/cli.js setup \
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

This outputs `.env` variables you need:
- `CPI_AUTH_CLIENT_ID` — Your OAuth client ID
- `CPI_AUTH_CLIENT_SECRET` — Your OAuth client secret (for web/m2m types)
- `CPI_AUTH_ISSUER` — The issuer URL (https://auth.cpi.dev)
- `CPI_AUTH_REDIRECT_URI` — Your callback URL
- `CPI_AUTH_AUTHORIZATION_ENDPOINT` — https://auth.cpi.dev/oauth/authorize
- `CPI_AUTH_TOKEN_ENDPOINT` — https://auth.cpi.dev/oauth/token
- `CPI_AUTH_USERINFO_ENDPOINT` — https://auth.cpi.dev/oauth/userinfo
- `CPI_AUTH_JWKS_URI` — https://auth.cpi.dev/.well-known/jwks.json

### Step 3: Integrate into your application

#### For React/Next.js (SPA):
Use `oidc-client-ts` or `react-oidc-context`:
```bash
npm install oidc-client-ts react-oidc-context
```

```tsx
import { AuthProvider } from 'react-oidc-context';

const oidcConfig = {
  authority: process.env.CPI_AUTH_ISSUER,
  client_id: process.env.CPI_AUTH_CLIENT_ID,
  redirect_uri: process.env.CPI_AUTH_REDIRECT_URI,
  scope: 'openid profile email',
  response_type: 'code',
};

<AuthProvider {...oidcConfig}>
  <App />
</AuthProvider>
```

#### For Node.js/Express (Web):
Use `openid-client`:
```bash
npm install openid-client
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
```

#### For any language:
CPI Auth is a standard OAuth 2.0 / OIDC server. Use any OAuth library:
1. Discover endpoints: `GET https://auth.cpi.dev/.well-known/openid-configuration`
2. Authorization: Redirect to `https://auth.cpi.dev/oauth/authorize?client_id=...&redirect_uri=...&response_type=code&scope=openid+profile+email&code_challenge=...&code_challenge_method=S256`
3. Token exchange: `POST https://auth.cpi.dev/oauth/token` with `grant_type=authorization_code&code=...&code_verifier=...`
4. Userinfo: `GET https://auth.cpi.dev/oauth/userinfo` with `Authorization: Bearer <token>`
5. Validate JWT: Fetch public keys from `https://auth.cpi.dev/.well-known/jwks.json`

### Step 4: Additional CLI commands (optional)

```bash
# Create additional users
cpi-auth users create --email user@example.com --password SecurePass --name "John Doe"

# Create custom roles
cpi-auth roles create --name editor --permissions "posts:read,posts:write"

# List all applications
cpi-auth apps list

# Rotate client secret (for web/m2m apps)
cpi-auth apps rotate-secret <APP_ID>

# Manage language strings for login pages
cpi-auth strings list --locale en
cpi-auth strings add "login.welcome" "Welcome to Our App" --locale en
```

### Available API Endpoints

| Category | Endpoints |
|----------|-----------|
| OAuth/OIDC | `/oauth/authorize`, `/oauth/token`, `/oauth/revoke`, `/oauth/userinfo` |
| Discovery | `/.well-known/openid-configuration`, `/.well-known/jwks.json` |
| Admin Auth | `/admin/auth/login`, `/admin/auth/me`, `/admin/auth/refresh` |
| Users | `/admin/users` (CRUD + block/unblock/sessions/roles) |
| Applications | `/admin/applications` (CRUD + permissions + rotate-secret) |
| Roles | `/admin/roles`, `/admin/permissions` (CRUD) |
| Templates | `/admin/page-templates`, `/admin/language-strings` |

### JWT Token Structure

Access tokens are RS256-signed JWTs with claims:
```json
{
  "iss": "https://auth.cpi.dev",
  "sub": "user-uuid",
  "aud": ["client-id"],
  "exp": 1234567890,
  "tenant_id": "tenant-uuid",
  "email": "user@example.com",
  "name": "User Name",
  "scope": "openid profile email",
  "permissions": ["posts:read", "posts:write"],
  "act": { "sub": "admin-uuid" }  // only present if impersonated
}
```

Validate with JWKS at `https://auth.cpi.dev/.well-known/jwks.json`.
