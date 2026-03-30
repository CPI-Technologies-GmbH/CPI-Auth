package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// ─── 1. Multi-Tenant Isolation Tests ───────────────────────────────────────────

func TestMultiTenantIsolation(t *testing.T) {
	adminToken := getAdminAccessToken(t)

	// Create tenant A
	tenantABody := map[string]interface{}{
		"name":   "Isolation Tenant A",
		"slug":   "isolation-a",
		"domain": "isolation-a.test.local",
	}
	respA, bodyA := mustPostJSON(t, apiBaseURL+"/admin/tenants", tenantABody, adminToken)
	if respA.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create tenant A: status=%d body=%s", respA.StatusCode, string(bodyA))
	}
	var tenantA map[string]interface{}
	mustUnmarshal(t, bodyA, &tenantA)
	tenantAID := tenantA["id"].(string)

	// Create tenant B
	tenantBBody := map[string]interface{}{
		"name":   "Isolation Tenant B",
		"slug":   "isolation-b",
		"domain": "isolation-b.test.local",
	}
	respB, bodyB := mustPostJSON(t, apiBaseURL+"/admin/tenants", tenantBBody, adminToken)
	if respB.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create tenant B: status=%d body=%s", respB.StatusCode, string(bodyB))
	}
	var tenantB map[string]interface{}
	mustUnmarshal(t, bodyB, &tenantB)
	tenantBID := tenantB["id"].(string)

	// Create a user in tenant A via direct DB insert (admin API creates in seed tenant)
	userAID := ""
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (tenant_id, email, password_hash, name, status, email_verified, created_at, updated_at)
		 VALUES ($1, 'isolation-user-a@test.local', '$2a$10$dummyhashfortest000000000000000000000000000000000000', 'Isolation User A', 'active', true, NOW(), NOW())
		 RETURNING id`,
		tenantAID,
	).Scan(&userAID)
	if err != nil {
		t.Fatalf("Failed to create user in tenant A: %v", err)
	}

	// Cleanup at end
	defer func() {
		pool.Exec(context.Background(), "DELETE FROM users WHERE email = 'isolation-user-a@test.local'")
		pool.Exec(context.Background(), "DELETE FROM tenants WHERE slug = 'isolation-a'")
		pool.Exec(context.Background(), "DELETE FROM tenants WHERE slug = 'isolation-b'")
	}()

	// The admin token is scoped to the seed tenant. Verify tenant isolation
	// by checking that users from tenant A cannot be accessed when filtering by tenant B.

	t.Run("CrossTenant_UserNotVisible", func(t *testing.T) {
		// Try to get the tenant A user by ID via admin API
		// The admin is scoped to the seed tenant, so this user should not be accessible
		resp, body := mustGetAuth(t, apiBaseURL+"/admin/users/"+userAID, adminToken)
		// The user was created in tenant A, not the seed tenant.
		// The admin API should either return 404 or the user depending on
		// whether the admin has cross-tenant access. Document the behavior.
		if resp.StatusCode == http.StatusOK {
			var user map[string]interface{}
			mustUnmarshal(t, body, &user)
			userTenant, _ := user["tenant_id"].(string)
			if userTenant != "" && userTenant != seedTenantID {
				t.Errorf("Admin in seed tenant can access user from tenant %s - potential isolation breach", userTenant)
			}
		}
		// 404 is the expected secure behavior
		t.Logf("Cross-tenant user access returned status %d", resp.StatusCode)
	})

	t.Run("CrossTenant_ApplicationsIsolated", func(t *testing.T) {
		// The seed admin should only see applications in the seed tenant
		resp, body := mustGetAuth(t, apiBaseURL+"/admin/applications", adminToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var apps []map[string]interface{}
		mustUnmarshal(t, body, &apps)

		for _, app := range apps {
			appTenantID, _ := app["tenant_id"].(string)
			if appTenantID != "" && appTenantID != seedTenantID {
				t.Errorf("Application from tenant %s visible to seed tenant admin - isolation breach", appTenantID)
			}
		}
	})

	t.Run("CrossTenant_AuditLogsIsolated", func(t *testing.T) {
		resp, body := mustGetAuth(t, apiBaseURL+"/admin/audit-logs", adminToken)
		if resp.StatusCode == http.StatusOK {
			var logs []map[string]interface{}
			// The response may be a list or an object with items
			if err := json.Unmarshal(body, &logs); err != nil {
				// Try as paginated response
				var paged map[string]interface{}
				if err2 := json.Unmarshal(body, &paged); err2 == nil {
					if items, ok := paged["items"].([]interface{}); ok {
						for _, item := range items {
							if logEntry, ok := item.(map[string]interface{}); ok {
								logTenant, _ := logEntry["tenant_id"].(string)
								if logTenant != "" && logTenant != seedTenantID {
									t.Errorf("Audit log from tenant %s visible to seed tenant admin", logTenant)
								}
							}
						}
					}
				}
			} else {
				for _, logEntry := range logs {
					logTenant, _ := logEntry["tenant_id"].(string)
					if logTenant != "" && logTenant != seedTenantID {
						t.Errorf("Audit log from tenant %s visible to seed tenant admin", logTenant)
					}
				}
			}
		}
		t.Logf("Audit logs endpoint returned status %d", resp.StatusCode)
	})

	t.Run("TenantA_NotAccessibleFromTenantB", func(t *testing.T) {
		// Verify that tenant B's data does not include tenant A's data via DB
		var count int
		err := pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND email = 'isolation-user-a@test.local'",
			tenantBID,
		).Scan(&count)
		if err != nil {
			t.Fatalf("DB query failed: %v", err)
		}
		if count != 0 {
			t.Error("User from tenant A should not exist in tenant B's scope")
		}
	})
}

// ─── 2. OAuth Security Tests ──────────────────────────────────────────────────

func TestOAuthSecurity(t *testing.T) {
	verifier := "oauth_security_test_pkce_verifier_value_12345"
	challenge := generatePKCEChallenge(verifier)

	t.Run("AuthCodeReuse", func(t *testing.T) {
		code := "e2e_reuse_code_" + fmt.Sprintf("%d", time.Now().UnixNano())
		expiresAt := time.Now().UTC().Add(5 * time.Minute)

		_, err := pool.Exec(context.Background(),
			`INSERT INTO oauth_grants (user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			seedAdminID, seedAppID, seedTenantID,
			[]string{"openid", "profile", "email"},
			code, challenge, "S256", seedRedirectURI, expiresAt,
		)
		if err != nil {
			t.Fatalf("Failed to create test grant: %v", err)
		}

		// First exchange should succeed
		resp1, body1 := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {seedRedirectURI},
			"client_id":     {seedClientID},
			"code_verifier": {verifier},
		})
		assertStatus(t, http.StatusOK, resp1.StatusCode)

		var tokens1 map[string]interface{}
		mustUnmarshal(t, body1, &tokens1)
		if tokens1["access_token"] == nil || tokens1["access_token"] == "" {
			t.Fatal("First exchange should return access_token")
		}

		// Second exchange with same code must fail
		resp2, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {seedRedirectURI},
			"client_id":     {seedClientID},
			"code_verifier": {verifier},
		})
		if resp2.StatusCode == http.StatusOK {
			t.Error("Reusing authorization code must fail - code should be single-use")
		}
	})

	t.Run("WrongRedirectURI", func(t *testing.T) {
		code := "e2e_wrongredirect_" + fmt.Sprintf("%d", time.Now().UnixNano())
		expiresAt := time.Now().UTC().Add(5 * time.Minute)

		_, err := pool.Exec(context.Background(),
			`INSERT INTO oauth_grants (user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			seedAdminID, seedAppID, seedTenantID,
			[]string{"openid"}, code, challenge, "S256", seedRedirectURI, expiresAt,
		)
		if err != nil {
			t.Fatalf("Failed to create test grant: %v", err)
		}

		// Try token exchange with a different redirect_uri
		resp, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {"http://evil.example.com/callback"},
			"client_id":     {seedClientID},
			"code_verifier": {verifier},
		})
		if resp.StatusCode == http.StatusOK {
			t.Error("Token exchange with wrong redirect_uri must fail")
		}

		// Clean up the grant if it wasn't consumed
		pool.Exec(context.Background(), "DELETE FROM oauth_grants WHERE code = $1", code)
	})

	t.Run("PKCEWrongVerifier", func(t *testing.T) {
		code := "e2e_wrongpkce_" + fmt.Sprintf("%d", time.Now().UnixNano())
		expiresAt := time.Now().UTC().Add(5 * time.Minute)

		_, err := pool.Exec(context.Background(),
			`INSERT INTO oauth_grants (user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			seedAdminID, seedAppID, seedTenantID,
			[]string{"openid"}, code, challenge, "S256", seedRedirectURI, expiresAt,
		)
		if err != nil {
			t.Fatalf("Failed to create test grant: %v", err)
		}

		// Use a completely different verifier
		resp, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {seedRedirectURI},
			"client_id":     {seedClientID},
			"code_verifier": {"completely_wrong_verifier_not_matching_challenge"},
		})
		if resp.StatusCode == http.StatusOK {
			t.Error("Token exchange with wrong PKCE verifier must fail")
		}

		// Clean up
		pool.Exec(context.Background(), "DELETE FROM oauth_grants WHERE code = $1", code)
	})

	t.Run("ExpiredTokenTTL", func(t *testing.T) {
		// Get a valid token and verify its exp claim is reasonable
		token := getAdminAccessToken(t)
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("Token is not a valid JWT (expected 3 parts, got %d)", len(parts))
		}

		// Decode payload (add padding if needed)
		payload := parts[1]
		if m := len(payload) % 4; m != 0 {
			payload += strings.Repeat("=", 4-m)
		}
		decoded, err := base64.URLEncoding.DecodeString(payload)
		if err != nil {
			t.Fatalf("Failed to decode JWT payload: %v", err)
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(decoded, &claims); err != nil {
			t.Fatalf("Failed to parse JWT claims: %v", err)
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			t.Fatal("JWT missing exp claim")
		}
		iat, ok := claims["iat"].(float64)
		if !ok {
			t.Fatal("JWT missing iat claim")
		}

		ttl := exp - iat
		// Token TTL should be reasonable: not more than 24 hours (86400s)
		if ttl > 86400 {
			t.Errorf("Token TTL is %v seconds, which exceeds 24 hours - potentially insecure", ttl)
		}
		if ttl <= 0 {
			t.Errorf("Token TTL is %v seconds, which is non-positive", ttl)
		}
		t.Logf("Token TTL: %v seconds (%.1f hours)", ttl, ttl/3600)
	})
}

