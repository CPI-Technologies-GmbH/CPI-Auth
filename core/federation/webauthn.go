package federation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// WebAuthnUser wraps our User model for the go-webauthn library.
type WebAuthnUser struct {
	user        *models.User
	credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return u.user.ID[:]
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.user.Email
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.user.Name != "" {
		return u.user.Name
	}
	return u.user.Email
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// WebAuthnService manages FIDO2/WebAuthn registration and authentication.
type WebAuthnService struct {
	wan        *webauthn.WebAuthn
	credRepo   models.WebAuthnCredentialRepository
	userRepo   models.UserRepository
	redis      *redis.Client
	logger     *zap.Logger
}

// NewWebAuthnService creates a new WebAuthn service.
func NewWebAuthnService(credRepo models.WebAuthnCredentialRepository, userRepo models.UserRepository, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) (*WebAuthnService, error) {
	wan, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "CPI Auth",
		RPID:          extractHost(cfg.Server.BaseURL),
		RPOrigins:     []string{cfg.Server.BaseURL},
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:        protocol.ResidentKeyRequirementPreferred,
			UserVerification:   protocol.VerificationPreferred,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initializing webauthn: %w", err)
	}

	return &WebAuthnService{
		wan:      wan,
		credRepo: credRepo,
		userRepo: userRepo,
		redis:    rdb,
		logger:   logger,
	}, nil
}

// BeginRegistration initiates WebAuthn credential registration.
func (s *WebAuthnService) BeginRegistration(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (*protocol.CredentialCreation, error) {
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	existingCreds, err := s.credRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	wanUser := &WebAuthnUser{
		user:        user,
		credentials: convertToWebAuthnCredentials(existingCreds),
	}

	creation, session, err := s.wan.BeginRegistration(wanUser)
	if err != nil {
		return nil, fmt.Errorf("begin registration: %w", err)
	}

	// Store session in Redis
	sessionJSON, _ := json.Marshal(session)
	s.redis.Set(ctx, fmt.Sprintf("webauthn_reg:%s", userID), sessionJSON, 0)

	return creation, nil
}

// FinishRegistration completes WebAuthn credential registration.
func (s *WebAuthnService) FinishRegistration(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, credName string, response *protocol.ParsedCredentialCreationData) error {
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	existingCreds, err := s.credRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	wanUser := &WebAuthnUser{
		user:        user,
		credentials: convertToWebAuthnCredentials(existingCreds),
	}

	// Retrieve session from Redis
	sessionJSON, err := s.redis.Get(ctx, fmt.Sprintf("webauthn_reg:%s", userID)).Bytes()
	if err != nil {
		return models.ErrWebAuthnFailed.WithMessage("Registration session not found or expired.")
	}
	s.redis.Del(ctx, fmt.Sprintf("webauthn_reg:%s", userID))

	var session webauthn.SessionData
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		return models.ErrWebAuthnFailed.Wrap(err)
	}

	credential, err := s.wan.CreateCredential(wanUser, session, response)
	if err != nil {
		return models.ErrWebAuthnFailed.Wrap(err)
	}

	// Store credential
	cred := &models.WebAuthnCredential{
		UserID:       userID,
		CredentialID: credential.ID,
		PublicKey:    credential.PublicKey,
		SignCount:    credential.Authenticator.SignCount,
		AAGUID:       credential.Authenticator.AAGUID,
		Name:         credName,
	}

	return s.credRepo.Create(ctx, cred)
}

// BeginLogin initiates WebAuthn authentication.
func (s *WebAuthnService) BeginLogin(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (*protocol.CredentialAssertion, error) {
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	creds, err := s.credRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(creds) == 0 {
		return nil, models.ErrWebAuthnFailed.WithMessage("No WebAuthn credentials registered.")
	}

	wanUser := &WebAuthnUser{
		user:        user,
		credentials: convertToWebAuthnCredentials(creds),
	}

	assertion, session, err := s.wan.BeginLogin(wanUser)
	if err != nil {
		return nil, models.ErrWebAuthnFailed.Wrap(err)
	}

	sessionJSON, _ := json.Marshal(session)
	s.redis.Set(ctx, fmt.Sprintf("webauthn_login:%s", userID), sessionJSON, 0)

	return assertion, nil
}

// FinishLogin completes WebAuthn authentication.
func (s *WebAuthnService) FinishLogin(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, response *protocol.ParsedCredentialAssertionData) error {
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	creds, err := s.credRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	wanUser := &WebAuthnUser{
		user:        user,
		credentials: convertToWebAuthnCredentials(creds),
	}

	sessionJSON, err := s.redis.Get(ctx, fmt.Sprintf("webauthn_login:%s", userID)).Bytes()
	if err != nil {
		return models.ErrWebAuthnFailed.WithMessage("Login session not found or expired.")
	}
	s.redis.Del(ctx, fmt.Sprintf("webauthn_login:%s", userID))

	var session webauthn.SessionData
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		return models.ErrWebAuthnFailed.Wrap(err)
	}

	credential, err := s.wan.ValidateLogin(wanUser, session, response)
	if err != nil {
		return models.ErrWebAuthnFailed.Wrap(err)
	}

	// Update sign count
	for _, c := range creds {
		if string(c.CredentialID) == string(credential.ID) {
			c.SignCount = credential.Authenticator.SignCount
			_ = s.credRepo.Update(ctx, &c)
			break
		}
	}

	return nil
}

// ListCredentials returns all WebAuthn credentials for a user.
func (s *WebAuthnService) ListCredentials(ctx context.Context, userID uuid.UUID) ([]models.WebAuthnCredential, error) {
	return s.credRepo.ListByUser(ctx, userID)
}

// DeleteCredential removes a WebAuthn credential.
func (s *WebAuthnService) DeleteCredential(ctx context.Context, credID uuid.UUID) error {
	return s.credRepo.Delete(ctx, credID)
}

func convertToWebAuthnCredentials(creds []models.WebAuthnCredential) []webauthn.Credential {
	result := make([]webauthn.Credential, len(creds))
	for i, c := range creds {
		result[i] = webauthn.Credential{
			ID:        c.CredentialID,
			PublicKey: c.PublicKey,
			Authenticator: webauthn.Authenticator{
				AAGUID:    c.AAGUID,
				SignCount: c.SignCount,
			},
		}
	}
	return result
}

func extractHost(baseURL string) string {
	if idx := len("https://"); len(baseURL) > idx {
		host := baseURL[idx:]
		if colonIdx := indexByte(host, ':'); colonIdx > 0 {
			return host[:colonIdx]
		}
		if slashIdx := indexByte(host, '/'); slashIdx > 0 {
			return host[:slashIdx]
		}
		return host
	}
	if idx := len("http://"); len(baseURL) > idx {
		host := baseURL[idx:]
		if colonIdx := indexByte(host, ':'); colonIdx > 0 {
			return host[:colonIdx]
		}
		if slashIdx := indexByte(host, '/'); slashIdx > 0 {
			return host[:slashIdx]
		}
		return host
	}
	return "localhost"
}

func indexByte(s string, b byte) int {
	for i := range s {
		if s[i] == b {
			return i
		}
	}
	return -1
}
