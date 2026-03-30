package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
)

// --- Mock Repositories ---

type mockAppRepo struct {
	apps         map[uuid.UUID]*models.Application
	byClientID   map[string]*models.Application
}

func newMockAppRepo() *mockAppRepo {
	return &mockAppRepo{
		apps:       make(map[uuid.UUID]*models.Application),
		byClientID: make(map[string]*models.Application),
	}
}

func (m *mockAppRepo) Create(_ context.Context, app *models.Application) error {
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	m.apps[app.ID] = app
	m.byClientID[app.ClientID] = app
	return nil
}

func (m *mockAppRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.Application, error) {
	app, ok := m.apps[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return app, nil
}

func (m *mockAppRepo) GetByClientID(_ context.Context, clientID string) (*models.Application, error) {
	app, ok := m.byClientID[clientID]
	if !ok {
		return nil, models.ErrNotFound
	}
	return app, nil
}

func (m *mockAppRepo) Update(_ context.Context, app *models.Application) error {
	m.apps[app.ID] = app
	m.byClientID[app.ClientID] = app
	return nil
}

func (m *mockAppRepo) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	if app, ok := m.apps[id]; ok {
		delete(m.byClientID, app.ClientID)
		delete(m.apps, id)
	}
	return nil
}

func (m *mockAppRepo) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Application], error) {
	var result []models.Application
	for _, app := range m.apps {
		if app.TenantID == tenantID {
			result = append(result, *app)
		}
	}
	return &models.PaginatedResult[models.Application]{Data: result, Total: int64(len(result))}, nil
}

type mockGrantRepo struct {
	grants   map[uuid.UUID]*models.OAuthGrant
	byCode   map[string]*models.OAuthGrant
}

func newMockGrantRepo() *mockGrantRepo {
	return &mockGrantRepo{
		grants: make(map[uuid.UUID]*models.OAuthGrant),
		byCode: make(map[string]*models.OAuthGrant),
	}
}

func (m *mockGrantRepo) Create(_ context.Context, grant *models.OAuthGrant) error {
	if grant.ID == uuid.Nil {
		grant.ID = uuid.New()
	}
	grant.CreatedAt = time.Now().UTC()
	m.grants[grant.ID] = grant
	m.byCode[grant.Code] = grant
	return nil
}

func (m *mockGrantRepo) GetByCode(_ context.Context, code string) (*models.OAuthGrant, error) {
	grant, ok := m.byCode[code]
	if !ok {
		return nil, models.ErrNotFound
	}
	return grant, nil
}

func (m *mockGrantRepo) Delete(_ context.Context, id uuid.UUID) error {
	if grant, ok := m.grants[id]; ok {
		delete(m.byCode, grant.Code)
		delete(m.grants, id)
	}
	return nil
}

func (m *mockGrantRepo) DeleteExpired(_ context.Context) error {
	now := time.Now().UTC()
	for id, grant := range m.grants {
		if grant.ExpiresAt.Before(now) {
			delete(m.byCode, grant.Code)
			delete(m.grants, id)
		}
	}
	return nil
}

type mockUserRepoOAuth struct {
	users map[uuid.UUID]*models.User
}

func newMockUserRepoOAuth() *mockUserRepoOAuth {
	return &mockUserRepoOAuth{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *mockUserRepoOAuth) Create(_ context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoOAuth) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepoOAuth) GetByEmail(_ context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email && u.TenantID == tenantID {
			return u, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *mockUserRepoOAuth) Update(_ context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoOAuth) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepoOAuth) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, search string) (*models.PaginatedResult[models.User], error) {
	return &models.PaginatedResult[models.User]{}, nil
}

func (m *mockUserRepoOAuth) Block(_ context.Context, tenantID, id uuid.UUID) error { return nil }
func (m *mockUserRepoOAuth) Unblock(_ context.Context, tenantID, id uuid.UUID) error { return nil }
func (m *mockUserRepoOAuth) CountByTenant(_ context.Context, tenantID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockUserRepoOAuth) GetPasswordHistory(_ context.Context, userID uuid.UUID, limit int) ([]models.PasswordHistory, error) {
	return nil, nil
}
func (m *mockUserRepoOAuth) AddPasswordHistory(_ context.Context, entry *models.PasswordHistory) error {
	return nil
}

