package saml

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Service manages SAML 2.0 Service Provider and Identity Provider operations.
type Service struct {
	sp         *saml.ServiceProvider
	userRepo   models.UserRepository
	identRepo  models.IdentityRepository
	cfg        *config.Config
	logger     *zap.Logger
	privateKey *rsa.PrivateKey
	cert       *x509.Certificate
}

// NewService creates a new SAML service.
func NewService(
	userRepo models.UserRepository,
	identRepo models.IdentityRepository,
	cfg *config.Config,
	logger *zap.Logger,
	privateKey *rsa.PrivateKey,
	cert *x509.Certificate,
) *Service {
	s := &Service{
		userRepo:   userRepo,
		identRepo:  identRepo,
		cfg:        cfg,
		logger:     logger,
		privateKey: privateKey,
		cert:       cert,
	}

	if privateKey != nil && cert != nil {
		s.sp = &saml.ServiceProvider{
			Key:         privateKey,
			Certificate: cert,
			MetadataURL: mustParseURL(cfg.Server.BaseURL + "/saml/metadata"),
			AcsURL:      mustParseURL(cfg.Server.BaseURL + "/saml/acs"),
			IDPMetadata: &saml.EntityDescriptor{},
		}
	}

	return s
}

// GetMetadata returns the SP metadata XML.
func (s *Service) GetMetadata() ([]byte, error) {
	if s.sp == nil {
		return nil, fmt.Errorf("SAML service not configured")
	}
	descriptor := s.sp.Metadata()
	return xml.MarshalIndent(descriptor, "", "  ")
}

// InitiateSSO generates an authentication request for SP-initiated SSO.
func (s *Service) InitiateSSO(w http.ResponseWriter, r *http.Request) error {
	if s.sp == nil {
		return fmt.Errorf("SAML service not configured")
	}

	// Generate AuthnRequest
	authnRequest, err := s.sp.MakeAuthenticationRequest(
		s.sp.GetSSOBindingLocation(saml.HTTPRedirectBinding),
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		return fmt.Errorf("creating authn request: %w", err)
	}

	redirectURL, err := authnRequest.Redirect("", s.sp)
	if err != nil {
		return fmt.Errorf("generating redirect URL: %w", err)
	}

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
	return nil
}

// HandleACS processes the SAML Assertion Consumer Service callback.
func (s *Service) HandleACS(ctx context.Context, tenantID uuid.UUID, samlResponse string) (*models.User, error) {
	if s.sp == nil {
		return nil, fmt.Errorf("SAML service not configured")
	}

	// Parse and validate the response
	assertion, err := s.sp.ParseResponse(nil, []string{})
	if err != nil {
		s.logger.Error("failed to parse SAML response", zap.Error(err))
		return nil, models.ErrSAMLFailed.Wrap(err)
	}

	// Extract attributes
	attrs := extractAttributes(assertion)
	email := attrs["email"]
	name := attrs["name"]
	providerID := assertion.Subject.NameID.Value

	if email == "" && providerID == "" {
		return nil, models.ErrSAMLFailed.WithMessage("SAML response missing email and NameID.")
	}

	// JIT provisioning
	identity, err := s.identRepo.GetByProvider(ctx, "saml", providerID)
	if err == nil {
		user, err := s.userRepo.GetByID(ctx, tenantID, identity.UserID)
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	// Check for existing user by email
	var user *models.User
	if email != "" {
		user, err = s.userRepo.GetByEmail(ctx, tenantID, email)
		if err != nil && !models.IsAppError(err, models.ErrNotFound) {
			return nil, err
		}
	}

	// Create new user
	if user == nil {
		user = &models.User{
			TenantID:      tenantID,
			Email:         email,
			Name:          name,
			Status:        models.StatusActive,
			EmailVerified: true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	// Link SAML identity
	identity = &models.Identity{
		UserID:         user.ID,
		Provider:       "saml",
		ProviderUserID: providerID,
	}
	if err := s.identRepo.Create(ctx, identity); err != nil {
		return nil, err
	}

	return user, nil
}

func extractAttributes(assertion *saml.Assertion) map[string]string {
	attrs := make(map[string]string)
	if assertion == nil {
		return attrs
	}

	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if len(attr.Values) > 0 {
				name := attr.FriendlyName
				if name == "" {
					name = attr.Name
				}
				attrs[name] = attr.Values[0].Value
			}
		}
	}
	return attrs
}

// IdPMetadata generates Identity Provider metadata when CPI Auth acts as an IdP.
type IdPMetadata struct {
	EntityID     string    `xml:"entityID,attr"`
	ValidUntil   time.Time `xml:"validUntil,attr,omitempty"`
	SSOServices  []SSOService
	NameIDFormat string
}

// SSOService describes an SSO endpoint.
type SSOService struct {
	Binding  string
	Location string
}

func mustParseURL(rawURL string) url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("invalid URL: %s", rawURL))
	}
	return *u
}

// GenerateIdPMetadata returns the IdP metadata for CPI Auth acting as an Identity Provider.
func (s *Service) GenerateIdPMetadata() (*IdPMetadata, error) {
	return &IdPMetadata{
		EntityID:   s.cfg.Server.BaseURL,
		ValidUntil: time.Now().Add(24 * time.Hour * 365),
		SSOServices: []SSOService{
			{
				Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
				Location: s.cfg.Server.BaseURL + "/saml/sso",
			},
			{
				Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
				Location: s.cfg.Server.BaseURL + "/saml/sso",
			},
		},
		NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
	}, nil
}
