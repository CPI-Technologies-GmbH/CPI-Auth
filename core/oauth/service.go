package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/policy"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
)

// Service implements the OAuth 2.1 / OIDC provider.
type Service struct {
	apps        models.ApplicationRepository
	grants      models.OAuthGrantRepository
	users       models.UserRepository
	tokenSvc    *tokens.Service
	rbacSvc     *policy.RBACService
	appPermRepo models.ApplicationPermissionRepository
	cfg         *config.Config
	logger      *zap.Logger
}

// NewService creates a new OAuth service.
func NewService(
	apps models.ApplicationRepository,
	grants models.OAuthGrantRepository,
	users models.UserRepository,
	tokenSvc *tokens.Service,
	rbacSvc *policy.RBACService,
	appPermRepo models.ApplicationPermissionRepository,
	cfg *config.Config,
	logger *zap.Logger,
) *Service {
	return &Service{
		apps:        apps,
		grants:      grants,
		users:       users,
		tokenSvc:    tokenSvc,
		rbacSvc:     rbacSvc,
		appPermRepo: appPermRepo,
		cfg:         cfg,
		logger:      logger,
	}
}

// AuthorizeRequest holds authorization endpoint parameters.
type AuthorizeRequest struct {
	ClientID            string   `json:"client_id"`
	RedirectURI         string   `json:"redirect_uri"`
	ResponseType        string   `json:"response_type"`
	Scope               string   `json:"scope"`
	State               string   `json:"state"`
	CodeChallenge       string   `json:"code_challenge"`
	CodeChallengeMethod string   `json:"code_challenge_method"`
	Nonce               string   `json:"nonce"`
}

// AuthorizeResponse holds the authorization response.
type AuthorizeResponse struct {
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectURI string `json:"redirect_uri"`
}

