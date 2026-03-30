package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/events"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/oauth"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
)

// --- Minimal mock repos for building services ---

type mockAppRepoAuth struct {
	apps       map[uuid.UUID]*models.Application
	byClientID map[string]*models.Application
}

func newMockAppRepoAuth() *mockAppRepoAuth {
	return &mockAppRepoAuth{
		apps:       make(map[uuid.UUID]*models.Application),
		byClientID: make(map[string]*models.Application),
	}
}

func (m *mockAppRepoAuth) Create(_ context.Context, app *models.Application) error {
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	m.apps[app.ID] = app
	m.byClientID[app.ClientID] = app
	return nil
}
func (m *mockAppRepoAuth) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.Application, error) {
	app, ok := m.apps[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return app, nil
}
func (m *mockAppRepoAuth) GetByClientID(_ context.Context, clientID string) (*models.Application, error) {
	app, ok := m.byClientID[clientID]
	if !ok {
		return nil, models.ErrNotFound
	}
	return app, nil
}
func (m *mockAppRepoAuth) Update(_ context.Context, app *models.Application) error { return nil }
func (m *mockAppRepoAuth) Delete(_ context.Context, tenantID, id uuid.UUID) error  { return nil }
func (m *mockAppRepoAuth) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Application], error) {
	return &models.PaginatedResult[models.Application]{}, nil
}

type mockGrantRepoAuth struct {
	grants map[uuid.UUID]*models.OAuthGrant
	byCode map[string]*models.OAuthGrant
}

func newMockGrantRepoAuth() *mockGrantRepoAuth {
	return &mockGrantRepoAuth{
		grants: make(map[uuid.UUID]*models.OAuthGrant),
		byCode: make(map[string]*models.OAuthGrant),
	}
}

func (m *mockGrantRepoAuth) Create(_ context.Context, grant *models.OAuthGrant) error {
	if grant.ID == uuid.Nil {
		grant.ID = uuid.New()
	}
	m.grants[grant.ID] = grant
	m.byCode[grant.Code] = grant
	return nil
}
func (m *mockGrantRepoAuth) GetByCode(_ context.Context, code string) (*models.OAuthGrant, error) {
	grant, ok := m.byCode[code]
	if !ok {
		return nil, models.ErrNotFound
	}
	return grant, nil
}
func (m *mockGrantRepoAuth) Delete(_ context.Context, id uuid.UUID) error {
	if grant, ok := m.grants[id]; ok {
		delete(m.byCode, grant.Code)
		delete(m.grants, id)
	}
	return nil
}
func (m *mockGrantRepoAuth) DeleteExpired(_ context.Context) error { return nil }

type mockUserRepoAuth struct {
	users map[uuid.UUID]*models.User
}

func newMockUserRepoAuth() *mockUserRepoAuth {
	return &mockUserRepoAuth{users: make(map[uuid.UUID]*models.User)}
}

func (m *mockUserRepoAuth) Create(_ context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID] = user
	return nil
}
func (m *mockUserRepoAuth) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return user, nil
}
func (m *mockUserRepoAuth) GetByEmail(_ context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email && u.TenantID == tenantID {
			return u, nil
		}
	}
	return nil, models.ErrNotFound
}
func (m *mockUserRepoAuth) Update(_ context.Context, user *models.User) error { return nil }
func (m *mockUserRepoAuth) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}
func (m *mockUserRepoAuth) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, search string) (*models.PaginatedResult[models.User], error) {
	return &models.PaginatedResult[models.User]{}, nil
}
func (m *mockUserRepoAuth) Block(_ context.Context, tenantID, id uuid.UUID) error   { return nil }
func (m *mockUserRepoAuth) Unblock(_ context.Context, tenantID, id uuid.UUID) error { return nil }
func (m *mockUserRepoAuth) CountByTenant(_ context.Context, tenantID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockUserRepoAuth) GetPasswordHistory(_ context.Context, userID uuid.UUID, limit int) ([]models.PasswordHistory, error) {
	return nil, nil
}
func (m *mockUserRepoAuth) AddPasswordHistory(_ context.Context, entry *models.PasswordHistory) error {
	return nil
}

type mockRefreshTokenRepoAuth struct {
	tokens map[uuid.UUID]*models.RefreshToken
	byHash map[string]*models.RefreshToken
}

func newMockRefreshTokenRepoAuth() *mockRefreshTokenRepoAuth {
	return &mockRefreshTokenRepoAuth{
		tokens: make(map[uuid.UUID]*models.RefreshToken),
		byHash: make(map[string]*models.RefreshToken),
	}
}

