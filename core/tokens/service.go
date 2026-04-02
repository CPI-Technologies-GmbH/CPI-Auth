package tokens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Service handles JWT access tokens, refresh tokens, and ID tokens.
type Service struct {
	keyManager    *crypto.KeyManager
	refreshTokens models.RefreshTokenRepository
	redis         *redis.Client
	cfg           *config.Config
	logger        *zap.Logger
}

// NewService creates a new token service.
func NewService(km *crypto.KeyManager, rt models.RefreshTokenRepository, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		keyManager:    km,
		refreshTokens: rt,
		redis:         rdb,
		cfg:           cfg,
		logger:        logger,
	}
}

// ActorClaim represents the acting party in token delegation/impersonation (RFC 8693 Section 4.1).
type ActorClaim struct {
	Sub string `json:"sub"`
}

// AccessTokenClaims holds the claims for an access token.
type AccessTokenClaims struct {
	jwt.RegisteredClaims
	TenantID    string      `json:"tenant_id"`
	Email       string      `json:"email,omitempty"`
	Name        string      `json:"name,omitempty"`
	Scope       string      `json:"scope,omitempty"`
	Permissions []string    `json:"permissions,omitempty"`
	Act         *ActorClaim `json:"act,omitempty"`
}

// IDTokenClaims holds the claims for an OIDC ID token.
type IDTokenClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	Phone         string `json:"phone_number,omitempty"`
	Picture       string `json:"picture,omitempty"`
	Nonce         string `json:"nonce,omitempty"`
	AuthTime      int64  `json:"auth_time,omitempty"`
	AtHash        string `json:"at_hash,omitempty"`
}

// TokenPair holds an access token + refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	UserID       uuid.UUID `json:"-"` // internal use only, not serialized
}

// IssueTokenPairInput holds the parameters for issuing a token pair.
type IssueTokenPairInput struct {
	UserID        uuid.UUID
	TenantID      uuid.UUID
	ApplicationID uuid.UUID
	Email         string
	Name          string
	Scopes        []string
	Permissions   []string
	Nonce         string
	EmailVerified bool
	Phone         string
	AvatarURL     string
	// Per-app TTL overrides (seconds). Nil means use global config.
	AccessTokenTTL  *int
	RefreshTokenTTL *int
	IDTokenTTL      *int
	// ActorID is set when the token is issued on behalf of another user (impersonation).
	ActorID *uuid.UUID
	// Issuer overrides the global issuer (for per-tenant domains).
	Issuer string
}

// IssueTokenPair creates a new access token, refresh token, and optionally an ID token.
func (s *Service) IssueTokenPair(ctx context.Context, input IssueTokenPairInput) (*TokenPair, error) {
	now := time.Now().UTC()
	activeKey := s.keyManager.ActiveKey()

	// Resolve TTLs: per-app override > global config
	accessTTL := s.cfg.Security.AccessTokenLifetime
	if input.AccessTokenTTL != nil && *input.AccessTokenTTL > 0 {
		accessTTL = time.Duration(*input.AccessTokenTTL) * time.Second
	}
	refreshTTL := s.cfg.Security.RefreshTokenLifetime
	if input.RefreshTokenTTL != nil && *input.RefreshTokenTTL > 0 {
		refreshTTL = time.Duration(*input.RefreshTokenTTL) * time.Second
	}
	idTTL := s.cfg.Security.IDTokenLifetime
	if input.IDTokenTTL != nil && *input.IDTokenTTL > 0 {
		idTTL = time.Duration(*input.IDTokenTTL) * time.Second
	}

	// Determine issuer (per-tenant domain or global default)
	issuer := s.cfg.Security.Issuer
	if input.Issuer != "" {
		issuer = input.Issuer
	}

	// Build access token
	accessClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   input.UserID.String(),
			Audience:  jwt.ClaimStrings{input.ApplicationID.String()},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			ID:        uuid.New().String(),
		},
		TenantID:    input.TenantID.String(),
		Email:       input.Email,
		Name:        input.Name,
		Scope:       strings.Join(input.Scopes, " "),
		Permissions: input.Permissions,
	}
	if input.ActorID != nil {
		accessClaims.Act = &ActorClaim{Sub: input.ActorID.String()}
	}

	accessToken := jwt.NewWithClaims(s.keyManager.GetSigningMethod(), accessClaims)
	accessToken.Header["kid"] = activeKey.ID

	accessTokenStr, err := accessToken.SignedString(activeKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	pair := &TokenPair{
		AccessToken: accessTokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   int(accessTTL.Seconds()),
		Scope:       strings.Join(input.Scopes, " "),
	}

	// Issue refresh token if offline_access scope is requested
	if containsScope(input.Scopes, "offline_access") {
		refreshTokenRaw, err := crypto.GenerateOpaqueToken()
		if err != nil {
			return nil, fmt.Errorf("generating refresh token: %w", err)
		}

		family := uuid.New().String()
		rt := &models.RefreshToken{
			UserID:        input.UserID,
			ApplicationID: input.ApplicationID,
			TenantID:      input.TenantID,
			TokenHash:     crypto.HashToken(refreshTokenRaw),
			Family:        family,
			Revoked:       false,
			ExpiresAt:     now.Add(refreshTTL),
		}

		if err := s.refreshTokens.Create(ctx, rt); err != nil {
			return nil, fmt.Errorf("storing refresh token: %w", err)
		}

		pair.RefreshToken = refreshTokenRaw
	}

	// Issue ID token if openid scope is requested
	if containsScope(input.Scopes, "openid") {
		idClaims := IDTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    issuer,
				Subject:   input.UserID.String(),
				Audience:  jwt.ClaimStrings{input.ApplicationID.String()},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(idTTL)),
			},
			Nonce:    input.Nonce,
			AuthTime: now.Unix(),
		}

		if containsScope(input.Scopes, "email") {
			idClaims.Email = input.Email
			idClaims.EmailVerified = input.EmailVerified
		}
		if containsScope(input.Scopes, "profile") {
			idClaims.Name = input.Name
			idClaims.Picture = input.AvatarURL
		}
		if containsScope(input.Scopes, "phone") {
			idClaims.Phone = input.Phone
		}

		idToken := jwt.NewWithClaims(s.keyManager.GetSigningMethod(), idClaims)
		idToken.Header["kid"] = activeKey.ID

		idTokenStr, err := idToken.SignedString(activeKey.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("signing id token: %w", err)
		}
		pair.IDToken = idTokenStr
	}

	return pair, nil
}