// TokenRequest holds token endpoint parameters.
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CodeVerifier string `json:"code_verifier"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// Authorize handles the authorization endpoint logic.
func (s *Service) Authorize(ctx context.Context, userID uuid.UUID, req AuthorizeRequest) (*AuthorizeResponse, error) {
	// Validate client
	app, err := s.apps.GetByClientID(ctx, req.ClientID)
	if err != nil {
		if models.IsAppError(err, models.ErrNotFound) {
			return nil, models.ErrInvalidClient.WithMessage("Unknown client_id.")
		}
		return nil, err
	}

	// Validate response_type
	if req.ResponseType != "code" {
		return nil, models.ErrBadRequest.WithMessage("Only response_type=code is supported (OAuth 2.1).")
	}

	// Validate redirect_uri
	if !isValidRedirectURI(app.RedirectURIs, req.RedirectURI) {
		return nil, models.ErrBadRequest.WithMessage("Invalid redirect_uri.")
	}

	// PKCE is mandatory for OAuth 2.1
	if req.CodeChallenge == "" {
		return nil, models.ErrBadRequest.WithMessage("code_challenge is required (PKCE mandatory).")
	}
	if req.CodeChallengeMethod == "" {
		req.CodeChallengeMethod = "S256"
	}
	if req.CodeChallengeMethod != "S256" {
		return nil, models.ErrBadRequest.WithMessage("Only S256 code_challenge_method is supported.")
	}

	// Validate scopes
	scopes := parseScopes(req.Scope)
	for _, scope := range scopes {
		if !isValidScope(scope) {
			return nil, models.ErrInvalidScope.WithMessage(fmt.Sprintf("Invalid scope: %s", scope))
		}
	}

	// Generate authorization code
	code, err := crypto.GenerateRandomString(32)
	if err != nil {
		return nil, models.ErrInternal.Wrap(err)
	}

	grant := &models.OAuthGrant{
		UserID:              userID,
		ApplicationID:       app.ID,
		TenantID:            app.TenantID,
		Scopes:              scopes,
		Code:                code,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		RedirectURI:         req.RedirectURI,
		Nonce:               req.Nonce,
		ExpiresAt:           time.Now().UTC().Add(s.cfg.Security.AuthCodeLifetime),
	}

	if err := s.grants.Create(ctx, grant); err != nil {
		return nil, models.ErrInternal.Wrap(err)
	}

	return &AuthorizeResponse{
		Code:        code,
		State:       req.State,
		RedirectURI: req.RedirectURI,
	}, nil
}

// Exchange handles the token endpoint for different grant types.
func (s *Service) Exchange(ctx context.Context, req TokenRequest) (*tokens.TokenPair, error) {
	switch req.GrantType {
	case "authorization_code":
		return s.exchangeAuthorizationCode(ctx, req)
	case "refresh_token":
		return s.exchangeRefreshToken(ctx, req)
	case "client_credentials":
		return s.exchangeClientCredentials(ctx, req)
	default:
		return nil, models.ErrUnsupportedGrantType
	}
}

func (s *Service) exchangeAuthorizationCode(ctx context.Context, req TokenRequest) (*tokens.TokenPair, error) {
	if req.Code == "" {
		return nil, models.ErrInvalidGrant.WithMessage("Authorization code is required.")
	}

	grant, err := s.grants.GetByCode(ctx, req.Code)
	if err != nil {
		if models.IsAppError(err, models.ErrNotFound) {
			return nil, models.ErrInvalidGrant.WithMessage("Invalid or expired authorization code.")
		}
		return nil, err
	}

	// Single-use: delete immediately
	defer s.grants.Delete(ctx, grant.ID)

	// Check expiration
	if time.Now().UTC().After(grant.ExpiresAt) {
		return nil, models.ErrInvalidGrant.WithMessage("Authorization code has expired.")
	}

	// Validate redirect_uri
	if grant.RedirectURI != req.RedirectURI {
		return nil, models.ErrInvalidGrant.WithMessage("redirect_uri mismatch.")
	}

	// Validate PKCE
	if req.CodeVerifier == "" {
		return nil, models.ErrInvalidGrant.WithMessage("code_verifier is required (PKCE).")
	}
	if !verifyCodeChallenge(grant.CodeChallenge, grant.CodeChallengeMethod, req.CodeVerifier) {
		return nil, models.ErrInvalidGrant.WithMessage("Invalid code_verifier.")
	}

	// Validate client
	app, err := s.apps.GetByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, models.ErrInvalidClient
	}
	if app.ID != grant.ApplicationID {
		return nil, models.ErrInvalidClient
	}

	// For confidential clients, verify client_secret
	if app.Type == models.AppTypeWeb || app.Type == models.AppTypeM2M {
		if app.ClientSecretHash != "" && req.ClientSecret != "" {
			if !crypto.TimingSafeEqual(crypto.HashToken(req.ClientSecret), app.ClientSecretHash) {
				return nil, models.ErrInvalidClient
			}
		}
	}

	// Get user for claims
	user, err := s.users.GetByID(ctx, app.TenantID, grant.UserID)
	if err != nil {
		return nil, models.ErrInternal.Wrap(err)
	}

	// Load user permissions
	var userPerms []string
	if s.rbacSvc != nil {
		var permErr error
		userPerms, permErr = s.rbacSvc.GetEffectivePermissions(ctx, user.ID)
		if permErr != nil {
			s.logger.Warn("failed to load user permissions", zap.Error(permErr))
		}
		userPerms = s.filterByAppPermissions(ctx, app.ID, userPerms)
	}

	return s.tokenSvc.IssueTokenPair(ctx, tokens.IssueTokenPairInput{
		UserID:          user.ID,
		TenantID:        user.TenantID,
		ApplicationID:   app.ID,
		Email:           user.Email,
		Name:            user.Name,
		Scopes:          grant.Scopes,
		Nonce:           grant.Nonce,
		EmailVerified:   user.EmailVerified,
		Phone:           user.Phone,
		AvatarURL:       user.AvatarURL,
		Permissions:     userPerms,
		AccessTokenTTL:  app.AccessTokenTTL,
		RefreshTokenTTL: app.RefreshTokenTTL,
		IDTokenTTL:      app.IDTokenTTL,
	})
}

func (s *Service) exchangeRefreshToken(ctx context.Context, req TokenRequest) (*tokens.TokenPair, error) {
	if req.RefreshToken == "" {
		return nil, models.ErrInvalidGrant.WithMessage("refresh_token is required.")
	}

	app, err := s.apps.GetByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, models.ErrInvalidClient
	}

	pair, err := s.tokenSvc.RefreshAccessToken(ctx, req.RefreshToken, app.ID)
	if err != nil {
		return nil, err
	}

	// Look up user to issue a new access token
	user, err := s.users.GetByID(ctx, app.TenantID, pair.UserID)
	if err != nil {
		return nil, models.ErrInternal.Wrap(err)
	}

	// Load user permissions
	var userPerms []string
	if s.rbacSvc != nil {
		var permErr error
		userPerms, permErr = s.rbacSvc.GetEffectivePermissions(ctx, user.ID)
		if permErr != nil {
			s.logger.Warn("failed to load user permissions for refresh", zap.Error(permErr))
		}
		userPerms = s.filterByAppPermissions(ctx, app.ID, userPerms)
	}

	// Issue full token pair with access token
	fullPair, err := s.tokenSvc.IssueTokenPair(ctx, tokens.IssueTokenPairInput{
		UserID:          user.ID,
		TenantID:        user.TenantID,
		ApplicationID:   app.ID,
		Email:           user.Email,
		Name:            user.Name,
		Scopes:          []string{"openid", "profile", "email"},
		EmailVerified:   user.EmailVerified,
		Phone:           user.Phone,
		AvatarURL:       user.AvatarURL,
		Permissions:     userPerms,
		AccessTokenTTL:  app.AccessTokenTTL,
		RefreshTokenTTL: app.RefreshTokenTTL,
		IDTokenTTL:      app.IDTokenTTL,
	})
	if err != nil {
		return nil, err
	}

	// Use the already-rotated refresh token instead of creating another one
	fullPair.RefreshToken = pair.RefreshToken

	return fullPair, nil
}

func (s *Service) exchangeClientCredentials(ctx context.Context, req TokenRequest) (*tokens.TokenPair, error) {
	if req.ClientID == "" || req.ClientSecret == "" {
		return nil, models.ErrInvalidClient
	}

	app, err := s.apps.GetByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, models.ErrInvalidClient
	}

	if app.Type != models.AppTypeM2M {
		return nil, models.ErrInvalidClient.WithMessage("client_credentials grant is only for M2M applications.")
	}

	if !crypto.TimingSafeEqual(crypto.HashToken(req.ClientSecret), app.ClientSecretHash) {
		return nil, models.ErrInvalidClient
	}

	scopes := parseScopes(req.Scope)

	// Load app-scoped permissions for M2M token
	var appPerms []string
	if s.appPermRepo != nil {
		appPerms, _ = s.appPermRepo.GetPermissions(ctx, app.ID)
	}

	// Resolve per-app access token TTL
	accessTTL := s.cfg.Security.AccessTokenLifetime
	if app.AccessTokenTTL != nil && *app.AccessTokenTTL > 0 {
		accessTTL = time.Duration(*app.AccessTokenTTL) * time.Second
	}

	now := time.Now().UTC()
	accessClaims := tokens.AccessTokenClaims{}
	accessClaims.Issuer = s.cfg.Security.Issuer
	accessClaims.Subject = app.ClientID
	accessClaims.Audience = []string{app.ID.String()}
	accessClaims.IssuedAt = &jwt.NumericDate{Time: now}
	accessClaims.ExpiresAt = &jwt.NumericDate{Time: now.Add(accessTTL)}
	accessClaims.ID = uuid.New().String()
	accessClaims.TenantID = app.TenantID.String()
	accessClaims.Scope = strings.Join(scopes, " ")
	accessClaims.Permissions = appPerms

	km := s.tokenSvc.GetKeyManager()
	activeKey := km.ActiveKey()

	token := jwt.NewWithClaims(km.GetSigningMethod(), accessClaims)
	token.Header["kid"] = activeKey.ID

	tokenStr, err := token.SignedString(activeKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("signing m2m token: %w", err)
	}

	return &tokens.TokenPair{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   int(accessTTL.Seconds()),
		Scope:       strings.Join(scopes, " "),
	}, nil
}

// Revoke handles token revocation (RFC 7009).
func (s *Service) Revoke(ctx context.Context, tokenStr, tokenTypeHint string) error {
	switch tokenTypeHint {
	case "refresh_token":
		return s.tokenSvc.RevokeRefreshToken(ctx, tokenStr)
	case "access_token", "":
		// Try as access token first, then as refresh token
		err := s.tokenSvc.RevokeAccessToken(ctx, tokenStr)
		if err != nil {
			return s.tokenSvc.RevokeRefreshToken(ctx, tokenStr)
		}
		return nil
	default:
		return models.ErrBadRequest.WithMessage("Unsupported token_type_hint.")
	}
}

// GetUserinfo returns the user's profile based on access token claims.
func (s *Service) GetUserinfo(ctx context.Context, claims *tokens.AccessTokenClaims) (map[string]interface{}, error) {
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, models.ErrBadRequest
	}
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return nil, models.ErrBadRequest
	}

	user, err := s.users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"sub": user.ID.String(),
	}

	scopes := strings.Split(claims.Scope, " ")
	for _, scope := range scopes {
		switch scope {
		case "profile":
			info["name"] = user.Name
			info["picture"] = user.AvatarURL
			info["updated_at"] = user.UpdatedAt.Unix()
		case "email":
			info["email"] = user.Email
			info["email_verified"] = user.EmailVerified
		case "phone":
			if user.Phone != "" {
				info["phone_number"] = user.Phone
				info["phone_number_verified"] = user.PhoneVerified
			}
		}
	}

	return info, nil
}

// DiscoveryDocument returns the OIDC discovery metadata.
func (s *Service) DiscoveryDocument() map[string]interface{} {
	issuer := s.cfg.Security.Issuer
	return map[string]interface{}{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oauth/authorize",
		"token_endpoint":                        issuer + "/oauth/token",
		"userinfo_endpoint":                     issuer + "/oauth/userinfo",
		"jwks_uri":                              issuer + "/.well-known/jwks.json",
		"revocation_endpoint":                   issuer + "/oauth/revoke",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{s.cfg.Security.JWTSigningAlgorithm},
		"scopes_supported":                      []string{"openid", "profile", "email", "phone", "address", "offline_access"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic", "none"},
		"claims_supported":                      []string{"sub", "iss", "aud", "exp", "iat", "email", "email_verified", "name", "phone_number", "picture", "nonce"},
		"code_challenge_methods_supported":      []string{"S256"},
	}
}

// --- Helpers ---

func verifyCodeChallenge(challenge, method, verifier string) bool {
	if method != "S256" {
		return false
	}
	h := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return crypto.TimingSafeEqual(computed, challenge)
}

func isValidRedirectURI(allowed []string, uri string) bool {
	for _, u := range allowed {
		if u == uri {
			return true
		}
	}
	return false
}

func parseScopes(scope string) []string {
	if scope == "" {
		return []string{}
	}
	return strings.Fields(scope)
}

var validScopes = map[string]bool{
	"openid": true, "profile": true, "email": true,
	"phone": true, "address": true, "offline_access": true,
}

func isValidScope(scope string) bool {
	if validScopes[scope] {
		return true
	}
	// Allow custom scopes (e.g., api:read, etc.)
	return strings.Contains(scope, ":")
}

// filterByAppPermissions intersects user permissions with an app's allowed permissions.
// If the app has no permissions configured (empty whitelist), all user permissions pass through.
func (s *Service) filterByAppPermissions(ctx context.Context, appID uuid.UUID, userPerms []string) []string {
	if s.appPermRepo == nil {
		return userPerms
	}
	appPerms, err := s.appPermRepo.GetPermissions(ctx, appID)
	if err != nil {
		s.logger.Warn("failed to load app permissions", zap.Error(err))
		return userPerms
	}
	if len(appPerms) == 0 {
		return userPerms
	}
	allowed := make(map[string]bool, len(appPerms))
	for _, p := range appPerms {
		allowed[p] = true
	}
	var filtered []string
	for _, p := range userPerms {
		if allowed[p] {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
