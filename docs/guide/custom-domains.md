# Custom Domains

Custom domains allow tenants to serve authentication pages from their own domain (e.g., `auth.acme.com`) instead of the default CPI Auth URLs. This provides a seamless, white-label experience for end users.

## How It Works

1. The tenant registers a custom domain with CPI Auth
2. CPI Auth generates a DNS TXT verification token
3. The tenant adds the TXT record to their DNS
4. CPI Auth verifies the DNS record
5. The tenant configures a CNAME pointing to the CPI Auth instance
6. All auth pages for that tenant are served on the custom domain

## Domain Verification Model

```json
{
  "id": "d1v2r3f4-a5b6-7890-cdef-012345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "domain": "auth.acme.com",
  "verification_token": "cpi-auth-verify=abc123def456",
  "verification_method": "dns_txt",
  "is_verified": false,
  "verified_at": null,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

## Verification Flow

### Step 1: Initiate Verification

```bash
curl -X POST http://localhost:5050/api/v1/domains/verification \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "auth.acme.com"
  }'
```

Response:

```json
{
  "id": "d1v2r3f4-a5b6-7890-cdef-012345678901",
  "domain": "auth.acme.com",
  "verification_token": "cpi-auth-verify=abc123def456",
  "verification_method": "dns_txt",
  "is_verified": false
}
```

### Step 2: Add DNS Records

Add two DNS records for your domain:

**TXT Record** (for verification):

```
_cpi-auth.auth.acme.com  TXT  "cpi-auth-verify=abc123def456"
```

**CNAME Record** (for routing):

```
auth.acme.com  CNAME  login.your-cpi-auth-instance.com
```

Allow time for DNS propagation (typically 5 minutes to 48 hours depending on your DNS provider).

### Step 3: Check Verification

```bash
curl -X POST http://localhost:5050/api/v1/domains/verification/{verification_id}/check \
  -H "Authorization: Bearer $TOKEN"
```

Response when successful:

```json
{
  "id": "d1v2r3f4-a5b6-7890-cdef-012345678901",
  "domain": "auth.acme.com",
  "is_verified": true,
  "verified_at": "2025-01-15T12:00:00Z"
}
```

Response when DNS record not found:

```json
{
  "id": "d1v2r3f4-a5b6-7890-cdef-012345678901",
  "domain": "auth.acme.com",
  "is_verified": false,
  "error": "TXT record not found. Ensure _cpi-auth.auth.acme.com has the correct TXT value."
}
```

### Step 4: Domain Is Active

Once verified, CPI Auth automatically resolves the tenant when requests arrive at `auth.acme.com`. The Login UI serves authentication pages on this domain.

## Managing Domain Verification

### Get Current Verification Status

```bash
curl http://localhost:5050/api/v1/domains/verification \
  -H "Authorization: Bearer $TOKEN"
```

### Remove a Domain Verification

```bash
curl -X DELETE http://localhost:5050/api/v1/domains/verification/{verification_id} \
  -H "Authorization: Bearer $TOKEN"
```

Removing a verification does not delete DNS records -- you must remove those from your DNS provider separately.

## TLS / SSL

CPI Auth does not handle TLS termination for custom domains directly. You need to set up TLS at your infrastructure layer:

### Option 1: Reverse Proxy (Recommended)

Use nginx, Caddy, or a cloud load balancer with automatic certificate provisioning:

```nginx
# nginx example
server {
    listen 443 ssl;
    server_name auth.acme.com;

    ssl_certificate /etc/letsencrypt/live/auth.acme.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/auth.acme.com/privkey.pem;

    location / {
        proxy_pass http://cpi-auth-login-ui:5053;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Option 2: Caddy (Auto TLS)

```
auth.acme.com {
    reverse_proxy cpi-auth-login-ui:5053
}
```

Caddy automatically provisions and renews Let's Encrypt certificates.

### Option 3: Cloud Load Balancer

AWS ALB, GCP Load Balancer, or Cloudflare can terminate TLS and route traffic to your CPI Auth instance.

## WebAuthn Considerations

When using custom domains with WebAuthn, ensure the `webauthn.rp_id` and `webauthn.rp_origins` in your configuration include the custom domain:

```yaml
webauthn:
  rp_id: "acme.com"
  rp_origins:
    - "https://auth.acme.com"
    - "https://account.acme.com"
```

The `rp_id` should be set to the registrable domain (e.g., `acme.com`) so it works for both `auth.acme.com` and `account.acme.com`.

## CORS Configuration

Add custom domains to the CORS allowed origins:

```yaml
security:
  cors_allowed_origins:
    - "https://auth.acme.com"
    - "https://app.acme.com"
```

Or via environment variable:

```bash
AF_SECURITY_CORS_ALLOWED_ORIGINS="https://auth.acme.com,https://app.acme.com"
```

## Social Login Redirect URIs

When using custom domains with social login providers, update the callback URLs in each provider's developer console to include the custom domain:

```
https://auth.acme.com/auth/callback/google
https://auth.acme.com/auth/callback/github
```

## Next Steps

- [Tenants](./tenants) -- Custom domains are a tenant-level feature
- [Configuration](./configuration) -- CORS and WebAuthn configuration
- [Social Login](./social-login) -- Update redirect URIs for custom domains