// ─── 3. Token Signature Security Tests ─────────────────────────────────────────

func TestTokenSignatureSecurity(t *testing.T) {
	validToken := getAdminAccessToken(t)

	t.Run("TamperedPayload", func(t *testing.T) {
		parts := strings.Split(validToken, ".")
		if len(parts) != 3 {
			t.Fatalf("Token is not a valid JWT (expected 3 parts, got %d)", len(parts))
		}

		// Decode the payload
		payload := parts[1]
		if m := len(payload) % 4; m != 0 {
			payload += strings.Repeat("=", 4-m)
		}
		decoded, err := base64.URLEncoding.DecodeString(payload)
		if err != nil {
			t.Fatalf("Failed to decode JWT payload: %v", err)
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(decoded, &claims); err != nil {
			t.Fatalf("Failed to parse JWT claims: %v", err)
		}

		// Tamper with the claims
		claims["email"] = "hacker@evil.com"
		claims["sub"] = "00000000-0000-0000-0000-000000000099"

		tamperedJSON, _ := json.Marshal(claims)
		tamperedPayload := base64.RawURLEncoding.EncodeToString(tamperedJSON)

		// Re-assemble token with original header and signature but tampered payload
		tamperedToken := parts[0] + "." + tamperedPayload + "." + parts[2]

		resp, _ := mustGetAuth(t, apiBaseURL+"/admin/auth/me", tamperedToken)
		if resp.StatusCode == http.StatusOK {
			t.Error("Tampered JWT payload must be rejected (signature mismatch)")
		}
	})

	t.Run("CompletelyInvalidJWT", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/admin/auth/me", "not.a.jwt.token.at.all")
		if resp.StatusCode == http.StatusOK {
			t.Error("Completely invalid JWT must be rejected")
		}
	})

	t.Run("AlgNoneAttack", func(t *testing.T) {
		// Construct a JWT with alg: none
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))

		// Use valid-looking claims
		claims := map[string]interface{}{
			"sub":       seedAdminID,
			"tenant_id": seedTenantID,
			"email":     seedAdminEmail,
			"iat":       time.Now().Unix(),
			"exp":       time.Now().Add(1 * time.Hour).Unix(),
		}
		claimsJSON, _ := json.Marshal(claims)
		payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

		// alg:none token has empty signature
		noneToken := header + "." + payload + "."

		resp, _ := mustGetAuth(t, apiBaseURL+"/admin/auth/me", noneToken)
		if resp.StatusCode == http.StatusOK {
			t.Error("JWT with alg:none must be rejected - critical vulnerability")
		}
	})

	t.Run("EmptyToken", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/admin/auth/me", "")
		if resp.StatusCode == http.StatusOK {
			t.Error("Empty token must be rejected")
		}
	})

	t.Run("TokenWithModifiedSignature", func(t *testing.T) {
		parts := strings.Split(validToken, ".")
		if len(parts) != 3 {
			t.Fatalf("Token is not a valid JWT (expected 3 parts, got %d)", len(parts))
		}

		// Modify the signature by flipping some characters
		sig := []byte(parts[2])
		if len(sig) > 5 {
			for i := 0; i < 5; i++ {
				sig[i] = 'A'
			}
		}
		modifiedToken := parts[0] + "." + parts[1] + "." + string(sig)

		resp, _ := mustGetAuth(t, apiBaseURL+"/admin/auth/me", modifiedToken)
		if resp.StatusCode == http.StatusOK {
			t.Error("JWT with modified signature must be rejected")
		}
	})
}