// RefreshAccessToken rotates a refresh token and issues a new token pair.
func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenRaw string, appID uuid.UUID) (*TokenPair, error) {
	tokenHash := crypto.HashToken(refreshTokenRaw)

	rt, err := s.refreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if models.IsAppError(err, models.ErrNotFound) {
			return nil, models.ErrInvalidGrant.WithMessage("Refresh token not found.")
		}
		return nil, err
	}

	// Reuse detection: if the token is already revoked, revoke the entire family
	if rt.Revoked {
		s.logger.Warn("refresh token reuse detected, revoking family",
			zap.String("family", rt.Family),
			zap.String("user_id", rt.UserID.String()),
		)
		_ = s.refreshTokens.RevokeByFamily(ctx, rt.Family)
		return nil, models.ErrTokenRevoked.WithMessage("Refresh token reuse detected. All tokens in the family have been revoked.")
	}

	// Check expiration
	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, models.ErrTokenExpired.WithMessage("Refresh token has expired.")
	}

	// Check application match
	if rt.ApplicationID != appID {
		return nil, models.ErrInvalidClient
	}

	// Revoke the old refresh token
	if err := s.refreshTokens.Revoke(ctx, rt.ID); err != nil {
		return nil, err
	}

	// Issue new refresh token in the same family
	newRefreshRaw, err := crypto.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generating new refresh token: %w", err)
	}

	newRT := &models.RefreshToken{
		UserID:        rt.UserID,
		ApplicationID: rt.ApplicationID,
		TenantID:      rt.TenantID,
		TokenHash:     crypto.HashToken(newRefreshRaw),
		Family:        rt.Family,
		Revoked:       false,
		ExpiresAt:     time.Now().UTC().Add(s.cfg.Security.RefreshTokenLifetime),
	}

	if err := s.refreshTokens.Create(ctx, newRT); err != nil {
		return nil, err
	}

	// Return rotated refresh token (caller issues new access token with user claims)
	pair := &TokenPair{
		RefreshToken: newRefreshRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.cfg.Security.AccessTokenLifetime.Seconds()),
		UserID:       rt.UserID,
	}

	return pair, nil
}

// ValidateAccessToken validates and parses a JWT access token.
func (s *Service) ValidateAccessToken(ctx context.Context, tokenStr string) (*AccessTokenClaims, error) {
	// Check blacklist (if Redis available)
	if s.redis != nil {
		blacklisted, _ := s.redis.Exists(ctx, "token_blacklist:"+tokenStr).Result()
		if blacklisted > 0 {
			return nil, models.ErrTokenRevoked
		}
	}

	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid header")
		}
		kp := s.keyManager.FindKeyByID(kid)
		if kp == nil {
			return nil, fmt.Errorf("unknown key id: %s", kid)
		}
		return kp.PublicKey, nil
	})

	if err != nil {
		return nil, models.ErrUnauthorized.Wrap(err)
	}

	if !token.Valid {
		return nil, models.ErrUnauthorized
	}

	return claims, nil
}

// RevokeAccessToken blacklists a JWT access token in Redis.
func (s *Service) RevokeAccessToken(ctx context.Context, tokenStr string) error {
	claims := &AccessTokenClaims{}
	// Parse without validation to get expiry
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	_, _, err := parser.ParseUnverified(tokenStr, claims)
	if err != nil {
		return fmt.Errorf("parsing token for revocation: %w", err)
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil // Already expired
	}

	if s.redis != nil {
		return s.redis.Set(ctx, "token_blacklist:"+tokenStr, "1", ttl).Err()
	}
	return nil
}

// RevokeRefreshToken revokes a specific refresh token.
func (s *Service) RevokeRefreshToken(ctx context.Context, tokenRaw string) error {
	tokenHash := crypto.HashToken(tokenRaw)
	rt, err := s.refreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	return s.refreshTokens.Revoke(ctx, rt.ID)
}

// RevokeAllUserTokens revokes all refresh tokens for a user.
func (s *Service) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokens.RevokeByUser(ctx, userID)
}

// GetJWKS returns the public JWKS.
func (s *Service) GetJWKS() crypto.JWKS {
	return s.keyManager.GetJWKS()
}

// GetKeyManager returns the key manager.
func (s *Service) GetKeyManager() *crypto.KeyManager {
	return s.keyManager
}

func containsScope(scopes []string, target string) bool {
	for _, s := range scopes {
		if s == target {
			return true
		}
	}
	return false
}
