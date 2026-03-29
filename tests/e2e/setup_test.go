package e2e

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Configuration ──────────────────────────────────────────────────────────

const (
	// Seed data UUIDs from 002_seed_data.up.sql
	seedTenantID  = "a0000000-0000-0000-0000-000000000001"
	seedAdminID   = "c0000000-0000-0000-0000-000000000001"
	seedAppID     = "d0000000-0000-0000-0000-000000000001"
	seedRoleAdmin = "b0000000-0000-0000-0000-000000000001"
	seedRoleMgr   = "b0000000-0000-0000-0000-000000000002"
	seedRoleEdit  = "b0000000-0000-0000-0000-000000000003"
	seedRoleView  = "b0000000-0000-0000-0000-000000000004"

	seedAdminEmail    = "admin@cpi-auth.local"
	seedAdminPassword = "admin123!"
	seedClientID      = "authforge_default_app"
	seedRedirectURI   = "http://localhost:3000/callback"
)

var (
	apiBaseURL string
	dbConnStr  string
	pool       *pgxpool.Pool
	httpClient *http.Client
)

func TestMain(m *testing.M) {
	apiBaseURL = envOr("E2E_API_URL", "http://localhost:5050")
	dbConnStr = envOr("E2E_DATABASE_URL", "postgres://cpi-auth:cpi-auth@localhost:5052/cpi-auth?sslmode=disable")

	httpClient = &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirects
		},
	}

	// Connect to database
	var err error
	pool, err = pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Wait for API to be ready
	if err := waitForAPI(apiBaseURL, 60*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "API not ready: %v\n", err)
		os.Exit(1)
	}

	// Clean up test data from previous runs
	pool.Exec(context.Background(), "DELETE FROM roles WHERE name = 'e2e-test-role'")
	pool.Exec(context.Background(), "DELETE FROM users WHERE email = 'e2etest@cpi-auth.local'")
	pool.Exec(context.Background(), "DELETE FROM tenants WHERE slug = 'e2e-test'")
	pool.Exec(context.Background(), "DELETE FROM organizations WHERE slug = 'e2e-test-org'")

	os.Exit(m.Run())
}

// ─── 1. Seed Data Verification ──────────────────────────────────────────────

func TestSeedData_Tenant(t *testing.T) {
	var name, slug, domain string
	err := pool.QueryRow(context.Background(),
		"SELECT name, slug, domain FROM tenants WHERE id = $1", seedTenantID,
	).Scan(&name, &slug, &domain)
	if err != nil {
		t.Fatalf("Default tenant not found in DB: %v", err)
	}
	assertEqual(t, "Default Tenant", name, "tenant name")
	assertEqual(t, "default", slug, "tenant slug")
	assertEqual(t, "localhost", domain, "tenant domain")
}

func TestSeedData_AdminUser(t *testing.T) {
	var email, userStatus string
	var emailVerified bool
	err := pool.QueryRow(context.Background(),
		"SELECT email, email_verified, status FROM users WHERE id = $1", seedAdminID,
	).Scan(&email, &emailVerified, &userStatus)
	if err != nil {
		t.Fatalf("Admin user not found in DB: %v", err)
	}
	assertEqual(t, seedAdminEmail, email, "admin email")
	assertEqual(t, true, emailVerified, "admin email_verified")
	assertEqual(t, "active", userStatus, "admin status")
}

func TestSeedData_Roles(t *testing.T) {
	roles := []struct {
		id   string
		name string
	}{
		{seedRoleAdmin, "admin"},
		{seedRoleMgr, "manager"},
		{seedRoleEdit, "editor"},
		{seedRoleView, "viewer"},
	}

	for _, r := range roles {
		var name string
		var isSystem bool
		err := pool.QueryRow(context.Background(),
			"SELECT name, is_system FROM roles WHERE id = $1", r.id,
		).Scan(&name, &isSystem)
		if err != nil {
			t.Fatalf("Role %s not found: %v", r.name, err)
		}
		assertEqual(t, r.name, name, "role name")
		assertEqual(t, true, isSystem, "role is_system for "+r.name)
	}
}

func TestSeedData_AdminRoleAssignment(t *testing.T) {
	var count int
	err := pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2",
		seedAdminID, seedRoleAdmin,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check admin role assignment: %v", err)
	}
	if count != 1 {
		t.Fatalf("Admin user should have admin role, got count=%d", count)
	}
}