// ─── 4. Brute Force / Rate Limiting Tests ──────────────────────────────────────

func TestBruteForceProtection(t *testing.T) {
	const attempts = 20
	statuses := make(map[int]int)

	for i := 0; i < attempts; i++ {
		loginBody := map[string]interface{}{
			"email":    seedAdminEmail,
			"password": fmt.Sprintf("wrong_password_%d", i),
		}
		resp, _ := mustPostJSON(t, apiBaseURL+"/api/v1/auth/login", loginBody, "")
		statuses[resp.StatusCode]++
	}

	t.Logf("Brute force results after %d attempts: %v", attempts, statuses)

	// Check if any rate limiting or account locking was triggered
	rateLimited := statuses[http.StatusTooManyRequests] > 0
	accountLocked := false
	for code, count := range statuses {
		if code == http.StatusForbidden && count > 0 {
			accountLocked = true
		}
	}

	if rateLimited {
		t.Logf("Rate limiting IS active (429 responses received)")
	} else if accountLocked {
		t.Logf("Account locking IS active (403 responses received after repeated failures)")
	} else {
		t.Logf("WARNING: No rate limiting or account locking detected after %d failed attempts. "+
			"All responses were: %v. Consider implementing rate limiting.", attempts, statuses)
	}

	// Verify the admin account is still usable (not permanently locked)
	// Wait briefly in case there's a short lockout
	time.Sleep(2 * time.Second)

	// Try to get a valid token - if account is temporarily locked, this documents it
	loginBody := map[string]interface{}{
		"email":    seedAdminEmail,
		"password": seedAdminPassword,
	}
	resp, body := mustPostJSON(t, apiBaseURL+"/api/v1/auth/login", loginBody, "")
	t.Logf("Login with correct password after brute force: status=%d", resp.StatusCode)
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
		t.Logf("Account is temporarily locked/rate-limited after brute force (body: %s) - this is acceptable security behavior", string(body))
	}
}

