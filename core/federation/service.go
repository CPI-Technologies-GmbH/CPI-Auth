package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// ProviderConfig holds the configuration for a social/OIDC provider.
type ProviderConfig struct {
	Name              string            `json:"name"`
	ClientID          string            `json:"client_id"`
	ClientSecret      string            `json:"client_secret"`
	AuthURL           string            `json:"auth_url"`
	TokenURL          string            `json:"token_url"`
	UserInfoURL       string            `json:"userinfo_url"`
	Scopes            []string          `json:"scopes"`
	ClaimMapping      map[string]string `json:"claim_mapping"`
	DiscoveryURL      string            `json:"discovery_url"`
}

// ProviderUser holds the normalized user data from a provider.
type ProviderUser struct {
	ProviderID   string          `json:"provider_id"`
	Email        string          `json:"email"`
	Name         string          `json:"name"`
	AvatarURL    string          `json:"avatar_url"`
	RawProfile   json.RawMessage `json:"raw_profile"`
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

// Service manages social login and OIDC federation.
type Service struct {
	identities models.IdentityRepository
	users      models.UserRepository
	providers  map[string]ProviderConfig
	cfg        *config.Config
	logger     *zap.Logger
	httpClient *http.Client
}

// NewService creates a new federation service.
func NewService(identities models.IdentityRepository, users models.UserRepository, cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		identities: identities,
		users:      users,
		providers:  make(map[string]ProviderConfig),
		cfg:        cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// RegisterProvider adds a social login provider configuration.
func (s *Service) RegisterProvider(name string, config ProviderConfig) {
	config.Name = name
	s.providers[name] = config
}

// GetAuthorizationURL generates the authorization URL for a provider.
func (s *Service) GetAuthorizationURL(providerName, state, redirectURI string) (string, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return "", models.ErrNotFound.WithMessage(fmt.Sprintf("Unknown provider: %s", providerName))
	}

	params := url.Values{
		"client_id":     {provider.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(provider.Scopes, " ")},
		"state":         {state},
	}

	return provider.AuthURL + "?" + params.Encode(), nil
}

// ExchangeCode exchanges an authorization code for tokens and fetches user profile.
func (s *Service) ExchangeCode(ctx context.Context, providerName, code, redirectURI string) (*ProviderUser, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return nil, models.ErrNotFound.WithMessage(fmt.Sprintf("Unknown provider: %s", providerName))
	}

	// Exchange code for tokens
	tokenResp, err := s.exchangeToken(ctx, provider, code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	// Fetch user profile
	return s.fetchUserProfile(ctx, provider, tokenResp)
}

// HandleCallback processes the callback from a social provider, performing JIT provisioning.
func (s *Service) HandleCallback(ctx context.Context, tenantID uuid.UUID, providerName, code, redirectURI string) (*models.User, *models.Identity, error) {
	providerUser, err := s.ExchangeCode(ctx, providerName, code, redirectURI)
	if err != nil {
		return nil, nil, err
	}

	// Check if identity already exists
	identity, err := s.identities.GetByProvider(ctx, providerName, providerUser.ProviderID)
	if err == nil {
		// Existing identity: update tokens and return existing user
		encryptedTokens, _ := crypto.Encrypt(
			[]byte(fmt.Sprintf(`{"access_token":"%s","refresh_token":"%s"}`, providerUser.AccessToken, providerUser.RefreshToken)),
			s.cfg.Security.EncryptionKey,
		)
		identity.TokensEncrypted = encryptedTokens
		identity.Profile = providerUser.RawProfile
		_ = s.identities.Update(ctx, identity)

		user, err := s.users.GetByID(ctx, tenantID, identity.UserID)
		if err != nil {
			return nil, nil, err
		}
		return user, identity, nil
	}

	// JIT provisioning: check if a user with this email exists
	var user *models.User
	if providerUser.Email != "" {
		user, err = s.users.GetByEmail(ctx, tenantID, providerUser.Email)
		if err != nil && !models.IsAppError(err, models.ErrNotFound) {
			return nil, nil, err
		}
	}

	// Create new user if not found
	if user == nil {
		user = &models.User{
			TenantID:      tenantID,
			Email:         providerUser.Email,
			Name:          providerUser.Name,
			AvatarURL:     providerUser.AvatarURL,
			Status:        models.StatusActive,
			EmailVerified: providerUser.Email != "",
		}
		if err := s.users.Create(ctx, user); err != nil {
			return nil, nil, fmt.Errorf("creating JIT user: %w", err)
		}
	}

	// Create identity link
	encryptedTokens, _ := crypto.Encrypt(
		[]byte(fmt.Sprintf(`{"access_token":"%s","refresh_token":"%s"}`, providerUser.AccessToken, providerUser.RefreshToken)),
		s.cfg.Security.EncryptionKey,
	)

	identity = &models.Identity{
		UserID:          user.ID,
		Provider:        providerName,
		ProviderUserID:  providerUser.ProviderID,
		TokensEncrypted: encryptedTokens,
		Profile:         providerUser.RawProfile,
	}
	if err := s.identities.Create(ctx, identity); err != nil {
		return nil, nil, fmt.Errorf("creating identity: %w", err)
	}

	return user, identity, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

func (s *Service) exchangeToken(ctx context.Context, provider ProviderConfig, code, redirectURI string) (*tokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {provider.ClientID},
		"client_secret": {provider.ClientSecret},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func (s *Service) fetchUserProfile(ctx context.Context, provider ProviderConfig, tokenResp *tokenResponse) (*ProviderUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var profile map[string]interface{}
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}

	user := &ProviderUser{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		RawProfile:   body,
	}

	if tokenResp.ExpiresIn > 0 {
		user.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	// Apply claim mapping
	mapping := provider.ClaimMapping
	if mapping == nil {
		mapping = defaultClaimMapping()
	}

	if field, ok := mapping["id"]; ok {
		if v, ok := profile[field]; ok {
			user.ProviderID = fmt.Sprintf("%v", v)
		}
	}
	if field, ok := mapping["email"]; ok {
		if v, ok := profile[field].(string); ok {
			user.Email = v
		}
	}
	if field, ok := mapping["name"]; ok {
		if v, ok := profile[field].(string); ok {
			user.Name = v
		}
	}
	if field, ok := mapping["avatar"]; ok {
		if v, ok := profile[field].(string); ok {
			user.AvatarURL = v
		}
	}

	return user, nil
}

func defaultClaimMapping() map[string]string {
	return map[string]string{
		"id":     "sub",
		"email":  "email",
		"name":   "name",
		"avatar": "picture",
	}
}

// RegisterDefaultProviders sets up well-known social login providers.
func (s *Service) RegisterDefaultProviders(configs map[string]ProviderConfig) {
	defaults := map[string]ProviderConfig{
		"google": {
			AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:    "https://oauth2.googleapis.com/token",
			UserInfoURL: "https://openidconnect.googleapis.com/v1/userinfo",
			Scopes:      []string{"openid", "email", "profile"},
			ClaimMapping: map[string]string{
				"id": "sub", "email": "email", "name": "name", "avatar": "picture",
			},
		},
		"github": {
			AuthURL:     "https://github.com/login/oauth/authorize",
			TokenURL:    "https://github.com/login/oauth/access_token",
			UserInfoURL: "https://api.github.com/user",
			Scopes:      []string{"user:email"},
			ClaimMapping: map[string]string{
				"id": "id", "email": "email", "name": "name", "avatar": "avatar_url",
			},
		},
		"microsoft": {
			AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			UserInfoURL: "https://graph.microsoft.com/v1.0/me",
			Scopes:      []string{"openid", "email", "profile"},
			ClaimMapping: map[string]string{
				"id": "id", "email": "mail", "name": "displayName", "avatar": "photo",
			},
		},
		"apple": {
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
			Scopes:   []string{"name", "email"},
			ClaimMapping: map[string]string{
				"id": "sub", "email": "email", "name": "name",
			},
		},
		"discord": {
			AuthURL:     "https://discord.com/api/oauth2/authorize",
			TokenURL:    "https://discord.com/api/oauth2/token",
			UserInfoURL: "https://discord.com/api/users/@me",
			Scopes:      []string{"identify", "email"},
			ClaimMapping: map[string]string{
				"id": "id", "email": "email", "name": "username", "avatar": "avatar",
			},
		},
		"slack": {
			AuthURL:     "https://slack.com/openid/connect/authorize",
			TokenURL:    "https://slack.com/api/openid.connect.token",
			UserInfoURL: "https://slack.com/api/openid.connect.userInfo",
			Scopes:      []string{"openid", "email", "profile"},
			ClaimMapping: map[string]string{
				"id": "sub", "email": "email", "name": "name", "avatar": "picture",
			},
		},
	}

	for name, defaultCfg := range defaults {
		if userCfg, ok := configs[name]; ok {
			if userCfg.AuthURL == "" {
				userCfg.AuthURL = defaultCfg.AuthURL
			}
			if userCfg.TokenURL == "" {
				userCfg.TokenURL = defaultCfg.TokenURL
			}
			if userCfg.UserInfoURL == "" {
				userCfg.UserInfoURL = defaultCfg.UserInfoURL
			}
			if len(userCfg.Scopes) == 0 {
				userCfg.Scopes = defaultCfg.Scopes
			}
			if userCfg.ClaimMapping == nil {
				userCfg.ClaimMapping = defaultCfg.ClaimMapping
			}
			s.RegisterProvider(name, userCfg)
		}
	}
}