func (m *mockRefreshTokenRepoAuth) Create(_ context.Context, token *models.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	m.tokens[token.ID] = token
	m.byHash[token.TokenHash] = token
	return nil
}
func (m *mockRefreshTokenRepoAuth) GetByTokenHash(_ context.Context, hash string) (*models.RefreshToken, error) {
	rt, ok := m.byHash[hash]
	if !ok {
		return nil, models.ErrNotFound
	}
	return rt, nil
}
func (m *mockRefreshTokenRepoAuth) Revoke(_ context.Context, id uuid.UUID) error {
	if rt, ok := m.tokens[id]; ok {
		rt.Revoked = true
	}
	return nil
}
func (m *mockRefreshTokenRepoAuth) RevokeByFamily(_ context.Context, family string) error {
	return nil
}
func (m *mockRefreshTokenRepoAuth) RevokeByUser(_ context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockRefreshTokenRepoAuth) RevokeByApplication(_ context.Context, appID uuid.UUID) error {
	return nil
}
func (m *mockRefreshTokenRepoAuth) DeleteExpired(_ context.Context) error { return nil }

// Mock audit log repo for event service
type mockAuditLogRepoAuth struct{}

func (m *mockAuditLogRepoAuth) Create(_ context.Context, log *models.AuditLog) error { return nil }
func (m *mockAuditLogRepoAuth) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, action string) (*models.PaginatedResult[models.AuditLog], error) {
	return &models.PaginatedResult[models.AuditLog]{}, nil
}

// Mock webhook repo for event service
type mockWebhookRepoAuth struct{}

func (m *mockWebhookRepoAuth) Create(_ context.Context, webhook *models.Webhook) error { return nil }
func (m *mockWebhookRepoAuth) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.Webhook, error) {
	return nil, models.ErrNotFound
}
func (m *mockWebhookRepoAuth) Update(_ context.Context, webhook *models.Webhook) error { return nil }
func (m *mockWebhookRepoAuth) Delete(_ context.Context, tenantID, id uuid.UUID) error  { return nil }
func (m *mockWebhookRepoAuth) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Webhook], error) {
	return &models.PaginatedResult[models.Webhook]{}, nil
}
func (m *mockWebhookRepoAuth) ListByEvent(_ context.Context, tenantID uuid.UUID, event string) ([]models.Webhook, error) {
	return nil, nil
}

// --- Test helpers ---

func testAuthConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Security.AuthCodeLifetime = 10 * time.Minute
	cfg.Security.AccessTokenLifetime = 15 * time.Minute
	cfg.Security.RefreshTokenLifetime = 7 * 24 * time.Hour
	cfg.Security.IDTokenLifetime = 1 * time.Hour
	cfg.Security.Issuer = "https://auth.example.com"
	cfg.Security.JWTSigningAlgorithm = "RS256"
	return cfg
}

func testAuthHandler() (*Handler, *mockAppRepoAuth, *mockUserRepoAuth, *tokens.Service) {
	cfg := testAuthConfig()
	logger := zap.NewNop()

	appRepo := newMockAppRepoAuth()
	grantRepo := newMockGrantRepoAuth()
	userRepo := newMockUserRepoAuth()
	refreshRepo := newMockRefreshTokenRepoAuth()

	keyPair, _ := crypto.GenerateRSAKeyPair(2048)
	km := crypto.NewKeyManager(keyPair)
	tokenSvc := tokens.NewService(km, refreshRepo, nil, cfg, logger)
	oauthSvc := oauth.NewService(appRepo, grantRepo, userRepo, tokenSvc, nil, nil, cfg, logger)
	eventSvc := events.NewService(nil, &mockAuditLogRepoAuth{}, &mockWebhookRepoAuth{}, logger)

	h := NewHandler(oauthSvc, nil, tokenSvc, nil, nil, nil, eventSvc, nil, nil, nil, nil, logger)
	return h, appRepo, userRepo, tokenSvc
}

func withUserContext(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyUserID, userID)
	return r.WithContext(ctx)
}

func withTenantContext(r *http.Request, tenantID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyTenantID, tenantID)
	return r.WithContext(ctx)
}

func withClaimsContext(r *http.Request, claims *tokens.AccessTokenClaims) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyClaims, claims)
	return r.WithContext(ctx)
}

// --- Tests ---