// --- Mock Refresh Token Repo for token service ---

type mockRefreshTokenRepoOAuth struct {
	tokens     map[uuid.UUID]*models.RefreshToken
	byHash     map[string]*models.RefreshToken
}

func newMockRefreshTokenRepoOAuth() *mockRefreshTokenRepoOAuth {
	return &mockRefreshTokenRepoOAuth{
		tokens: make(map[uuid.UUID]*models.RefreshToken),
		byHash: make(map[string]*models.RefreshToken),
	}
}

func (m *mockRefreshTokenRepoOAuth) Create(_ context.Context, token *models.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	token.CreatedAt = time.Now().UTC()
	m.tokens[token.ID] = token
	m.byHash[token.TokenHash] = token
	return nil
}

func (m *mockRefreshTokenRepoOAuth) GetByTokenHash(_ context.Context, hash string) (*models.RefreshToken, error) {
	rt, ok := m.byHash[hash]
	if !ok {
		return nil, models.ErrNotFound
	}
	return rt, nil
}

func (m *mockRefreshTokenRepoOAuth) Revoke(_ context.Context, id uuid.UUID) error {
	if rt, ok := m.tokens[id]; ok {
		rt.Revoked = true
	}
	return nil
}

func (m *mockRefreshTokenRepoOAuth) RevokeByFamily(_ context.Context, family string) error {
	for _, rt := range m.tokens {
		if rt.Family == family {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepoOAuth) RevokeByUser(_ context.Context, userID uuid.UUID) error {
	for _, rt := range m.tokens {
		if rt.UserID == userID {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepoOAuth) RevokeByApplication(_ context.Context, appID uuid.UUID) error {
	for _, rt := range m.tokens {
		if rt.ApplicationID == appID {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepoOAuth) DeleteExpired(_ context.Context) error {
	return nil
}

// --- Test helpers ---

func testOAuthConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Security.AuthCodeLifetime = 10 * time.Minute
	cfg.Security.AccessTokenLifetime = 15 * time.Minute
	cfg.Security.RefreshTokenLifetime = 7 * 24 * time.Hour
	cfg.Security.IDTokenLifetime = 1 * time.Hour
	cfg.Security.Issuer = "https://auth.example.com"
	return cfg
}

func testOAuthService() (*Service, *mockAppRepo, *mockGrantRepo, *mockUserRepoOAuth, *tokens.Service) {
	cfg := testOAuthConfig()
	appRepo := newMockAppRepo()
	grantRepo := newMockGrantRepo()
	userRepo := newMockUserRepoOAuth()
	refreshRepo := newMockRefreshTokenRepoOAuth()
	logger := zap.NewNop()

	keyPair, _ := crypto.GenerateRSAKeyPair(2048)
	km := crypto.NewKeyManager(keyPair)
	tokenSvc := tokens.NewService(km, refreshRepo, nil, cfg, logger)

	svc := NewService(appRepo, grantRepo, userRepo, tokenSvc, nil, nil, cfg, logger)
	return svc, appRepo, grantRepo, userRepo, tokenSvc
}

func createTestApp(appRepo *mockAppRepo, tenantID uuid.UUID, appType models.ApplicationType) *models.Application {
	app := &models.Application{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         "Test App",
		Type:         appType,
		ClientID:     "test-client-id-" + uuid.New().String()[:8],
		RedirectURIs: []string{"https://app.example.com/callback"},
	}
	appRepo.apps[app.ID] = app
	appRepo.byClientID[app.ClientID] = app
	return app
}

func createTestUser(userRepo *mockUserRepoOAuth, tenantID uuid.UUID) *models.User {
	user := &models.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
		Phone:         "+15551234567",
		AvatarURL:     "https://example.com/avatar.png",
		Status:        models.StatusActive,
		UpdatedAt:     time.Now().UTC(),
	}
	userRepo.users[user.ID] = user
	return user
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// --- Tests ---

func TestNewService(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestAuthorize_Success(t *testing.T) {
	svc, appRepo, grantRepo, _, _ := testOAuthService()
	tenantID := uuid.New()
	userID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := generateCodeChallenge(verifier)

	req := AuthorizeRequest{
		ClientID:            app.ClientID,
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		Scope:               "openid profile email",
		State:               "random-state",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		Nonce:               "nonce-123",
	}

	resp, err := svc.Authorize(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("Authorize returned error: %v", err)
	}

	if resp.Code == "" {
		t.Error("Code should not be empty")
	}
	if resp.State != "random-state" {
		t.Errorf("State = %q, want %q", resp.State, "random-state")
	}
	if resp.RedirectURI != "https://app.example.com/callback" {
		t.Errorf("RedirectURI = %q, want %q", resp.RedirectURI, "https://app.example.com/callback")
	}

	// Verify grant was stored
	if len(grantRepo.grants) != 1 {
		t.Errorf("expected 1 grant in repo, got %d", len(grantRepo.grants))
	}
}

func TestAuthorize_UnknownClient(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	req := AuthorizeRequest{
		ClientID:      "unknown-client",
		RedirectURI:   "https://example.com/cb",
		ResponseType:  "code",
		CodeChallenge: "test-challenge",
	}

	_, err := svc.Authorize(context.Background(), uuid.New(), req)
	if err == nil {
		t.Fatal("Authorize should fail for unknown client")
	}
	if !models.IsAppError(err, models.ErrInvalidClient) {
		t.Errorf("expected ErrInvalidClient, got %v", err)
	}
}

func TestAuthorize_InvalidResponseType(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	req := AuthorizeRequest{
		ClientID:      app.ClientID,
		RedirectURI:   "https://app.example.com/callback",
		ResponseType:  "token", // OAuth 2.1 only supports "code"
		CodeChallenge: "test-challenge",
	}

	_, err := svc.Authorize(context.Background(), uuid.New(), req)
	if err == nil {
		t.Fatal("Authorize should fail for response_type=token")
	}
	if !models.IsAppError(err, models.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

func TestAuthorize_InvalidRedirectURI(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	req := AuthorizeRequest{
		ClientID:      app.ClientID,
		RedirectURI:   "https://evil.com/callback",
		ResponseType:  "code",
		CodeChallenge: "test-challenge",
	}

	_, err := svc.Authorize(context.Background(), uuid.New(), req)
	if err == nil {
		t.Fatal("Authorize should fail for invalid redirect_uri")
	}
}

func TestAuthorize_MissingPKCE(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	req := AuthorizeRequest{
		ClientID:     app.ClientID,
		RedirectURI:  "https://app.example.com/callback",
		ResponseType: "code",
		Scope:        "openid",
		// Missing CodeChallenge
	}

	_, err := svc.Authorize(context.Background(), uuid.New(), req)
	if err == nil {
		t.Fatal("Authorize should fail without PKCE code_challenge")
	}
}

func TestAuthorize_UnsupportedCodeChallengeMethod(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	req := AuthorizeRequest{
		ClientID:            app.ClientID,
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "plain", // Only S256 is supported
	}

	_, err := svc.Authorize(context.Background(), uuid.New(), req)
	if err == nil {
		t.Fatal("Authorize should fail for code_challenge_method=plain")
	}
}

func TestAuthorize_InvalidScope(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	req := AuthorizeRequest{
		ClientID:            app.ClientID,
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		Scope:               "openid invalid_scope",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
	}

	_, err := svc.Authorize(context.Background(), uuid.New(), req)
	if err == nil {
		t.Fatal("Authorize should fail for invalid scope")
	}
	if !models.IsAppError(err, models.ErrInvalidScope) {
		t.Errorf("expected ErrInvalidScope, got %v", err)
	}
}

func TestAuthorize_CustomScopeAllowed(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	verifier := "test-verifier-12345678901234567890"
	challenge := generateCodeChallenge(verifier)

	req := AuthorizeRequest{
		ClientID:            app.ClientID,
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		Scope:               "openid api:read",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}

	resp, err := svc.Authorize(context.Background(), uuid.New(), req)
	if err != nil {
		t.Fatalf("Custom scopes with ':' should be allowed, got error: %v", err)
	}
	if resp.Code == "" {
		t.Error("Code should not be empty")
	}
}

func TestExchange_AuthorizationCode_Success(t *testing.T) {
	svc, appRepo, grantRepo, userRepo, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)
	user := createTestUser(userRepo, tenantID)

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := generateCodeChallenge(verifier)

	// Create a grant manually
	code := "test-auth-code-12345"
	grant := &models.OAuthGrant{
		ID:                  uuid.New(),
		UserID:              user.ID,
		ApplicationID:       app.ID,
		Scopes:              []string{"openid", "profile"},
		Code:                code,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		RedirectURI:         "https://app.example.com/callback",
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	grantRepo.grants[grant.ID] = grant
	grantRepo.byCode[code] = grant

	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     app.ClientID,
		CodeVerifier: verifier,
	}

	pair, err := svc.Exchange(context.Background(), req)
	if err != nil {
		t.Fatalf("Exchange returned error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", pair.TokenType, "Bearer")
	}
}

func TestExchange_AuthorizationCode_InvalidCode(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	createTestApp(appRepo, tenantID, models.AppTypeSPA)

	req := TokenRequest{
		GrantType: "authorization_code",
		Code:      "nonexistent-code",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for invalid code")
	}
}

func TestExchange_AuthorizationCode_EmptyCode(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	req := TokenRequest{
		GrantType: "authorization_code",
		Code:      "",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for empty code")
	}
}

func TestExchange_AuthorizationCode_ExpiredCode(t *testing.T) {
	svc, appRepo, grantRepo, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	code := "expired-code"
	grant := &models.OAuthGrant{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		ApplicationID:       app.ID,
		Code:                code,
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		RedirectURI:         "https://app.example.com/callback",
		ExpiresAt:           time.Now().UTC().Add(-1 * time.Hour), // Expired
	}
	grantRepo.grants[grant.ID] = grant
	grantRepo.byCode[code] = grant

	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     app.ClientID,
		CodeVerifier: "some-verifier",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for expired code")
	}
}

func TestExchange_AuthorizationCode_RedirectURIMismatch(t *testing.T) {
	svc, appRepo, grantRepo, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	code := "redirect-mismatch-code"
	grant := &models.OAuthGrant{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		ApplicationID:       app.ID,
		Code:                code,
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		RedirectURI:         "https://app.example.com/callback",
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	grantRepo.grants[grant.ID] = grant
	grantRepo.byCode[code] = grant

	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "https://different.example.com/callback",
		ClientID:     app.ClientID,
		CodeVerifier: "some-verifier",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for redirect_uri mismatch")
	}
}

func TestExchange_AuthorizationCode_MissingCodeVerifier(t *testing.T) {
	svc, appRepo, grantRepo, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)

	code := "no-verifier-code"
	grant := &models.OAuthGrant{
		ID:                  uuid.New(),
		UserID:              uuid.New(),
		ApplicationID:       app.ID,
		Code:                code,
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		RedirectURI:         "https://app.example.com/callback",
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	grantRepo.grants[grant.ID] = grant
	grantRepo.byCode[code] = grant

	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     app.ClientID,
		CodeVerifier: "", // Missing
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail without code_verifier")
	}
}

func TestExchange_AuthorizationCode_InvalidCodeVerifier(t *testing.T) {
	svc, appRepo, grantRepo, userRepo, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)
	user := createTestUser(userRepo, tenantID)

	verifier := "correct-verifier-12345678901234567890"
	challenge := generateCodeChallenge(verifier)

	code := "invalid-verifier-code"
	grant := &models.OAuthGrant{
		ID:                  uuid.New(),
		UserID:              user.ID,
		ApplicationID:       app.ID,
		Code:                code,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		RedirectURI:         "https://app.example.com/callback",
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	grantRepo.grants[grant.ID] = grant
	grantRepo.byCode[code] = grant

	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     app.ClientID,
		CodeVerifier: "wrong-verifier-completely-different",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for invalid code_verifier")
	}
}

func TestExchange_AuthorizationCode_ClientMismatch(t *testing.T) {
	svc, appRepo, grantRepo, userRepo, _ := testOAuthService()
	tenantID := uuid.New()
	app1 := createTestApp(appRepo, tenantID, models.AppTypeSPA)
	app2 := createTestApp(appRepo, tenantID, models.AppTypeSPA)
	user := createTestUser(userRepo, tenantID)

	verifier := "verifier-for-mismatch-test-123456789"
	challenge := generateCodeChallenge(verifier)

	code := "client-mismatch-code"
	grant := &models.OAuthGrant{
		ID:                  uuid.New(),
		UserID:              user.ID,
		ApplicationID:       app1.ID, // Granted for app1
		Code:                code,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		RedirectURI:         "https://app.example.com/callback",
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	grantRepo.grants[grant.ID] = grant
	grantRepo.byCode[code] = grant

	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     app2.ClientID, // Using app2 client ID
		CodeVerifier: verifier,
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for client mismatch")
	}
}

func TestExchange_AuthorizationCode_SingleUse(t *testing.T) {
	svc, appRepo, grantRepo, userRepo, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA)
	user := createTestUser(userRepo, tenantID)

	verifier := "single-use-verifier-12345678901234567"
	challenge := generateCodeChallenge(verifier)

	code := "single-use-code"
	grant := &models.OAuthGrant{
		ID:                  uuid.New(),
		UserID:              user.ID,
		ApplicationID:       app.ID,
		Scopes:              []string{"openid"},
		Code:                code,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		RedirectURI:         "https://app.example.com/callback",
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	grantRepo.grants[grant.ID] = grant
	grantRepo.byCode[code] = grant

	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     app.ClientID,
		CodeVerifier: verifier,
	}

	// First exchange should succeed
	_, err := svc.Exchange(context.Background(), req)
	if err != nil {
		t.Fatalf("First Exchange returned error: %v", err)
	}

	// Second exchange should fail (code was deleted)
	_, err = svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Second Exchange should fail (single-use code)")
	}
}

func TestExchange_RefreshToken_EmptyToken(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	req := TokenRequest{
		GrantType:    "refresh_token",
		RefreshToken: "",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for empty refresh token")
	}
}

func TestExchange_ClientCredentials_Success(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeM2M)

	secret := "test-client-secret"
	app.ClientSecretHash = crypto.HashToken(secret)
	appRepo.apps[app.ID] = app
	appRepo.byClientID[app.ClientID] = app

	req := TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     app.ClientID,
		ClientSecret: secret,
		Scope:        "api:read api:write",
	}

	pair, err := svc.Exchange(context.Background(), req)
	if err != nil {
		t.Fatalf("Exchange returned error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", pair.TokenType, "Bearer")
	}
	// Client credentials should not include a refresh token
	if pair.RefreshToken != "" {
		t.Error("client_credentials should not include a refresh token")
	}
}

