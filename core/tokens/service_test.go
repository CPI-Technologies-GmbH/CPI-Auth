package tokens

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// --- Mock Refresh Token Repository ---

type mockRefreshTokenRepo struct {
	tokens map[uuid.UUID]*models.RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{
		tokens: make(map[uuid.UUID]*models.RefreshToken),
	}
}

func (m *mockRefreshTokenRepo) Create(_ context.Context, token *models.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	token.CreatedAt = time.Now().UTC()
	m.tokens[token.ID] = token
	return nil
}

func (m *mockRefreshTokenRepo) GetByTokenHash(_ context.Context, hash string) (*models.RefreshToken, error) {
	for _, t := range m.tokens {
		if t.TokenHash == hash {
			return t, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *mockRefreshTokenRepo) Revoke(_ context.Context, id uuid.UUID) error {
	if t, ok := m.tokens[id]; ok {
		t.Revoked = true
		return nil
	}
	return models.ErrNotFound
}

func (m *mockRefreshTokenRepo) RevokeByFamily(_ context.Context, family string) error {
	for _, t := range m.tokens {
		if t.Family == family {
			t.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) RevokeByUser(_ context.Context, userID uuid.UUID) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) RevokeByApplication(_ context.Context, appID uuid.UUID) error {
	for _, t := range m.tokens {
		if t.ApplicationID == appID {
			t.Revoked = true
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired(_ context.Context) error {
	now := time.Now().UTC()
	for id, t := range m.tokens {
		if now.After(t.ExpiresAt) {
			delete(m.tokens, id)
		}
	}
	return nil
}

// --- Test Helpers ---

func testTokenService() (*Service, *mockRefreshTokenRepo) {
	kp, _ := crypto.GenerateRSAKeyPair(2048)
	km := crypto.NewKeyManager(kp)
	repo := newMockRefreshTokenRepo()
	cfg := config.DefaultConfig()
	cfg.Security.Issuer = "https://test.cpi-auth.local"
	cfg.Security.AccessTokenLifetime = 15 * time.Minute
	cfg.Security.RefreshTokenLifetime = 7 * 24 * time.Hour
	cfg.Security.IDTokenLifetime = 1 * time.Hour
	logger := zap.NewNop()

	svc := &Service{
		keyManager:    km,
		refreshTokens: repo,
		redis:         nil, // Redis ops will be skipped
		cfg:           cfg,
		logger:        logger,
	}

	return svc, repo
}

// --- Tests ---

func TestNewTokenService(t *testing.T) {
	svc, _ := testTokenService()
	if svc == nil {
		t.Fatal("service should not be nil")
	}
}

func TestIssueTokenPair_AccessTokenOnly(t *testing.T) {
	svc, _ := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Email:         "user@test.com",
		Name:          "Test User",
		Scopes:        []string{"profile", "email"},
		Permissions:   []string{"read", "write"},
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", pair.TokenType, "Bearer")
	}
	if pair.ExpiresIn != int(svc.cfg.Security.AccessTokenLifetime.Seconds()) {
		t.Errorf("ExpiresIn = %d, want %d", pair.ExpiresIn, int(svc.cfg.Security.AccessTokenLifetime.Seconds()))
	}
	if pair.RefreshToken != "" {
		t.Error("RefreshToken should be empty without offline_access scope")
	}
	if pair.IDToken != "" {
		t.Error("IDToken should be empty without openid scope")
	}
}

func TestIssueTokenPair_WithRefreshToken(t *testing.T) {
	svc, repo := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Scopes:        []string{"offline_access"},
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	if pair.RefreshToken == "" {
		t.Error("RefreshToken should not be empty with offline_access scope")
	}

	// Verify refresh token was stored in repo
	if len(repo.tokens) != 1 {
		t.Errorf("expected 1 refresh token in repo, got %d", len(repo.tokens))
	}
}

func TestIssueTokenPair_WithIDToken(t *testing.T) {
	svc, _ := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Email:         "user@test.com",
		Name:          "Test User",
		Scopes:        []string{"openid", "email", "profile"},
		Nonce:         "test-nonce-123",
		EmailVerified: true,
		AvatarURL:     "https://example.com/avatar.jpg",
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	if pair.IDToken == "" {
		t.Error("IDToken should not be empty with openid scope")
	}

	// Parse the ID token to verify claims
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	idClaims := &IDTokenClaims{}
	_, _, err = parser.ParseUnverified(pair.IDToken, idClaims)
	if err != nil {
		t.Fatalf("failed to parse ID token: %v", err)
	}
	if idClaims.Email != "user@test.com" {
		t.Errorf("Email = %q, want %q", idClaims.Email, "user@test.com")
	}
	if !idClaims.EmailVerified {
		t.Error("EmailVerified should be true")
	}
	if idClaims.Name != "Test User" {
		t.Errorf("Name = %q, want %q", idClaims.Name, "Test User")
	}
	if idClaims.Nonce != "test-nonce-123" {
		t.Errorf("Nonce = %q, want %q", idClaims.Nonce, "test-nonce-123")
	}
}

func TestIssueTokenPair_WithPhoneScope(t *testing.T) {
	svc, _ := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Scopes:        []string{"openid", "phone"},
		Phone:         "+1234567890",
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	idClaims := &IDTokenClaims{}
	_, _, err = parser.ParseUnverified(pair.IDToken, idClaims)
	if err != nil {
		t.Fatalf("failed to parse ID token: %v", err)
	}
	if idClaims.Phone != "+1234567890" {
		t.Errorf("Phone = %q, want %q", idClaims.Phone, "+1234567890")
	}
}

func TestValidateAccessToken(t *testing.T) {
	svc, _ := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Email:         "validate@test.com",
		Scopes:        []string{"profile"},
		Permissions:   []string{"read"},
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	claims, err := svc.ValidateAccessToken(context.Background(), pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}

	if claims.Subject != input.UserID.String() {
		t.Errorf("Subject = %q, want %q", claims.Subject, input.UserID.String())
	}
	if claims.TenantID != input.TenantID.String() {
		t.Errorf("TenantID = %q, want %q", claims.TenantID, input.TenantID.String())
	}
	if claims.Email != "validate@test.com" {
		t.Errorf("Email = %q, want %q", claims.Email, "validate@test.com")
	}
	if claims.Issuer != "https://test.cpi-auth.local" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "https://test.cpi-auth.local")
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	svc, _ := testTokenService()

	_, err := svc.ValidateAccessToken(context.Background(), "not-a-valid-token")
	if err == nil {
		t.Error("ValidateAccessToken should fail for invalid token")
	}
}

func TestValidateAccessToken_WrongKey(t *testing.T) {
	svc, _ := testTokenService()

	// Create a token with a different key
	otherKP, _ := crypto.GenerateRSAKeyPair(2048)
	otherKM := crypto.NewKeyManager(otherKP)

	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test",
			Subject:   uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			ID:        uuid.New().String(),
		},
		TenantID: uuid.New().String(),
	}

	token := jwt.NewWithClaims(otherKM.GetSigningMethod(), claims)
	token.Header["kid"] = otherKP.ID
	tokenStr, _ := token.SignedString(otherKP.PrivateKey)

	_, err := svc.ValidateAccessToken(context.Background(), tokenStr)
	if err == nil {
		t.Error("ValidateAccessToken should fail for token signed with different key")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	svc, repo := testTokenService()

	appID := uuid.New()
	userID := uuid.New()

	// Issue a token pair with offline_access
	input := IssueTokenPairInput{
		UserID:        userID,
		TenantID:      uuid.New(),
		ApplicationID: appID,
		Scopes:        []string{"offline_access"},
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	// Refresh the token
	newPair, err := svc.RefreshAccessToken(context.Background(), pair.RefreshToken, appID)
	if err != nil {
		t.Fatalf("RefreshAccessToken returned error: %v", err)
	}

	if newPair.RefreshToken == "" {
		t.Error("new refresh token should be issued")
	}
	if newPair.RefreshToken == pair.RefreshToken {
		t.Error("new refresh token should be different from old one")
	}

	// Old token should be revoked
	oldHash := crypto.HashToken(pair.RefreshToken)
	oldRT, _ := repo.GetByTokenHash(context.Background(), oldHash)
	if oldRT != nil && !oldRT.Revoked {
		t.Error("old refresh token should be revoked after rotation")
	}
}

func TestRefreshAccessToken_ReuseDetection(t *testing.T) {
	svc, _ := testTokenService()

	appID := uuid.New()
	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: appID,
		Scopes:        []string{"offline_access"},
	}

	pair, _ := svc.IssueTokenPair(context.Background(), input)

	// First refresh - should succeed
	_, err := svc.RefreshAccessToken(context.Background(), pair.RefreshToken, appID)
	if err != nil {
		t.Fatalf("first RefreshAccessToken returned error: %v", err)
	}

	// Second refresh with same token - should detect reuse
	_, err = svc.RefreshAccessToken(context.Background(), pair.RefreshToken, appID)
	if err == nil {
		t.Error("second RefreshAccessToken should fail (reuse detection)")
	}
	if !models.IsAppError(err, models.ErrTokenRevoked) {
		t.Errorf("expected ErrTokenRevoked, got %v", err)
	}
}

func TestRefreshAccessToken_ExpiredToken(t *testing.T) {
	svc, repo := testTokenService()

	appID := uuid.New()
	refreshRaw := "expired-refresh-token"
	tokenHash := crypto.HashToken(refreshRaw)

	// Manually create an expired refresh token
	rtID := uuid.New()
	repo.tokens[rtID] = &models.RefreshToken{
		ID:            rtID,
		UserID:        uuid.New(),
		ApplicationID: appID,
		TokenHash:     tokenHash,
		Family:        uuid.New().String(),
		Revoked:       false,
		ExpiresAt:     time.Now().UTC().Add(-1 * time.Hour), // Expired
	}

	_, err := svc.RefreshAccessToken(context.Background(), refreshRaw, appID)
	if err == nil {
		t.Error("RefreshAccessToken should fail for expired token")
	}
	if !models.IsAppError(err, models.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestRefreshAccessToken_WrongApp(t *testing.T) {
	svc, _ := testTokenService()

	appID := uuid.New()
	wrongAppID := uuid.New()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: appID,
		Scopes:        []string{"offline_access"},
	}

	pair, _ := svc.IssueTokenPair(context.Background(), input)

	_, err := svc.RefreshAccessToken(context.Background(), pair.RefreshToken, wrongAppID)
	if err == nil {
		t.Error("RefreshAccessToken should fail with wrong app ID")
	}
	if !models.IsAppError(err, models.ErrInvalidClient) {
		t.Errorf("expected ErrInvalidClient, got %v", err)
	}
}

func TestRefreshAccessToken_InvalidToken(t *testing.T) {
	svc, _ := testTokenService()

	_, err := svc.RefreshAccessToken(context.Background(), "non-existent-token", uuid.New())
	if err == nil {
		t.Error("RefreshAccessToken should fail for non-existent token")
	}
}

func TestRevokeRefreshToken(t *testing.T) {
	svc, repo := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Scopes:        []string{"offline_access"},
	}

	pair, _ := svc.IssueTokenPair(context.Background(), input)

	err := svc.RevokeRefreshToken(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("RevokeRefreshToken returned error: %v", err)
	}

	// Check it's revoked
	hash := crypto.HashToken(pair.RefreshToken)
	rt, _ := repo.GetByTokenHash(context.Background(), hash)
	if rt != nil && !rt.Revoked {
		t.Error("refresh token should be revoked")
	}
}

func TestRevokeAllUserTokens(t *testing.T) {
	svc, repo := testTokenService()

	userID := uuid.New()

	// Create multiple tokens for the user
	for i := 0; i < 3; i++ {
		input := IssueTokenPairInput{
			UserID:        userID,
			TenantID:      uuid.New(),
			ApplicationID: uuid.New(),
			Scopes:        []string{"offline_access"},
		}
		_, _ = svc.IssueTokenPair(context.Background(), input)
	}

	err := svc.RevokeAllUserTokens(context.Background(), userID)
	if err != nil {
		t.Fatalf("RevokeAllUserTokens returned error: %v", err)
	}

	// All tokens should be revoked
	for _, rt := range repo.tokens {
		if rt.UserID == userID && !rt.Revoked {
			t.Error("all user's refresh tokens should be revoked")
		}
	}
}

func TestGetJWKS(t *testing.T) {
	svc, _ := testTokenService()

	jwks := svc.GetJWKS()
	if len(jwks.Keys) == 0 {
		t.Error("JWKS should contain at least one key")
	}
}

func TestGetKeyManager(t *testing.T) {
	svc, _ := testTokenService()

	km := svc.GetKeyManager()
	if km == nil {
		t.Error("GetKeyManager should not return nil")
	}
}

func TestContainsScope(t *testing.T) {
	tests := []struct {
		scopes []string
		target string
		want   bool
	}{
		{[]string{"openid", "profile"}, "openid", true},
		{[]string{"openid", "profile"}, "email", false},
		{[]string{}, "openid", false},
		{[]string{"offline_access"}, "offline_access", true},
	}

	for _, tt := range tests {
		got := containsScope(tt.scopes, tt.target)
		if got != tt.want {
			t.Errorf("containsScope(%v, %q) = %v, want %v", tt.scopes, tt.target, got, tt.want)
		}
	}
}

func TestIssueTokenPair_AccessTokenHasKid(t *testing.T) {
	svc, _ := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Scopes:        []string{"profile"},
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	// Parse without validation to check header
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(pair.AccessToken, &AccessTokenClaims{})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok || kid == "" {
		t.Error("access token should have kid header")
	}
}

func TestIssueTokenPair_ScopeInToken(t *testing.T) {
	svc, _ := testTokenService()

	input := IssueTokenPairInput{
		UserID:        uuid.New(),
		TenantID:      uuid.New(),
		ApplicationID: uuid.New(),
		Scopes:        []string{"openid", "profile", "email"},
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	if pair.Scope != "openid profile email" {
		t.Errorf("Scope = %q, want %q", pair.Scope, "openid profile email")
	}

	// Validate the scope is in the access token claims too
	claims, _ := svc.ValidateAccessToken(context.Background(), pair.AccessToken)
	if claims != nil {
		scopeParts := strings.Fields(claims.Scope)
		if len(scopeParts) != 3 {
			t.Errorf("expected 3 scope parts in claims, got %d", len(scopeParts))
		}
	}
}

func TestIssueTokenPair_PerAppTTLOverride(t *testing.T) {
	svc, _ := testTokenService()

	customTTL := 3600 // 1 hour
	input := IssueTokenPairInput{
		UserID:         uuid.New(),
		TenantID:       uuid.New(),
		ApplicationID:  uuid.New(),
		Email:          "ttl@test.com",
		Scopes:         []string{"profile"},
		AccessTokenTTL: &customTTL,
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	// ExpiresIn should be 3600 (1 hour), not the default 900 (15 min)
	if pair.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", pair.ExpiresIn)
	}

	// Parse token to verify actual expiry
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	claims := &AccessTokenClaims{}
	_, _, err = parser.ParseUnverified(pair.AccessToken, claims)
	if err != nil {
		t.Fatalf("failed to parse access token: %v", err)
	}
	diff := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if diff != time.Hour {
		t.Errorf("token expiry duration = %v, want 1h", diff)
	}
}

func TestIssueTokenPair_NilTTLUsesDefault(t *testing.T) {
	svc, _ := testTokenService()

	input := IssueTokenPairInput{
		UserID:         uuid.New(),
		TenantID:       uuid.New(),
		ApplicationID:  uuid.New(),
		Scopes:         []string{"profile"},
		AccessTokenTTL: nil, // should use default
	}

	pair, err := svc.IssueTokenPair(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}

	expected := int(svc.cfg.Security.AccessTokenLifetime.Seconds())
	if pair.ExpiresIn != expected {
		t.Errorf("ExpiresIn = %d, want %d (default)", pair.ExpiresIn, expected)
	}
}