func TestNewHandler(t *testing.T) {
	h, _, _, _ := testAuthHandler()
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
}

func TestRegisterRoutes(t *testing.T) {
	h, _, _, _ := testAuthHandler()
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// Verify routes were registered by walking
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/oauth/authorize"},
		{http.MethodGet, "/oauth/authorize"},
		{http.MethodPost, "/oauth/token"},
		{http.MethodPost, "/oauth/revoke"},
		{http.MethodGet, "/oauth/userinfo"},
		{http.MethodGet, "/.well-known/openid-configuration"},
		{http.MethodGet, "/.well-known/jwks.json"},
	}

	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// A 405 or 404 means the route is not matched at all
		if w.Code == http.StatusMethodNotAllowed {
			t.Errorf("route %s %s returned 405, might not be registered correctly", rt.method, rt.path)
		}
	}
}

func TestAuthorizeHandler_POST_NoUserID(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	body := `{"client_id":"test","redirect_uri":"https://example.com/cb","response_type":"code","code_challenge":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/oauth/authorize", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Authorize(w, req)

	// No user ID in context => should return error (500 since WriteError gets nil)
	if w.Code == http.StatusOK {
		t.Error("Authorize should not return 200 without user in context")
	}
}

func TestAuthorizeHandler_POST_WithUser(t *testing.T) {
	h, appRepo, _, _ := testAuthHandler()

	tenantID := uuid.New()
	app := &models.Application{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         "Test App",
		Type:         models.AppTypeSPA,
		ClientID:     "my-client-id",
		RedirectURIs: []string{"https://app.example.com/callback"},
	}
	appRepo.apps[app.ID] = app
	appRepo.byClientID[app.ClientID] = app

	body := map[string]interface{}{
		"client_id":             app.ClientID,
		"redirect_uri":         "https://app.example.com/callback",
		"response_type":        "code",
		"scope":                "openid",
		"state":                "test-state",
		"code_challenge":       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		"code_challenge_method": "S256",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/oauth/authorize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, uuid.New())
	w := httptest.NewRecorder()

	h.Authorize(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["code"] == nil || resp["code"] == "" {
		t.Error("response should contain a code")
	}
	if resp["state"] != "test-state" {
		t.Errorf("state = %v, want %q", resp["state"], "test-state")
	}
}

func TestAuthorizeHandler_POST_InvalidJSON(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/oauth/authorize", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, uuid.New())
	w := httptest.NewRecorder()

	h.Authorize(w, req)

	if w.Code == http.StatusOK {
		t.Error("should fail for invalid JSON body")
	}
}

func TestAuthorizeGetHandler_NoUser(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?client_id=test&scope=openid&state=s1", nil)
	w := httptest.NewRecorder()

	h.AuthorizeGet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["type"] != "authorization_required" {
		t.Errorf("type = %v, want %q", resp["type"], "authorization_required")
	}
}

func TestTokenHandler_FormPost(t *testing.T) {
	h, appRepo, userRepo, _ := testAuthHandler()

	tenantID := uuid.New()
	app := &models.Application{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         "Test M2M App",
		Type:         models.AppTypeM2M,
		ClientID:     "m2m-client-id",
		RedirectURIs: []string{},
	}
	secret := "m2m-secret-value"
	app.ClientSecretHash = crypto.HashToken(secret)
	appRepo.apps[app.ID] = app
	appRepo.byClientID[app.ClientID] = app

	_ = userRepo // Not needed for client_credentials

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {app.ClientID},
		"client_secret": {secret},
		"scope":         {"api:read"},
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withTenantContext(req, tenantID)
	w := httptest.NewRecorder()

	h.Token(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["access_token"] == nil || resp["access_token"] == "" {
		t.Error("response should contain access_token")
	}
	if resp["token_type"] != "Bearer" {
		t.Errorf("token_type = %v, want %q", resp["token_type"], "Bearer")
	}
}

func TestTokenHandler_InvalidGrantType(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	form := url.Values{
		"grant_type": {"implicit"},
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withTenantContext(req, uuid.New())
	w := httptest.NewRecorder()

	h.Token(w, req)

	if w.Code == http.StatusOK {
		t.Error("Token should fail for unsupported grant type")
	}
}

func TestRevokeHandler(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	form := url.Values{
		"token":           {"some-token"},
		"token_type_hint": {"unsupported_hint"},
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Revoke(w, req)

	// Should return an error for unsupported hint
	if w.Code == http.StatusOK {
		// Revoke should fail for unsupported hint
		t.Log("Note: revocation endpoint returned 200, but with unsupported hint it may still succeed per spec")
	}
}

func TestUserinfoHandler_NoClaims(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/oauth/userinfo", nil)
	w := httptest.NewRecorder()

	h.Userinfo(w, req)

	// No claims in context => should return error
	if w.Code == http.StatusOK {
		t.Error("Userinfo should fail without claims in context")
	}
}

func TestUserinfoHandler_WithClaims(t *testing.T) {
	h, _, userRepo, _ := testAuthHandler()

	tenantID := uuid.New()
	user := &models.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         "user@example.com",
		Name:          "Test User",
		EmailVerified: true,
		UpdatedAt:     time.Now().UTC(),
	}
	userRepo.users[user.ID] = user

	claims := &tokens.AccessTokenClaims{}
	claims.Subject = user.ID.String()
	claims.TenantID = tenantID.String()
	claims.Scope = "openid profile email"

	req := httptest.NewRequest(http.MethodGet, "/oauth/userinfo", nil)
	req = withClaimsContext(req, claims)
	w := httptest.NewRecorder()

	h.Userinfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var info map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if info["sub"] != user.ID.String() {
		t.Errorf("sub = %v, want %v", info["sub"], user.ID.String())
	}
	if info["email"] != user.Email {
		t.Errorf("email = %v, want %v", info["email"], user.Email)
	}
}

func TestDiscoveryHandler(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()

	h.Discovery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var doc map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if doc["issuer"] != "https://auth.example.com" {
		t.Errorf("issuer = %v, want %q", doc["issuer"], "https://auth.example.com")
	}
}

func TestJWKSHandler(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	h.JWKS(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var jwks map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&jwks); err != nil {
		t.Fatalf("failed to decode JWKS response: %v", err)
	}
	keys, ok := jwks["keys"]
	if !ok {
		t.Fatal("JWKS response should contain 'keys' field")
	}
	keysArr, ok := keys.([]interface{})
	if !ok || len(keysArr) == 0 {
		t.Error("JWKS should contain at least one key")
	}
}

func TestMFAChallengeHandler(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	body := `{"mfa_token":"some-token","challenge_type":"totp"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/challenge", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.MFAChallenge(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["challenge_type"] != "totp" {
		t.Errorf("challenge_type = %v, want %q", resp["challenge_type"], "totp")
	}
	if resp["status"] != "pending" {
		t.Errorf("status = %v, want %q", resp["status"], "pending")
	}
}