func TestExchange_ClientCredentials_NonM2MApp(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeSPA) // SPA, not M2M

	secret := "test-secret"
	app.ClientSecretHash = crypto.HashToken(secret)
	appRepo.apps[app.ID] = app
	appRepo.byClientID[app.ClientID] = app

	req := TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     app.ClientID,
		ClientSecret: secret,
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for non-M2M app with client_credentials")
	}
}

func TestExchange_ClientCredentials_WrongSecret(t *testing.T) {
	svc, appRepo, _, _, _ := testOAuthService()
	tenantID := uuid.New()
	app := createTestApp(appRepo, tenantID, models.AppTypeM2M)

	app.ClientSecretHash = crypto.HashToken("correct-secret")
	appRepo.apps[app.ID] = app
	appRepo.byClientID[app.ClientID] = app

	req := TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     app.ClientID,
		ClientSecret: "wrong-secret",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for wrong client secret")
	}
}

func TestExchange_ClientCredentials_MissingCredentials(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	req := TokenRequest{
		GrantType: "client_credentials",
		// Missing ClientID and ClientSecret
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for missing credentials")
	}
}

func TestExchange_UnsupportedGrantType(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	req := TokenRequest{
		GrantType: "implicit",
	}

	_, err := svc.Exchange(context.Background(), req)
	if err == nil {
		t.Fatal("Exchange should fail for unsupported grant type")
	}
	if !models.IsAppError(err, models.ErrUnsupportedGrantType) {
		t.Errorf("expected ErrUnsupportedGrantType, got %v", err)
	}
}