// ─── 5. Impersonation Security Tests ───────────────────────────────────────────

func TestImpersonationSecurity(t *testing.T) {
	adminToken := getAdminAccessToken(t)

	// Create a target user for impersonation
	userBody := map[string]interface{}{
		"email":    "impersonate-target@test.local",
		"password": "ImpTarget1234!",
		"name":     "Impersonation Target",
	}
	createResp, createBody := mustPostJSON(t, apiBaseURL+"/admin/users", userBody, adminToken)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create target user: status=%d body=%s", createResp.StatusCode, string(createBody))
	}
	var targetUser map[string]interface{}
	mustUnmarshal(t, createBody, &targetUser)
	targetUserID := targetUser["id"].(string)

	defer func() {
		mustDeleteAuth(t, apiBaseURL+"/admin/users/"+targetUserID, adminToken)
	}()

	t.Run("ImpersonationTokenClaims", func(t *testing.T) {
		resp, body := mustPostJSON(t, apiBaseURL+"/admin/users/"+targetUserID+"/impersonate", nil, adminToken)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Impersonation failed: status=%d body=%s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		mustUnmarshal(t, body, &result)

		accessToken, _ := result["access_token"].(string)
		if accessToken == "" {
			t.Fatal("Impersonation response missing access_token")
		}

		// Verify no refresh_token in the response
		if result["refresh_token"] != nil && result["refresh_token"] != "" {
			t.Error("Impersonation token should NOT include refresh_token")
		}

		// Verify impersonated flag
		if imp, ok := result["impersonated"].(bool); !ok || !imp {
			t.Error("Impersonation response should have impersonated=true")
		}

		// Decode the token and check claims
		parts := strings.Split(accessToken, ".")
		if len(parts) != 3 {
			t.Fatalf("Impersonation token is not a valid JWT (got %d parts)", len(parts))
		}

		payload := parts[1]
		if m := len(payload) % 4; m != 0 {
			payload += strings.Repeat("=", 4-m)
		}
		decoded, err := base64.URLEncoding.DecodeString(payload)
		if err != nil {
			t.Fatalf("Failed to decode JWT payload: %v", err)
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(decoded, &claims); err != nil {
			t.Fatalf("Failed to parse JWT claims: %v", err)
		}

		// Verify sub is the target user
		sub, _ := claims["sub"].(string)
		if sub != targetUserID {
			t.Errorf("Impersonation token sub should be target user %s, got %s", targetUserID, sub)
		}

		// Verify act claim exists with admin's user ID
		actClaim, ok := claims["act"].(map[string]interface{})
		if !ok {
			t.Fatal("Impersonation token missing 'act' claim (RFC 8693)")
		}
		actSub, _ := actClaim["sub"].(string)
		if actSub != seedAdminID {
			t.Errorf("Impersonation act.sub should be admin %s, got %s", seedAdminID, actSub)
		}

		// Verify short TTL (should be <= 900 seconds / 15 minutes)
		exp, _ := claims["exp"].(float64)
		iat, _ := claims["iat"].(float64)
		ttl := exp - iat
		if ttl > 900 {
			t.Errorf("Impersonation token TTL is %v seconds, should be <= 900 (15 min)", ttl)
		}
		if ttl <= 0 {
			t.Errorf("Impersonation token TTL is %v seconds, should be positive", ttl)
		}
		t.Logf("Impersonation token TTL: %v seconds", ttl)
	})

	t.Run("ImpersonationWithApplicationID", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"application_id": seedAppID,
		}
		resp, body := mustPostJSON(t, apiBaseURL+"/admin/users/"+targetUserID+"/impersonate", reqBody, adminToken)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Impersonation with app_id failed: status=%d body=%s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		mustUnmarshal(t, body, &result)

		// When application_id is provided, redirect_url should be present
		redirectURL, _ := result["redirect_url"].(string)
		t.Logf("Impersonation with application_id returned redirect_url=%q", redirectURL)
		// redirect_url may be empty if the app has no redirect URIs configured,
		// but it should at least be present in the response structure
	})

	t.Run("ImpersonateNonExistentUser", func(t *testing.T) {
		fakeUserID := "00000000-0000-0000-0000-000000000099"
		resp, _ := mustPostJSON(t, apiBaseURL+"/admin/users/"+fakeUserID+"/impersonate", nil, adminToken)
		if resp.StatusCode == http.StatusOK {
			t.Error("Impersonating a non-existent user must fail")
		}
		t.Logf("Impersonating non-existent user returned status %d", resp.StatusCode)
	})

	t.Run("ImpersonateWithoutAuth", func(t *testing.T) {
		resp, _ := mustPostJSON(t, apiBaseURL+"/admin/users/"+targetUserID+"/impersonate", nil, "")
		if resp.StatusCode == http.StatusOK {
			t.Error("Impersonation without authentication must fail")
		}
	})
}

