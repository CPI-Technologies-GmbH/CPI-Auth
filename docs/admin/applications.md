# Applications Management

The Applications section of the Admin Console lets you register and configure OAuth 2.0 clients. Each application represents an external service that authenticates users through CPI Auth.

## Application List

The main page displays all registered applications in a card or table layout.

### Table Columns

| Column | Description |
|--------|-------------|
| **Name** | Application display name |
| **Type** | Badge: `SPA` (blue), `Web` (green), `Native` (purple), `M2M` (orange) |
| **Client ID** | The public client identifier |
| **Status** | Active (green dot) or Inactive (gray dot) |
| **Created** | Registration date |

Click any application to open its detail page.

### Create Application

Click **"Create Application"** to open the creation dialog:

1. Enter a **name** and optional **description**
2. Select the **type** (SPA, Web, Native, M2M)
3. The form adjusts based on type (e.g., no redirect URIs for M2M)
4. Click **Create** to register the application

For `web` and `m2m` types, a client secret is generated and displayed once. Copy it immediately as it cannot be retrieved again.

---

## Application Detail Page

The detail page is organized into tabs.

### Overview Tab

The overview tab displays key application identifiers and quick configuration.

#### Client Credentials

| Field | Description |
|-------|-------------|
| **Client ID** | Public identifier (copyable) |
| **Client Secret** | Shown masked, with a copy button. Only available for `web` and `m2m` types. |

A **"Rotate Secret"** button opens a confirmation dialog to generate a new secret.

#### Status Toggle

A toggle switch to enable or disable the application. Disabled applications reject all authentication requests.

#### Grant Types

Checkboxes to enable or disable supported grant types:

| Grant Type | Applicable Types |
|------------|-----------------|
| Authorization Code | SPA, Web, Native |
| Refresh Token | SPA, Web, Native |
| Client Credentials | M2M |

Changes to grant types take effect immediately after saving.

### Settings Tab

#### Redirect URIs

A list editor for managing allowed redirect URIs. Each URI must be an exact match (no wildcards).

```
http://localhost:3000/callback          [Remove]
https://myapp.com/callback              [Remove]
https://staging.myapp.com/callback      [Remove]

[+ Add Redirect URI]
```

#### Allowed Origins

CORS origins permitted to make browser requests to the OAuth endpoints. Required for SPA applications.

```
http://localhost:3000                    [Remove]
https://myapp.com                       [Remove]

[+ Add Origin]
```

#### Logout URLs

Post-logout redirect URLs. After logout, the user is redirected to one of these URLs.

```
http://localhost:3000                    [Remove]
https://myapp.com                       [Remove]

[+ Add Logout URL]
```

#### Token TTLs

Configurable time-to-live values for each token type:

| Token | Default | Description |
|-------|---------|-------------|
| **Access Token TTL** | 3600s (1 hour) | Duration of access token validity |
| **Refresh Token TTL** | 604800s (7 days) | Duration of refresh token validity |
| **ID Token TTL** | 3600s (1 hour) | Duration of ID token validity |

Each has a numeric input with seconds as the unit. Common presets are available (15 min, 1 hour, 1 day, 7 days, 30 days).

### Permissions Tab

A checkbox grid showing all permissions defined in the tenant. Checked permissions form the application's permission whitelist.

```
 [x] users:read          [ ] users:write
 [x] posts:read          [x] posts:write
 [ ] settings:read       [ ] settings:write
 [x] reports:generate    [ ] billing:manage
```

When a whitelist is set, tokens issued to this application only include permissions that are both assigned to the user and present in this whitelist. If no permissions are checked, all user permissions are included (no filtering).

### Connections Tab

Shows which authentication methods and social connections are enabled for this application. Typically configured at the tenant level, but can be overridden per application.

### API Tab

Provides quick-start code snippets for integrating the application:

#### JavaScript/TypeScript

```javascript
const config = {
  clientId: 'app_k8f2m9x1',
  redirectUri: 'http://localhost:3000/callback',
  authorizationEndpoint: 'http://localhost:5054/oauth/authorize',
  tokenEndpoint: 'http://localhost:5054/oauth/token',
  scope: 'openid profile email'
};
```

#### curl

```bash
# Authorization Code + PKCE flow
curl -X POST http://localhost:5054/oauth/authorize \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "app_k8f2m9x1",
    "redirect_uri": "http://localhost:3000/callback",
    "response_type": "code",
    "scope": "openid profile email",
    "code_challenge": "{generated_challenge}",
    "code_challenge_method": "S256"
  }'
```

---

## Secret Rotation Dialog

When rotating a client secret:

1. Click **"Rotate Secret"** on the Overview tab
2. A confirmation dialog warns that the current secret will be invalidated immediately
3. Confirm to generate the new secret
4. The new secret is displayed once in the dialog with a copy button
5. All services using the old secret must be updated

::: warning
Secret rotation is immediate and irreversible. Ensure all consuming services are ready to use the new secret before rotating.
:::

---

## Application Type Guidelines

### SPA (Single Page Application)

- No client secret (public client)
- PKCE required for all authorization flows
- Configure allowed origins for CORS
- Set redirect URIs for callback handling

### Web (Server-Side Application)

- Has a client secret (confidential client)
- Secret stored securely on the server
- PKCE optional but recommended
- Redirect URIs must use HTTPS in production

### Native (Mobile / Desktop)

- No client secret (public client)
- PKCE required
- Use custom URL schemes for redirect URIs (e.g., `myapp://callback`)
- No allowed origins needed

### M2M (Machine-to-Machine)

- Has a client secret (confidential client)
- Uses client credentials grant only
- No redirect URIs or allowed origins
- Typically has a longer access token TTL
- No user context in tokens
