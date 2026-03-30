package domains

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Service handles domain verification operations.
type Service struct {
	dvRepo     models.DomainVerificationRepository
	tenantRepo models.TenantRepository
	logger     *zap.Logger
}

// NewService creates a new domain verification service.
func NewService(dvRepo models.DomainVerificationRepository, tenantRepo models.TenantRepository, logger *zap.Logger) *Service {
	return &Service{dvRepo: dvRepo, tenantRepo: tenantRepo, logger: logger}
}

// InitiateVerification creates a new domain verification record with a random token.
func (s *Service) InitiateVerification(ctx context.Context, tenantID uuid.UUID, domain string) (*models.DomainVerification, error) {
	domain = normalizeDomain(domain)
	if domain == "" {
		return nil, models.ErrBadRequest.WithMessage("Domain is required.")
	}

	// Check if domain is already verified by another tenant
	existing, err := s.dvRepo.GetByDomain(ctx, domain)
	if err == nil && existing.TenantID != tenantID {
		return nil, models.ErrConflict.WithMessage("Domain is already claimed by another tenant.")
	}

	// If this tenant already has a verification for this domain, return it
	if err == nil && existing.TenantID == tenantID {
		return existing, nil
	}

	token := generateToken()
	dv := &models.DomainVerification{
		TenantID:           tenantID,
		Domain:             domain,
		VerificationToken:  token,
		VerificationMethod: "TXT",
		IsVerified:         false,
	}

	if err := s.dvRepo.Create(ctx, dv); err != nil {
		return nil, fmt.Errorf("creating domain verification: %w", err)
	}

	return dv, nil
}

// CheckVerification performs DNS lookup and marks the domain as verified if the TXT record is found.
func (s *Service) CheckVerification(ctx context.Context, id uuid.UUID) (*models.DomainVerification, error) {
	dv, err := s.dvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dv.IsVerified {
		return dv, nil
	}

	verified := s.checkDNS(dv.Domain, dv.VerificationToken)
	if verified {
		now := time.Now().UTC()
		dv.IsVerified = true
		dv.VerifiedAt = &now

		if err := s.dvRepo.Update(ctx, dv); err != nil {
			return nil, fmt.Errorf("updating domain verification: %w", err)
		}

		// Update tenant's domain
		tenant, err := s.tenantRepo.GetByID(ctx, dv.TenantID)
		if err == nil {
			tenant.Domain = dv.Domain
			s.tenantRepo.Update(ctx, tenant)
		}

		s.logger.Info("domain verified", zap.String("domain", dv.Domain), zap.String("tenant_id", dv.TenantID.String()))
	}

	return dv, nil
}

// GetForTenant returns the domain verification record for a tenant.
func (s *Service) GetForTenant(ctx context.Context, tenantID uuid.UUID) (*models.DomainVerification, error) {
	return s.dvRepo.GetByTenant(ctx, tenantID)
}

// RemoveVerification deletes a domain verification and clears the tenant's domain.
func (s *Service) RemoveVerification(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	dv, err := s.dvRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if dv.TenantID != tenantID {
		return models.ErrForbidden
	}

	// Clear tenant domain if it matches
	tenant, err := s.tenantRepo.GetByID(ctx, dv.TenantID)
	if err == nil && tenant.Domain == dv.Domain {
		tenant.Domain = ""
		s.tenantRepo.Update(ctx, tenant)
	}

	return s.dvRepo.Delete(ctx, id)
}

// DNSInstructions returns the DNS record that needs to be created.
func DNSInstructions(dv *models.DomainVerification) map[string]string {
	return map[string]string{
		"record_type": "TXT",
		"host":        "_cpi-auth-verification." + dv.Domain,
		"value":       "cpi-auth-verify=" + dv.VerificationToken,
	}
}

func (s *Service) checkDNS(domain, token string) bool {
	expected := "cpi-auth-verify=" + token
	host := "_cpi-auth-verification." + domain

	records, err := net.LookupTXT(host)
	if err != nil {
		s.logger.Debug("DNS lookup failed", zap.String("host", host), zap.Error(err))
		return false
	}

	for _, record := range records {
		if strings.TrimSpace(record) == expected {
			return true
		}
	}
	return false
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	// Strip port if present
	if idx := strings.LastIndex(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	return domain
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