func TestGetUserinfo_ProfileScope(t *testing.T) {
	svc, _, _, userRepo, _ := testOAuthService()
	tenantID := uuid.New()
	user := createTestUser(userRepo, tenantID)

	claims := &tokens.AccessTokenClaims{}
	claims.Subject = user.ID.String()
	claims.TenantID = tenantID.String()
	claims.Scope = "openid profile email phone"

	info, err := svc.GetUserinfo(context.Background(), claims)
	if err != nil {
		t.Fatalf("GetUserinfo returned error: %v", err)
	}

	if info["sub"] != user.ID.String() {
		t.Errorf("sub = %v, want %v", info["sub"], user.ID.String())
	}
	if info["name"] != user.Name {
		t.Errorf("name = %v, want %v", info["name"], user.Name)
	}
	if info["email"] != user.Email {
		t.Errorf("email = %v, want %v", info["email"], user.Email)
	}
	if info["email_verified"] != true {
		t.Errorf("email_verified = %v, want true", info["email_verified"])
	}
	if info["phone_number"] != user.Phone {
		t.Errorf("phone_number = %v, want %v", info["phone_number"], user.Phone)
	}
}

func TestGetUserinfo_EmailScopeOnly(t *testing.T) {
	svc, _, _, userRepo, _ := testOAuthService()
	tenantID := uuid.New()
	user := createTestUser(userRepo, tenantID)

	claims := &tokens.AccessTokenClaims{}
	claims.Subject = user.ID.String()
	claims.TenantID = tenantID.String()
	claims.Scope = "openid email"

	info, err := svc.GetUserinfo(context.Background(), claims)
	if err != nil {
		t.Fatalf("GetUserinfo returned error: %v", err)
	}

	if info["email"] != user.Email {
		t.Errorf("email = %v, want %v", info["email"], user.Email)
	}
	// Profile fields should NOT be present
	if _, ok := info["name"]; ok {
		t.Error("name should not be present without profile scope")
	}
}