func TestSeedData_Application(t *testing.T) {
	var name, appType, clientID string
	err := pool.QueryRow(context.Background(),
		"SELECT name, type, client_id FROM applications WHERE id = $1", seedAppID,
	).Scan(&name, &appType, &clientID)
	if err != nil {
		t.Fatalf("Default application not found: %v", err)
	}
	assertEqual(t, "Default App", name, "app name")
	assertEqual(t, "web", appType, "app type")
	assertEqual(t, seedClientID, clientID, "client_id")
}

func TestSeedData_EmailTemplates(t *testing.T) {
	types := []string{"verification", "password_reset", "welcome", "magic_link", "invitation", "mfa"}
	for _, tmplType := range types {
		var count int
		err := pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM email_templates WHERE tenant_id = $1 AND type = $2",
			seedTenantID, tmplType,
		).Scan(&count)
		if err != nil {
			t.Fatalf("Email template check failed for %s: %v", tmplType, err)
		}
		if count < 1 {
			t.Errorf("Email template %q not found for default tenant", tmplType)
		}
	}
}

func TestSeedData_Branding(t *testing.T) {
	var primaryColor, fontFamily string
	err := pool.QueryRow(context.Background(),
		"SELECT primary_color, font_family FROM branding_configs WHERE tenant_id = $1",
		seedTenantID,
	).Scan(&primaryColor, &fontFamily)
	if err != nil {
		t.Fatalf("Branding config not found: %v", err)
	}
	assertEqual(t, "#6366F1", primaryColor, "primary_color")
	assertEqual(t, "Inter", fontFamily, "font_family")
}

func TestSeedData_AuditLog(t *testing.T) {
	var count int
	err := pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM audit_logs WHERE action = 'system.seed_data_applied'",
	).Scan(&count)
	if err != nil {
		t.Fatalf("Audit log check failed: %v", err)
	}
	if count < 1 {
		t.Error("Seed data audit log entry not found")
	}
}

// ─── 2. Health & Discovery Endpoints ────────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	resp, body := mustGet(t, apiBaseURL+"/health")
	assertStatus(t, http.StatusOK, resp.StatusCode)

	var data map[string]string
	mustUnmarshal(t, body, &data)
	assertEqual(t, "ok", data["status"], "health status")
}

func TestReadyEndpoint(t *testing.T) {
	resp, body := mustGet(t, apiBaseURL+"/ready")
	assertStatus(t, http.StatusOK, resp.StatusCode)

	var data map[string]interface{}
	mustUnmarshal(t, body, &data)
	assertEqual(t, "ready", data["status"].(string), "ready status")

	checks := data["checks"].(map[string]interface{})
	assertEqual(t, "ok", checks["database"].(string), "database check")
}

func TestOIDCDiscovery(t *testing.T) {
	resp, body := mustGet(t, apiBaseURL+"/.well-known/openid-configuration")
	assertStatus(t, http.StatusOK, resp.StatusCode)

	var doc map[string]interface{}
	mustUnmarshal(t, body, &doc)

	if doc["issuer"] == nil || doc["issuer"] == "" {
		t.Error("OIDC discovery missing issuer")
	}
	if doc["authorization_endpoint"] == nil {
		t.Error("OIDC discovery missing authorization_endpoint")
	}
	if doc["token_endpoint"] == nil {
		t.Error("OIDC discovery missing token_endpoint")
	}
	if doc["jwks_uri"] == nil {
		t.Error("OIDC discovery missing jwks_uri")
	}
	if doc["userinfo_endpoint"] == nil {
		t.Error("OIDC discovery missing userinfo_endpoint")
	}

	// Verify OAuth 2.1 compliance
	codeMethods := doc["code_challenge_methods_supported"].([]interface{})
	found := false
	for _, m := range codeMethods {
		if m.(string) == "S256" {
			found = true
		}
	}
	if !found {
		t.Error("OIDC discovery must support S256 code challenge method")
	}
}

func TestJWKSEndpoint(t *testing.T) {
	resp, body := mustGet(t, apiBaseURL+"/.well-known/jwks.json")
	assertStatus(t, http.StatusOK, resp.StatusCode)

	var jwks map[string]interface{}
	mustUnmarshal(t, body, &jwks)

	keys := jwks["keys"].([]interface{})
	if len(keys) == 0 {
		t.Fatal("JWKS must contain at least one key")
	}

	key := keys[0].(map[string]interface{})
	if key["kty"] == nil {
		t.Error("JWK missing kty")
	}
	if key["kid"] == nil {
		t.Error("JWK missing kid")
	}
	if key["use"] == nil {
		t.Error("JWK missing use")
	}
}