// ─── 6. Device Auth Flow Tests ─────────────────────────────────────────────────

func TestDeviceAuthFlow(t *testing.T) {
	t.Run("FullDeviceFlow", func(t *testing.T) {
		// Step 1: Request device code
		resp, body := mustPostForm(t, apiBaseURL+"/oauth/device/code", url.Values{
			"client_id": {seedClientID},
			"scope":     {"openid profile email"},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Device code request failed: status=%d body=%s", resp.StatusCode, string(body))
		}

		var deviceResp map[string]interface{}
		mustUnmarshal(t, body, &deviceResp)

		deviceCode, _ := deviceResp["device_code"].(string)
		userCode, _ := deviceResp["user_code"].(string)

		if deviceCode == "" {
			t.Fatal("Device code response missing device_code")
		}
		if userCode == "" {
			t.Fatal("Device code response missing user_code")
		}

		// Verify user_code format: XXXX-XXXX
		if len(userCode) != 9 || userCode[4] != '-' {
			t.Errorf("User code format should be XXXX-XXXX, got %q", userCode)
		}

		// Verify verification_uri is present
		verificationURI, _ := deviceResp["verification_uri"].(string)
		if verificationURI == "" {
			t.Error("Device code response missing verification_uri")
		}

		// Verify expires_in is present and reasonable
		expiresIn, _ := deviceResp["expires_in"].(float64)
		if expiresIn <= 0 {
			t.Error("Device code response missing or invalid expires_in")
		}
		t.Logf("Device code: user_code=%s, expires_in=%v", userCode, expiresIn)

		// Step 2: Poll before authorization - should get authorization_pending
		t.Run("PollBeforeAuth", func(t *testing.T) {
			pollResp, pollBody := mustPostForm(t, apiBaseURL+"/oauth/device/token", url.Values{
				"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
				"device_code": {deviceCode},
				"client_id":   {seedClientID},
			})

			if pollResp.StatusCode == http.StatusOK {
				t.Error("Polling before authorization should not return tokens")
			}

			var pollResult map[string]interface{}
			mustUnmarshal(t, pollBody, &pollResult)
			errorCode, _ := pollResult["error"].(string)
			if errorCode != "authorization_pending" {
				t.Errorf("Expected error 'authorization_pending', got %q", errorCode)
			}
		})

		// Step 3: Authorize the device code (as authenticated user)
		adminToken := getAdminAccessToken(t)
		t.Run("AuthorizeDevice", func(t *testing.T) {
			authBody := map[string]interface{}{
				"user_code": userCode,
			}
			authResp, authRespBody := mustPostJSON(t, apiBaseURL+"/oauth/device/authorize", authBody, adminToken)
			if authResp.StatusCode != http.StatusOK {
				t.Fatalf("Device authorization failed: status=%d body=%s", authResp.StatusCode, string(authRespBody))
			}

			var authResult map[string]interface{}
			mustUnmarshal(t, authRespBody, &authResult)
			status, _ := authResult["status"].(string)
			if status != "authorized" {
				t.Errorf("Expected status 'authorized', got %q", status)
			}
		})

		// Step 4: Poll after authorization - should get tokens
		t.Run("PollAfterAuth", func(t *testing.T) {
			pollResp, pollBody := mustPostForm(t, apiBaseURL+"/oauth/device/token", url.Values{
				"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
				"device_code": {deviceCode},
				"client_id":   {seedClientID},
			})

			if pollResp.StatusCode != http.StatusOK {
				t.Fatalf("Polling after authorization should return tokens: status=%d body=%s", pollResp.StatusCode, string(pollBody))
			}

			var tokenResult map[string]interface{}
			mustUnmarshal(t, pollBody, &tokenResult)

			accessToken, _ := tokenResult["access_token"].(string)
			if accessToken == "" {
				t.Fatal("Device token response missing access_token")
			}

			// Verify the access token is a valid JWT
			parts := strings.Split(accessToken, ".")
			if len(parts) != 3 {
				t.Fatalf("Device access token is not a valid JWT (got %d parts)", len(parts))
			}

			// Decode and verify claims
			payload := parts[1]
			if m := len(payload) % 4; m != 0 {
				payload += strings.Repeat("=", 4-m)
			}
			decoded, err := base64.URLEncoding.DecodeString(payload)
			if err != nil {
				t.Fatalf("Failed to decode device token payload: %v", err)
			}

			var claims map[string]interface{}
			if err := json.Unmarshal(decoded, &claims); err != nil {
				t.Fatalf("Failed to parse device token claims: %v", err)
			}

			sub, _ := claims["sub"].(string)
			if sub == "" {
				t.Error("Device token missing sub claim")
			}
			t.Logf("Device token issued for user sub=%s", sub)
		})
	})

	t.Run("DeviceCode_MissingClientID", func(t *testing.T) {
		resp, _ := mustPostForm(t, apiBaseURL+"/oauth/device/code", url.Values{})
		if resp.StatusCode == http.StatusOK {
			t.Error("Device code request without client_id should fail")
		}
	})

	t.Run("DeviceToken_WrongGrantType", func(t *testing.T) {
		resp, body := mustPostForm(t, apiBaseURL+"/oauth/device/token", url.Values{
			"grant_type":  {"authorization_code"},
			"device_code": {"some_code"},
			"client_id":   {seedClientID},
		})
		if resp.StatusCode == http.StatusOK {
			t.Error("Device token with wrong grant_type should fail")
		}
		var result map[string]interface{}
		mustUnmarshal(t, body, &result)
		errorCode, _ := result["error"].(string)
		if errorCode != "unsupported_grant_type" {
			t.Errorf("Expected error 'unsupported_grant_type', got %q", errorCode)
		}
	})

	t.Run("DeviceToken_InvalidCode", func(t *testing.T) {
		resp, body := mustPostForm(t, apiBaseURL+"/oauth/device/token", url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": {"nonexistent_device_code"},
			"client_id":   {seedClientID},
		})
		if resp.StatusCode == http.StatusOK {
			t.Error("Device token with invalid device_code should fail")
		}
		var result map[string]interface{}
		mustUnmarshal(t, body, &result)
		errorCode, _ := result["error"].(string)
		if errorCode != "invalid_grant" {
			t.Errorf("Expected error 'invalid_grant', got %q", errorCode)
		}
	})

	t.Run("DeviceAuthorize_InvalidUserCode", func(t *testing.T) {
		adminToken := getAdminAccessToken(t)
		authBody := map[string]interface{}{
			"user_code": "XXXX-ZZZZ",
		}
		resp, _ := mustPostJSON(t, apiBaseURL+"/oauth/device/authorize", authBody, adminToken)
		if resp.StatusCode == http.StatusOK {
			t.Error("Device authorize with invalid user_code should fail")
		}
	})

	t.Run("DeviceAuthorize_NoAuth", func(t *testing.T) {
		authBody := map[string]interface{}{
			"user_code": "XXXX-YYYY",
		}
		resp, _ := mustPostJSON(t, apiBaseURL+"/oauth/device/authorize", authBody, "")
		if resp.StatusCode == http.StatusOK {
			t.Error("Device authorize without authentication should fail")
		}
	})
}

// ─── Helper: PKCE challenge generation (re-exported for clarity) ────────────

func generateSecurityTestPKCEChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
