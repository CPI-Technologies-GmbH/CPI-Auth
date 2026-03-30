package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

type domainVerificationRepo struct {
	pool *pgxpool.Pool
}

func NewDomainVerificationRepository(pool *pgxpool.Pool) models.DomainVerificationRepository {
	return &domainVerificationRepo{pool: pool}
}

func (r *domainVerificationRepo) Create(ctx context.Context, dv *models.DomainVerification) error {
	dv.ID = uuid.New()
	now := time.Now().UTC()
	dv.CreatedAt = now
	dv.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO domain_verifications (id, tenant_id, domain, verification_token, verification_method, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		dv.ID, dv.TenantID, dv.Domain, dv.VerificationToken, dv.VerificationMethod,
		dv.IsVerified, dv.CreatedAt, dv.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting domain verification: %w", err)
	}
	return nil
}

func (r *domainVerificationRepo) scanDV(row pgx.Row) (*models.DomainVerification, error) {
	var dv models.DomainVerification
	err := row.Scan(&dv.ID, &dv.TenantID, &dv.Domain, &dv.VerificationToken,
		&dv.VerificationMethod, &dv.IsVerified, &dv.VerifiedAt, &dv.CreatedAt, &dv.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &dv, nil
}

const dvSelectCols = `id, tenant_id, domain, verification_token, verification_method, is_verified, verified_at, created_at, updated_at`

func (r *domainVerificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.DomainVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+dvSelectCols+` FROM domain_verifications WHERE id = $1`, id)
	dv, err := r.scanDV(row)
	if err != nil {
		return nil, fmt.Errorf("querying domain verification by id: %w", err)
	}
	return dv, nil
}

func (r *domainVerificationRepo) GetByDomain(ctx context.Context, domain string) (*models.DomainVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+dvSelectCols+` FROM domain_verifications WHERE domain = $1`, domain)
	dv, err := r.scanDV(row)
	if err != nil {
		return nil, fmt.Errorf("querying domain verification by domain: %w", err)
	}
	return dv, nil
}

func (r *domainVerificationRepo) GetByTenant(ctx context.Context, tenantID uuid.UUID) (*models.DomainVerification, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+dvSelectCols+` FROM domain_verifications WHERE tenant_id = $1`, tenantID)
	dv, err := r.scanDV(row)
	if err != nil {
		return nil, fmt.Errorf("querying domain verification by tenant: %w", err)
	}
	return dv, nil
}

func (r *domainVerificationRepo) Update(ctx context.Context, dv *models.DomainVerification) error {
	dv.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE domain_verifications SET domain = $1, verification_token = $2, verification_method = $3,
			is_verified = $4, verified_at = $5, updated_at = $6
		WHERE id = $7`,
		dv.Domain, dv.VerificationToken, dv.VerificationMethod,
		dv.IsVerified, dv.VerifiedAt, dv.UpdatedAt, dv.ID,
	)
	if err != nil {
		return fmt.Errorf("updating domain verification: %w", err)
	}
	return nil
}

func (r *domainVerificationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM domain_verifications WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting domain verification: %w", err)
	}
	return nil
}