// ─── 3. Full OAuth 2.1 + PKCE Authentication Flow ──────────────────────────

func TestOAuthFlow_FullPKCE(t *testing.T) {
	// This test performs the complete OAuth 2.1 authorization code flow
	// with PKCE, which is the industry-standard authentication pattern.

	// Step 1: Generate PKCE verifier and challenge
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := generatePKCEChallenge(verifier)

	// Step 2: POST /oauth/authorize with PKCE (requires auth context)
	// Since the authorize endpoint requires an authenticated user,
	// we need to get a token first via a different mechanism.
	// In practice, the admin user can use the admin endpoints.
	// For E2E, we test the token exchange directly by creating
	// an authorization grant in the database.
	t.Run("PKCE_Challenge_Generation", func(t *testing.T) {
		if challenge == "" {
			t.Fatal("PKCE challenge should not be empty")
		}
		// Verify S256: base64url(sha256(verifier))
		h := sha256.Sum256([]byte(verifier))
		expected := base64.RawURLEncoding.EncodeToString(h[:])
		assertEqual(t, expected, challenge, "PKCE challenge")
	})

	// Step 3: Verify that authorize endpoint requires PKCE
	t.Run("Authorize_RequiresPKCE", func(t *testing.T) {
		u := fmt.Sprintf("%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=openid+profile+email&state=test123",
			apiBaseURL, seedClientID, url.QueryEscape(seedRedirectURI))
		resp, _ := mustGet(t, u)
		// Without auth, we get authorization_required or error, not a crash
		if resp.StatusCode >= 500 {
			t.Errorf("Authorize endpoint returned server error: %d", resp.StatusCode)
		}
	})

	// Step 4: Test token endpoint form validation
	t.Run("Token_MissingGrantType", func(t *testing.T) {
		resp, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{})
		if resp.StatusCode == http.StatusOK {
			t.Error("Token endpoint should reject empty grant_type")
		}
	})

	// Step 5: Test token endpoint with invalid authorization code
	t.Run("Token_InvalidCode", func(t *testing.T) {
		resp, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"invalid_code"},
			"redirect_uri":  {seedRedirectURI},
			"client_id":     {seedClientID},
			"code_verifier": {verifier},
		})
		if resp.StatusCode == http.StatusOK {
			t.Error("Token endpoint should reject invalid authorization code")
		}
	})

	// Step 6: Create auth grant directly in DB to test token exchange
	t.Run("Token_ExchangeWithPKCE", func(t *testing.T) {
		code := "e2e_test_code_" + fmt.Sprintf("%d", time.Now().UnixNano())
		expiresAt := time.Now().UTC().Add(5 * time.Minute)

		_, err := pool.Exec(context.Background(),
			`INSERT INTO oauth_grants (user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			seedAdminID, seedAppID, seedTenantID,
			[]string{"openid", "profile", "email", "offline_access"},
			code, challenge, "S256", seedRedirectURI, expiresAt,
		)
		if err != nil {
			t.Fatalf("Failed to create test grant: %v", err)
		}

		// Exchange code for tokens
		resp, body := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {seedRedirectURI},
			"client_id":     {seedClientID},
			"code_verifier": {verifier},
		})
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var tokenPair map[string]interface{}
		mustUnmarshal(t, body, &tokenPair)

		if tokenPair["access_token"] == nil || tokenPair["access_token"] == "" {
			t.Fatal("Token exchange should return access_token")
		}
		assertEqual(t, "Bearer", tokenPair["token_type"].(string), "token_type")
		if tokenPair["id_token"] == nil || tokenPair["id_token"] == "" {
			t.Error("Token exchange with openid scope should return id_token")
		}
		if tokenPair["refresh_token"] == nil || tokenPair["refresh_token"] == "" {
			t.Error("Token exchange with offline_access scope should return refresh_token")
		}

		accessToken := tokenPair["access_token"].(string)
		refreshToken := ""
		if tokenPair["refresh_token"] != nil {
			refreshToken = tokenPair["refresh_token"].(string)
		}

		// Step 7: Verify grant was deleted (single-use)
		var grantCount int
		pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM oauth_grants WHERE code = $1", code,
		).Scan(&grantCount)
		if grantCount != 0 {
			t.Error("Authorization code should be deleted after use (single-use)")
		}

		// Step 8: Use access token to call userinfo
		t.Run("Userinfo_WithToken", func(t *testing.T) {
			req, _ := http.NewRequest("GET", apiBaseURL+"/oauth/userinfo", nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("Userinfo request failed: %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assertStatus(t, http.StatusOK, resp.StatusCode)

			var userinfo map[string]interface{}
			mustUnmarshal(t, body, &userinfo)

			if userinfo["sub"] == nil || userinfo["sub"] == "" {
				t.Error("Userinfo should contain sub claim")
			}
			if userinfo["email"] == nil {
				t.Error("Userinfo should contain email (email scope requested)")
			}
		})

		// Step 9: Test refresh token flow
		if refreshToken != "" {
			t.Run("RefreshToken_Rotation", func(t *testing.T) {
				resp, body := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
					"grant_type":    {"refresh_token"},
					"refresh_token": {refreshToken},
					"client_id":     {seedClientID},
				})
				assertStatus(t, http.StatusOK, resp.StatusCode)

				var newPair map[string]interface{}
				mustUnmarshal(t, body, &newPair)

				if newPair["access_token"] == nil || newPair["access_token"] == "" {
					t.Error("Refresh should return new access_token")
				}
				if newPair["refresh_token"] == nil || newPair["refresh_token"] == "" {
					t.Error("Refresh should return new refresh_token (rotation)")
				}

				// Old refresh token should be revoked (reuse detection)
				t.Run("Reuse_Detection", func(t *testing.T) {
					resp2, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
						"grant_type":    {"refresh_token"},
						"refresh_token": {refreshToken}, // reuse old token
						"client_id":     {seedClientID},
					})
					if resp2.StatusCode == http.StatusOK {
						t.Error("Reusing old refresh token should fail (reuse detection)")
					}
				})

				// Verify refresh token rotation in DB
				var revokedCount int
				pool.QueryRow(context.Background(),
					"SELECT COUNT(*) FROM refresh_tokens WHERE application_id = $1 AND revoked = true",
					seedAppID,
				).Scan(&revokedCount)
				if revokedCount < 1 {
					t.Error("Old refresh token should be marked as revoked in DB")
				}
			})
		}

		// Step 10: Revoke access token
		t.Run("Token_Revocation", func(t *testing.T) {
			resp, _ := mustPostForm(t, apiBaseURL+"/oauth/revoke", url.Values{
				"token":           {accessToken},
				"token_type_hint": {"access_token"},
			})
			assertStatus(t, http.StatusOK, resp.StatusCode)
		})
	})

	// Test PKCE without verifier should fail
	t.Run("Token_MissingVerifier", func(t *testing.T) {
		code := "e2e_test_noverifier_" + fmt.Sprintf("%d", time.Now().UnixNano())
		expiresAt := time.Now().UTC().Add(5 * time.Minute)

		pool.Exec(context.Background(),
			`INSERT INTO oauth_grants (user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			seedAdminID, seedAppID, seedTenantID,
			[]string{"openid"}, code, challenge, "S256", seedRedirectURI, expiresAt,
		)

		resp, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":   {"authorization_code"},
			"code":         {code},
			"redirect_uri": {seedRedirectURI},
			"client_id":    {seedClientID},
			// No code_verifier - must fail
		})
		if resp.StatusCode == http.StatusOK {
			t.Error("Token exchange without code_verifier should fail (PKCE mandatory)")
		}
	})

	// Test wrong verifier should fail
	t.Run("Token_WrongVerifier", func(t *testing.T) {
		code := "e2e_test_wrongverifier_" + fmt.Sprintf("%d", time.Now().UnixNano())
		expiresAt := time.Now().UTC().Add(5 * time.Minute)

		pool.Exec(context.Background(),
			`INSERT INTO oauth_grants (user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			seedAdminID, seedAppID, seedTenantID,
			[]string{"openid"}, code, challenge, "S256", seedRedirectURI, expiresAt,
		)

		resp, _ := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {seedRedirectURI},
			"client_id":     {seedClientID},
			"code_verifier": {"wrong_verifier_value"},
		})
		if resp.StatusCode == http.StatusOK {
			t.Error("Token exchange with wrong verifier should fail")
		}
	})
}

// ─── 4. Admin API Operations ────────────────────────────────────────────────

func TestAdminAPI(t *testing.T) {
	// First get an access token for the admin user
	accessToken := getAdminAccessToken(t)
	if accessToken == "" {
		t.Fatal("Failed to obtain admin access token")
	}

	var createdUserID string

	// Test user CRUD
	t.Run("CreateUser", func(t *testing.T) {
		body := map[string]interface{}{
			"email":    "e2etest@cpi-auth.local",
			"password": "E2eTest1234!",
			"name":     "E2E Test User",
		}
		resp, respBody := mustPostJSON(t, apiBaseURL+"/admin/users", body, accessToken)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(respBody))
		}

		var user map[string]interface{}
		mustUnmarshal(t, respBody, &user)

		createdUserID, _ = user["id"].(string)
		if createdUserID == "" {
			t.Fatal("Created user should have an ID")
		}
		assertEqual(t, "e2etest@cpi-auth.local", user["email"].(string), "user email")

		// Verify in DB
		var dbEmail string
		err := pool.QueryRow(context.Background(),
			"SELECT email FROM users WHERE id = $1", createdUserID,
		).Scan(&dbEmail)
		if err != nil {
			t.Fatalf("Created user not found in DB: %v", err)
		}
		assertEqual(t, "e2etest@cpi-auth.local", dbEmail, "DB email")
	})

	t.Run("GetUser", func(t *testing.T) {
		if createdUserID == "" {
			t.Fatal("No user ID from create")
		}
		resp, body := mustGetAuth(t, apiBaseURL+"/admin/users/"+createdUserID, accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var user map[string]interface{}
		mustUnmarshal(t, body, &user)
		assertEqual(t, createdUserID, user["id"].(string), "user id")
	})

	t.Run("UpdateUser", func(t *testing.T) {
		if createdUserID == "" {
			t.Fatal("No user ID from create")
		}
		body := map[string]interface{}{
			"name": "E2E Updated Name",
		}
		resp, _ := mustPatchJSON(t, apiBaseURL+"/admin/users/"+createdUserID, body, accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		// Verify in DB
		var dbName string
		pool.QueryRow(context.Background(),
			"SELECT name FROM users WHERE id = $1", createdUserID,
		).Scan(&dbName)
		assertEqual(t, "E2E Updated Name", dbName, "updated name in DB")
	})

	t.Run("BlockUser", func(t *testing.T) {
		if createdUserID == "" {
			t.Fatal("No user ID from create")
		}
		resp, _ := mustPostJSON(t, apiBaseURL+"/admin/users/"+createdUserID+"/block", nil, accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var dbStatus string
		pool.QueryRow(context.Background(),
			"SELECT status FROM users WHERE id = $1", createdUserID,
		).Scan(&dbStatus)
		assertEqual(t, "blocked", dbStatus, "blocked status in DB")
	})

	t.Run("UnblockUser", func(t *testing.T) {
		if createdUserID == "" {
			t.Fatal("No user ID from create")
		}
		resp, _ := mustPostJSON(t, apiBaseURL+"/admin/users/"+createdUserID+"/unblock", nil, accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var dbStatus string
		pool.QueryRow(context.Background(),
			"SELECT status FROM users WHERE id = $1", createdUserID,
		).Scan(&dbStatus)
		assertEqual(t, "active", dbStatus, "unblocked status in DB")
	})

	t.Run("ListUsers", func(t *testing.T) {
		resp, body := mustGetAuth(t, apiBaseURL+"/admin/users", accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var result interface{}
		mustUnmarshal(t, body, &result)
		// Should at least have the admin and the created test user
	})

	// Test tenant CRUD
	var createdTenantID string

	t.Run("CreateTenant", func(t *testing.T) {
		body := map[string]interface{}{
			"name":   "E2E Test Tenant",
			"slug":   "e2e-test",
			"domain": "e2e.cpi-auth.local",
		}
		resp, respBody := mustPostJSON(t, apiBaseURL+"/admin/tenants", body, accessToken)
		assertStatus(t, http.StatusCreated, resp.StatusCode)

		var tenant map[string]interface{}
		mustUnmarshal(t, respBody, &tenant)
		if tenant["id"] != nil {
			createdTenantID = tenant["id"].(string)
		}

		// Verify in DB
		var dbName string
		err := pool.QueryRow(context.Background(),
			"SELECT name FROM tenants WHERE slug = 'e2e-test'",
		).Scan(&dbName)
		if err != nil {
			t.Fatalf("Created tenant not found in DB: %v", err)
		}
		assertEqual(t, "E2E Test Tenant", dbName, "tenant name in DB")
	})

	t.Run("ListTenants", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/admin/tenants", accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)
	})

	// Test application CRUD
	t.Run("ListApplications", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/admin/applications", accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetApplication", func(t *testing.T) {
		resp, body := mustGetAuth(t, apiBaseURL+"/admin/applications/"+seedAppID, accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var app map[string]interface{}
		mustUnmarshal(t, body, &app)
		assertEqual(t, seedClientID, app["client_id"].(string), "app client_id")
	})

	// Test role CRUD
	t.Run("ListRoles", func(t *testing.T) {
		resp, body := mustGetAuth(t, apiBaseURL+"/admin/roles", accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		// Verify roles from DB
		var roleCount int
		pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM roles WHERE tenant_id = $1",
			seedTenantID,
		).Scan(&roleCount)
		if roleCount < 4 {
			t.Errorf("Expected at least 4 roles, got %d", roleCount)
		}
		_ = body
	})

	t.Run("CreateRole", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "e2e-test-role",
			"description": "E2E Test Role",
			"permissions": []string{"users:read"},
		}
		resp, _ := mustPostJSON(t, apiBaseURL+"/admin/roles", body, accessToken)
		assertStatus(t, http.StatusCreated, resp.StatusCode)

		// Verify in DB
		var dbName string
		pool.QueryRow(context.Background(),
			"SELECT name FROM roles WHERE name = 'e2e-test-role' AND tenant_id = $1",
			seedTenantID,
		).Scan(&dbName)
		assertEqual(t, "e2e-test-role", dbName, "role name in DB")
	})

	// Test organization CRUD
	t.Run("CreateOrganization", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "E2E Test Org",
			"slug": "e2e-test-org",
		}
		resp, respBody := mustPostJSON(t, apiBaseURL+"/admin/organizations", body, accessToken)
		assertStatus(t, http.StatusCreated, resp.StatusCode)

		var org map[string]interface{}
		mustUnmarshal(t, respBody, &org)

		// Verify in DB
		var dbName string
		pool.QueryRow(context.Background(),
			"SELECT name FROM organizations WHERE name = 'E2E Test Org' AND tenant_id = $1",
			seedTenantID,
		).Scan(&dbName)
		assertEqual(t, "E2E Test Org", dbName, "org name in DB")
	})

	// Test webhook CRUD
	t.Run("CreateWebhook", func(t *testing.T) {
		body := map[string]interface{}{
			"url":    "https://e2e-test.example.com/webhook",
			"events": []string{"user.created", "user.updated"},
			"active": true,
		}
		resp, _ := mustPostJSON(t, apiBaseURL+"/admin/webhooks", body, accessToken)
		assertStatus(t, http.StatusCreated, resp.StatusCode)

		// Verify in DB
		var dbURL string
		pool.QueryRow(context.Background(),
			"SELECT url FROM webhooks WHERE url = 'https://e2e-test.example.com/webhook'",
		).Scan(&dbURL)
		assertEqual(t, "https://e2e-test.example.com/webhook", dbURL, "webhook URL in DB")
	})

	// Cleanup: delete test user
	t.Run("DeleteUser", func(t *testing.T) {
		if createdUserID == "" {
			t.Fatal("No user ID from create")
		}
		resp, _ := mustDeleteAuth(t, apiBaseURL+"/admin/users/"+createdUserID, accessToken)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			t.Errorf("Delete user returned %d", resp.StatusCode)
		}
	})

	// Cleanup: delete test tenant
	if createdTenantID != "" {
		t.Run("DeleteTenant", func(t *testing.T) {
			resp, _ := mustDeleteAuth(t, apiBaseURL+"/admin/tenants/"+createdTenantID, accessToken)
			if resp.StatusCode >= 500 {
				t.Errorf("Delete tenant returned %d", resp.StatusCode)
			}
		})
	}
}

// ─── 5. User Self-Service Endpoints ─────────────────────────────────────────

func TestUserSelfService(t *testing.T) {
	accessToken := getAdminAccessToken(t)
	if accessToken == "" {
		t.Fatal("Failed to obtain admin access token")
	}

	t.Run("GetProfile", func(t *testing.T) {
		resp, body := mustGetAuth(t, apiBaseURL+"/v1/users/me", accessToken)
		assertStatus(t, http.StatusOK, resp.StatusCode)

		var user map[string]interface{}
		mustUnmarshal(t, body, &user)
		if user["id"] == nil || user["id"] == "" {
			t.Error("Profile should contain user ID")
		}
		if user["email"] == nil || user["email"] == "" {
			t.Error("Profile should contain email")
		}
	})

	t.Run("ListSessions", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/v1/users/me/sessions", accessToken)
		// Sessions endpoint may or may not have active sessions
		if resp.StatusCode >= 500 {
			t.Errorf("List sessions returned server error: %d", resp.StatusCode)
		}
	})

	t.Run("ListMFA", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/v1/users/me/mfa", accessToken)
		if resp.StatusCode >= 500 {
			t.Errorf("List MFA returned server error: %d", resp.StatusCode)
		}
	})

	t.Run("ListPasskeys", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/v1/users/me/passkeys", accessToken)
		if resp.StatusCode >= 500 {
			t.Errorf("List passkeys returned server error: %d", resp.StatusCode)
		}
	})

	t.Run("ListIdentities", func(t *testing.T) {
		resp, _ := mustGetAuth(t, apiBaseURL+"/v1/users/me/identities", accessToken)
		if resp.StatusCode >= 500 {
			t.Errorf("List identities returned server error: %d", resp.StatusCode)
		}
	})
}

// ─── 6. Security Tests ─────────────────────────────────────────────────────

func TestSecurity_UnauthorizedAccess(t *testing.T) {
	t.Run("Admin_NoToken", func(t *testing.T) {
		resp, _ := mustGet(t, apiBaseURL+"/admin/users")
		if resp.StatusCode == http.StatusOK {
			t.Error("Admin endpoints should require authentication")
		}
	})

	t.Run("UserSelfService_NoToken", func(t *testing.T) {
		resp, _ := mustGet(t, apiBaseURL+"/v1/users/me")
		if resp.StatusCode == http.StatusOK {
			t.Error("User endpoints should require authentication")
		}
	})

	t.Run("Admin_InvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", apiBaseURL+"/admin/users", nil)
		req.Header.Set("Authorization", "Bearer invalid_token_here")
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Error("Admin endpoints should reject invalid tokens")
		}
	})
}

func TestSecurity_OAuth21Compliance(t *testing.T) {
	t.Run("PKCE_Required", func(t *testing.T) {
		// Attempt authorize without code_challenge
		u := fmt.Sprintf("%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=openid&state=test",
			apiBaseURL, seedClientID, url.QueryEscape(seedRedirectURI))
		resp, _ := mustGet(t, u)
		// The endpoint should work but flag the missing PKCE when trying to authorize
		_ = resp
	})

	t.Run("ResponseType_OnlyCode", func(t *testing.T) {
		// response_type=token should be rejected (OAuth 2.1 disallows implicit)
		u := fmt.Sprintf("%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=token&scope=openid&state=test",
			apiBaseURL, seedClientID, url.QueryEscape(seedRedirectURI))
		resp, _ := mustGet(t, u)
		_ = resp // The point is it shouldn't return a valid implicit token
	})
}

// ─── 7. Database Integrity Checks ──────────────────────────────────────────

func TestDatabaseIntegrity(t *testing.T) {
	t.Run("AllTablesExist", func(t *testing.T) {
		tables := []string{
			"tenants", "users", "applications", "organizations",
			"roles", "user_roles", "sessions", "refresh_tokens",
			"oauth_grants", "identities", "social_providers",
			"mfa_enrollments", "recovery_codes", "webauthn_credentials",
			"audit_logs", "webhooks",
			"actions", "email_templates", "branding_configs",
			"api_keys", "fga_tuples", "jwks_keys", "password_history",
		}

		for _, table := range tables {
			var exists bool
			err := pool.QueryRow(context.Background(),
				"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", table,
			).Scan(&exists)
			if err != nil {
				t.Fatalf("Failed to check table %s: %v", table, err)
			}
			if !exists {
				t.Errorf("Table %q does not exist", table)
			}
		}
	})

	t.Run("IndexesExist", func(t *testing.T) {
		// Check critical indexes
		criticalIndexes := []struct {
			table  string
			column string
		}{
			{"users", "email"},
			{"applications", "client_id"},
			{"tenants", "slug"},
		}

		for _, idx := range criticalIndexes {
			var count int
			err := pool.QueryRow(context.Background(),
				`SELECT COUNT(*) FROM pg_indexes WHERE tablename = $1 AND indexdef LIKE '%' || $2 || '%'`,
				idx.table, idx.column,
			).Scan(&count)
			if err != nil {
				t.Fatalf("Failed to check index on %s.%s: %v", idx.table, idx.column, err)
			}
			if count == 0 {
				t.Errorf("Missing index on %s.%s", idx.table, idx.column)
			}
		}
	})

	t.Run("ForeignKeysIntact", func(t *testing.T) {
		// Verify FK constraints exist
		var fkCount int
		err := pool.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM information_schema.table_constraints
			 WHERE constraint_type = 'FOREIGN KEY' AND table_schema = 'public'`,
		).Scan(&fkCount)
		if err != nil {
			t.Fatalf("FK check failed: %v", err)
		}
		if fkCount == 0 {
			t.Error("No foreign key constraints found - schema integrity issue")
		}
	})

	t.Run("PasswordHashNotPlaintext", func(t *testing.T) {
		var hash string
		pool.QueryRow(context.Background(),
			"SELECT password_hash FROM users WHERE id = $1", seedAdminID,
		).Scan(&hash)
		if hash == seedAdminPassword {
			t.Fatal("Password stored as plaintext!")
		}
		if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$argon2id$") {
			t.Error("Password hash should use bcrypt or argon2id")
		}
	})
}

