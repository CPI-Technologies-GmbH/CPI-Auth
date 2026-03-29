# Getting Started

This guide walks you through running CPI Auth locally and creating your first application.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Docker & Docker Compose | 20.10+ | Running all services |
| Go | 1.22+ | Backend development (optional) |
| Node.js | 20+ | Frontend development (optional) |
| Git | Any | Cloning the repository |

Docker is the only hard requirement for running CPI Auth. Go and Node.js are needed only if you plan to develop against the source.

## Quick Start

Clone the repository and start all services:

```bash
git clone https://github.com/cpi-auth/cpi-auth.git
cd cpi-auth
docker compose up -d
```

Wait for all containers to become healthy (about 30 seconds on first run):

```bash
docker compose ps
```

You should see all services in the `running (healthy)` state.

## Services and Ports

Once running, the following services are available on your machine:

| Service | URL | Port |
|---------|-----|------|
| Backend API | `http://localhost:5050` | 5050 |
| Login UI | `http://localhost:5053` | 5053 |
| Admin UI | `http://localhost:5054` | 5054 |
| Account UI | `http://localhost:5055` | 5055 |
| PostgreSQL | `localhost:5052` | 5052 |
| Redis | `localhost:5056` | 5056 |
| NATS | `localhost:5057` | 5057 |
| NATS Monitor | `http://localhost:5058` | 5058 |
| MailHog (Web) | `http://localhost:5059` | 5059 |
| MailHog (SMTP) | `localhost:5060` | 5060 |

## First Login

Open the Admin UI at [http://localhost:5054](http://localhost:5054) and sign in with the default credentials:

- **Email:** `admin@cpi-auth.local`
- **Password:** `admin123!`

::: warning
Change the default admin password immediately in production environments.
:::

## Verify the Backend

Check the health endpoint:

```bash
curl http://localhost:5050/health
```

Query the OIDC discovery document:

```bash
curl http://localhost:5050/.well-known/openid-configuration
```

## Creating Your First Application

### Via the Admin UI

1. Log in to the Admin UI at `http://localhost:5054`
2. Navigate to **Applications** in the sidebar
3. Click **Create Application**
4. Fill in the details:
   - **Name:** My SPA
   - **Type:** Single Page Application (SPA)
   - **Redirect URIs:** `http://localhost:3000/callback`
   - **Allowed Origins:** `http://localhost:3000`
   - **Grant Types:** Authorization Code
5. Click **Save**
6. Copy the **Client ID** from the application detail page

### Via the API

```bash
# Authenticate and get an access token
TOKEN=$(curl -s -X POST http://localhost:5050/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "password",
    "email": "admin@cpi-auth.local",
    "password": "admin123!",
    "client_id": "admin-ui"
  }' | jq -r '.access_token')

# Create an application
curl -X POST http://localhost:5050/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My SPA",
    "type": "spa",
    "redirect_uris": ["http://localhost:3000/callback"],
    "allowed_origins": ["http://localhost:3000"],
    "grant_types": ["authorization_code"]
  }'
```

The response includes a `client_id` you will use in your frontend application.

## Connecting a Frontend

Use the Authorization Code flow with PKCE from your SPA:

```javascript
// Redirect to CPI Auth login
const params = new URLSearchParams({
  client_id: 'YOUR_CLIENT_ID',
  redirect_uri: 'http://localhost:3000/callback',
  response_type: 'code',
  scope: 'openid profile email',
  code_challenge: codeChallenge,
  code_challenge_method: 'S256',
  state: randomState
});

window.location.href = `http://localhost:5053/login?${params}`;
```

After login, exchange the authorization code for tokens:

```javascript
const response = await fetch('http://localhost:5050/oauth/token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    grant_type: 'authorization_code',
    code: authorizationCode,
    redirect_uri: 'http://localhost:3000/callback',
    client_id: 'YOUR_CLIENT_ID',
    code_verifier: codeVerifier
  })
});

const { access_token, id_token, refresh_token } = await response.json();
```

## Check Emails

All emails sent during development (verification, password reset, etc.) are captured by MailHog. Open [http://localhost:5059](http://localhost:5059) to view them.

## Running Tests

Run the full test suite:

```bash
# Go backend tests
go test ./...

# Run all tests including browser E2E
./run-all-tests.sh
```

## Stopping Services

```bash
docker compose down
```

To remove all data volumes as well:

```bash
docker compose down -v
```

## Next Steps

- [Architecture](./architecture) -- Understand how the system is structured
- [Configuration](./configuration) -- Customize settings for your environment
- [Applications](./applications) -- Deep dive into application types and settings
- [Auth Flows](./auth-flows) -- Learn about all supported authentication flows