func TestGetUserinfo_InvalidSubject(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	claims := &tokens.AccessTokenClaims{}
	claims.Subject = "not-a-uuid"
	claims.TenantID = uuid.New().String()

	_, err := svc.GetUserinfo(context.Background(), claims)
	if err == nil {
		t.Fatal("GetUserinfo should fail for invalid subject UUID")
	}
}

func TestGetUserinfo_InvalidTenantID(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	claims := &tokens.AccessTokenClaims{}
	claims.Subject = uuid.New().String()
	claims.TenantID = "not-a-uuid"

	_, err := svc.GetUserinfo(context.Background(), claims)
	if err == nil {
		t.Fatal("GetUserinfo should fail for invalid tenant UUID")
	}
}

func TestGetUserinfo_UserNotFound(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	claims := &tokens.AccessTokenClaims{}
	claims.Subject = uuid.New().String()
	claims.TenantID = uuid.New().String()
	claims.Scope = "openid profile"

	_, err := svc.GetUserinfo(context.Background(), claims)
	if err == nil {
		t.Fatal("GetUserinfo should fail for non-existent user")
	}
}

func TestDiscoveryDocument(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	doc := svc.DiscoveryDocument()

	tests := []struct {
		key  string
		want interface{}
	}{
		{"issuer", "https://auth.example.com"},
		{"authorization_endpoint", "https://auth.example.com/oauth/authorize"},
		{"token_endpoint", "https://auth.example.com/oauth/token"},
		{"userinfo_endpoint", "https://auth.example.com/oauth/userinfo"},
		{"jwks_uri", "https://auth.example.com/.well-known/jwks.json"},
		{"revocation_endpoint", "https://auth.example.com/oauth/revoke"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := doc[tt.key]
			if !ok {
				t.Errorf("missing key %q in discovery document", tt.key)
				return
			}
			if val != tt.want {
				t.Errorf("%s = %v, want %v", tt.key, val, tt.want)
			}
		})
	}

	// Check arrays
	responseTypes := doc["response_types_supported"].([]string)
	if len(responseTypes) != 1 || responseTypes[0] != "code" {
		t.Errorf("response_types_supported = %v, want [code]", responseTypes)
	}

	grantTypes := doc["grant_types_supported"].([]string)
	if len(grantTypes) != 3 {
		t.Errorf("expected 3 grant types, got %d", len(grantTypes))
	}

	codeMethods := doc["code_challenge_methods_supported"].([]string)
	if len(codeMethods) != 1 || codeMethods[0] != "S256" {
		t.Errorf("code_challenge_methods_supported = %v, want [S256]", codeMethods)
	}
}