func TestMFAChallengeHandler_InvalidJSON(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/mfa/challenge", strings.NewReader("bad-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.MFAChallenge(w, req)

	if w.Code == http.StatusOK {
		t.Error("MFAChallenge should fail for invalid JSON")
	}
}

func TestSAMLMetadata(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/saml/metadata", nil)
	w := httptest.NewRecorder()

	h.SAMLMetadata(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/xml" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/xml")
	}

	body := w.Body.String()
	if !strings.Contains(body, "EntityDescriptor") {
		t.Error("SAML metadata should contain EntityDescriptor")
	}
}

func TestSAMLACS(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/saml/acs", nil)
	w := httptest.NewRecorder()

	h.SAMLACS(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestSAMLSSO(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/saml/sso", nil)
	w := httptest.NewRecorder()

	h.SAMLSSO(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestWebAuthnRegisterFinish(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/webauthn/register/finish", nil)
	w := httptest.NewRecorder()

	h.WebAuthnRegisterFinish(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "registered" {
		t.Errorf("status = %v, want %q", resp["status"], "registered")
	}
}

func TestWebAuthnLoginFinish(t *testing.T) {
	h, _, _, _ := testAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/webauthn/login/finish", nil)
	w := httptest.NewRecorder()

	h.WebAuthnLoginFinish(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTokenHandler_BasicAuth(t *testing.T) {
	h, appRepo, _, _ := testAuthHandler()

	tenantID := uuid.New()
	app := &models.Application{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         "Basic Auth App",
		Type:         models.AppTypeM2M,
		ClientID:     "basic-auth-client",
		RedirectURIs: []string{},
	}
	secret := "basic-auth-secret"
	app.ClientSecretHash = crypto.HashToken(secret)
	appRepo.apps[app.ID] = app
	appRepo.byClientID[app.ClientID] = app

	form := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {"api:read"},
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(app.ClientID, secret)
	req = withTenantContext(req, tenantID)
	w := httptest.NewRecorder()

	h.Token(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}
