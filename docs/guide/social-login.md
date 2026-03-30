# Social Login

CPI Auth supports federated authentication through popular social identity providers. Users can sign in with their existing accounts, reducing friction and improving conversion.

## Supported Providers

| Provider | Code | OAuth/OIDC |
|----------|------|------------|
| Google | `google` | OIDC |
| GitHub | `github` | OAuth 2.0 |
| Microsoft | `microsoft` | OIDC |
| Apple | `apple` | OIDC |
| Facebook | `facebook` | OAuth 2.0 |
| Twitter | `twitter` | OAuth 2.0 |

## Configuring a Social Provider

Social providers are configured per tenant through the Admin UI or API.

### Google

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create an OAuth 2.0 Client ID
3. Set the authorized redirect URI to: `http://localhost:5050/auth/callback/google`
4. Configure in CPI Auth:

```bash
curl -X POST http://localhost:5050/api/v1/connections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "google",
    "client_id": "YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com",
    "client_secret": "YOUR_GOOGLE_CLIENT_SECRET",
    "scopes": ["openid", "profile", "email"],
    "enabled": true
  }'
```

### GitHub

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Create a new OAuth App
3. Set the callback URL to: `http://localhost:5050/auth/callback/github`

```bash
curl -X POST http://localhost:5050/api/v1/connections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "github",
    "client_id": "YOUR_GITHUB_CLIENT_ID",
    "client_secret": "YOUR_GITHUB_CLIENT_SECRET",
    "scopes": ["user:email"],
    "enabled": true
  }'
```

### Microsoft

1. Go to the [Azure Portal](https://portal.azure.com/) > App registrations
2. Create a new registration
3. Add a redirect URI: `http://localhost:5050/auth/callback/microsoft`

```bash
curl -X POST http://localhost:5050/api/v1/connections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "microsoft",
    "client_id": "YOUR_AZURE_CLIENT_ID",
    "client_secret": "YOUR_AZURE_CLIENT_SECRET",
    "scopes": ["openid", "profile", "email"],
    "enabled": true
  }'
```

### Apple

1. Go to the [Apple Developer Portal](https://developer.apple.com/)
2. Create a Services ID and configure Sign In with Apple
3. Register the redirect URI: `http://localhost:5050/auth/callback/apple`

```bash
curl -X POST http://localhost:5050/api/v1/connections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "apple",
    "client_id": "YOUR_APPLE_SERVICES_ID",
    "client_secret": "YOUR_APPLE_KEY",
    "scopes": ["name", "email"],
    "enabled": true
  }'
```

### Facebook

1. Go to [Meta for Developers](https://developers.facebook.com/)
2. Create an app and add Facebook Login
3. Set the redirect URI: `http://localhost:5050/auth/callback/facebook`

```bash
curl -X POST http://localhost:5050/api/v1/connections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "facebook",
    "client_id": "YOUR_FACEBOOK_APP_ID",
    "client_secret": "YOUR_FACEBOOK_APP_SECRET",
    "scopes": ["email", "public_profile"],
    "enabled": true
  }'
```

### Twitter

1. Go to the [Twitter Developer Portal](https://developer.twitter.com/)
2. Create a project and app with OAuth 2.0 enabled
3. Set the callback URL: `http://localhost:5050/auth/callback/twitter`

```bash
curl -X POST http://localhost:5050/api/v1/connections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "twitter",
    "client_id": "YOUR_TWITTER_CLIENT_ID",
    "client_secret": "YOUR_TWITTER_CLIENT_SECRET",
    "scopes": ["tweet.read", "users.read"],
    "enabled": true
  }'
```

## Account Linking

When a user signs in with a social provider, CPI Auth checks if an account with that email already exists:

- **No existing account** -- A new user is created with the social identity linked
- **Existing account, same email** -- The social identity is linked to the existing account (if email is verified on both sides)
- **Existing account, different email** -- A new user is created

Users can link multiple social providers to a single account. This is visible in the user's identities:

```bash
curl http://localhost:5050/api/v1/users/{user_id}/identities \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
[
  {
    "id": "i1d2e3f4-a5b6-7890-cdef-012345678901",
    "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "provider": "google",
    "provider_user_id": "118234567890123456789",
    "profile": {
      "name": "Jane Smith",
      "email": "jane@example.com",
      "picture": "https://lh3.googleusercontent.com/..."
    },
    "created_at": "2025-01-15T10:30:00Z"
  },
  {
    "id": "i2d3e4f5-b6c7-8901-defg-123456789012",
    "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "provider": "github",
    "provider_user_id": "12345678",
    "profile": {
      "login": "janesmith",
      "name": "Jane Smith"
    },
    "created_at": "2025-02-01T08:00:00Z"
  }
]
```

## Identity Model

Each social login creates an `Identity` record:

| Field | Description |
|-------|-------------|
| `id` | UUID of the identity record |
| `user_id` | The CPI Auth user this identity belongs to |
| `provider` | Provider name (google, github, etc.) |
| `provider_user_id` | The user's ID at the provider |
| `profile` | JSON profile data from the provider |
| `tokens_encrypted` | Encrypted OAuth tokens (not exposed via API) |

Provider tokens are encrypted at rest using the configured encryption key. They can be used to make API calls to the provider on behalf of the user.

## Login Flow with Social Providers

1. User clicks "Sign in with Google" on the Login UI
2. Browser redirects to Google's OAuth consent screen
3. User authorizes the application
4. Google redirects back to CPI Auth with an authorization code
5. CPI Auth exchanges the code for tokens and fetches the user profile
6. CPI Auth creates or links the user account
7. CPI Auth issues its own access/refresh/ID tokens
8. User is redirected back to the application

## Disabling a Provider

```bash
curl -X PATCH http://localhost:5050/api/v1/connections/{connection_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": false
  }'
```

Disabling a provider prevents new logins via that provider but does not unlink existing identities.

## Next Steps

- [Auth Flows](./auth-flows) -- Full authentication flow documentation
- [Users](./users) -- User management and identity linking
- [Custom Domains](./custom-domains) -- Custom login domains for social redirect URIs