func TestRevoke_RefreshTokenHint(t *testing.T) {
	// Revoke with hint "refresh_token" delegates to tokenSvc.RevokeRefreshToken.
	// Since there's no matching token, it should return an error from the repo.
	svc, _, _, _, _ := testOAuthService()

	err := svc.Revoke(context.Background(), "nonexistent-token", "refresh_token")
	if err == nil {
		t.Fatal("Revoke should return error for non-existent refresh token")
	}
}

func TestRevoke_UnsupportedHint(t *testing.T) {
	svc, _, _, _, _ := testOAuthService()

	err := svc.Revoke(context.Background(), "some-token", "unsupported_hint")
	if err == nil {
		t.Fatal("Revoke should fail for unsupported token_type_hint")
	}
}

// --- Helper function tests ---

func TestVerifyCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	if !verifyCodeChallenge(challenge, "S256", verifier) {
		t.Error("verifyCodeChallenge should return true for correct verifier")
	}

	if verifyCodeChallenge(challenge, "S256", "wrong-verifier") {
		t.Error("verifyCodeChallenge should return false for wrong verifier")
	}

	if verifyCodeChallenge(challenge, "plain", verifier) {
		t.Error("verifyCodeChallenge should return false for non-S256 method")
	}
}

func TestIsValidRedirectURI(t *testing.T) {
	allowed := []string{"https://app.example.com/callback", "https://app.example.com/other"}

	if !isValidRedirectURI(allowed, "https://app.example.com/callback") {
		t.Error("should return true for allowed URI")
	}

	if isValidRedirectURI(allowed, "https://evil.com/callback") {
		t.Error("should return false for non-allowed URI")
	}

	if isValidRedirectURI([]string{}, "https://any.com") {
		t.Error("should return false for empty allowed list")
	}
}

func TestParseScopes(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"openid", 1},
		{"openid profile email", 3},
		{"openid  profile", 2}, // extra space handled by Fields
	}

	for _, tt := range tests {
		scopes := parseScopes(tt.input)
		if len(scopes) != tt.want {
			t.Errorf("parseScopes(%q) = %d scopes, want %d", tt.input, len(scopes), tt.want)
		}
	}
}

func TestIsValidScope(t *testing.T) {
	validOnes := []string{"openid", "profile", "email", "phone", "address", "offline_access", "api:read"}
	for _, scope := range validOnes {
		if !isValidScope(scope) {
			t.Errorf("isValidScope(%q) should be true", scope)
		}
	}

	invalidOnes := []string{"invalid_scope", "random"}
	for _, scope := range invalidOnes {
		if isValidScope(scope) {
			t.Errorf("isValidScope(%q) should be false", scope)
		}
	}
}