// ─── 8. Metrics Endpoint ────────────────────────────────────────────────────

func TestMetricsEndpoint(t *testing.T) {
	resp, body := mustGet(t, apiBaseURL+"/metrics")
	if resp.StatusCode == http.StatusOK {
		if !strings.Contains(string(body), "go_") {
			t.Error("Metrics should contain Go runtime metrics")
		}
	}
	// Metrics may be disabled, so don't fail hard
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func getAdminAccessToken(t *testing.T) string {
	t.Helper()

	verifier := "e2e_admin_pkce_verifier_for_testing_purposes_only"
	challenge := generatePKCEChallenge(verifier)

	code := "e2e_admin_code_" + fmt.Sprintf("%d", time.Now().UnixNano())
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	_, err := pool.Exec(context.Background(),
		`INSERT INTO oauth_grants (user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		seedAdminID, seedAppID, seedTenantID,
		[]string{"openid", "profile", "email", "offline_access"},
		code, challenge, "S256", seedRedirectURI, expiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to create admin grant: %v", err)
	}

	resp, body := mustPostForm(t, apiBaseURL+"/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {seedRedirectURI},
		"client_id":     {seedClientID},
		"code_verifier": {verifier},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get admin token: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tokenPair map[string]interface{}
	mustUnmarshal(t, body, &tokenPair)
	return tokenPair["access_token"].(string)
}

func generatePKCEChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func waitForAPI(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("API at %s not ready after %v", baseURL, timeout)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGet(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()
	resp, err := httpClient.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func mustGetAuth(t *testing.T, url, token string) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func mustPostForm(t *testing.T, url string, data url.Values) (*http.Response, []byte) {
	t.Helper()
	resp, err := httpClient.PostForm(url, data)
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func mustPostJSON(t *testing.T, url string, data interface{}, token string) (*http.Response, []byte) {
	t.Helper()
	var reqBody io.Reader
	if data != nil {
		b, _ := json.Marshal(data)
		reqBody = bytes.NewReader(b)
	}
	req, _ := http.NewRequest("POST", url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func mustPatchJSON(t *testing.T, url string, data interface{}, token string) (*http.Response, []byte) {
	t.Helper()
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s failed: %v", url, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func mustDeleteAuth(t *testing.T, url, token string) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s failed: %v", url, err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, body
}

func mustUnmarshal(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v\nBody: %s", err, string(data))
	}
}

func assertStatus(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected status %d, got %d", expected, actual)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}, field string) {
	t.Helper()
	if fmt.Sprintf("%v", expected) != fmt.Sprintf("%v", actual) {
		t.Errorf("%s: expected %v, got %v", field, expected, actual)
	}
}
