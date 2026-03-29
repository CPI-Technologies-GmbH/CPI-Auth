package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

type tenantRepo struct {
	pool *pgxpool.Pool
}

func NewTenantRepository(pool *pgxpool.Pool) models.TenantRepository {
	return &tenantRepo{pool: pool}
}

func (r *tenantRepo) Create(ctx context.Context, tenant *models.Tenant) error {
	tenant.ID = uuid.New()
	now := time.Now().UTC()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now
	if tenant.Settings == nil {
		tenant.Settings = json.RawMessage(`{}`)
	}
	if tenant.Branding == nil {
		tenant.Branding = json.RawMessage(`{}`)
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO tenants (id, name, slug, domain, parent_id, settings, branding, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Domain, tenant.ParentID,
		tenant.Settings, tenant.Branding, tenant.CreatedAt, tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting tenant: %w", err)
	}
	return nil
}

func (r *tenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var t models.Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(domain, ''), parent_id, settings, branding, created_at, updated_at
		FROM tenants WHERE id = $1`, id).
		Scan(&t.ID, &t.Name, &t.Slug, &t.Domain, &t.ParentID, &t.Settings, &t.Branding, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("querying tenant: %w", err)
	}
	return &t, nil
}

func (r *tenantRepo) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(domain, ''), parent_id, settings, branding, created_at, updated_at
		FROM tenants WHERE slug = $1`, slug).
		Scan(&t.ID, &t.Name, &t.Slug, &t.Domain, &t.ParentID, &t.Settings, &t.Branding, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("querying tenant by slug: %w", err)
	}
	return &t, nil
}

func (r *tenantRepo) GetByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	var t models.Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(domain, ''), parent_id, settings, branding, created_at, updated_at
		FROM tenants WHERE domain = $1`, domain).
		Scan(&t.ID, &t.Name, &t.Slug, &t.Domain, &t.ParentID, &t.Settings, &t.Branding, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("querying tenant by domain: %w", err)
	}
	return &t, nil
}

func (r *tenantRepo) Update(ctx context.Context, tenant *models.Tenant) error {
	tenant.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE tenants SET name = $1, slug = $2, domain = $3, parent_id = $4, settings = $5, branding = $6, updated_at = $7
		WHERE id = $8`,
		tenant.Name, tenant.Slug, tenant.Domain, tenant.ParentID, tenant.Settings, tenant.Branding, tenant.UpdatedAt, tenant.ID,
	)
	if err != nil {
		return fmt.Errorf("updating tenant: %w", err)
	}
	return nil
}

func (r *tenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting tenant: %w", err)
	}
	return nil
}

func (r *tenantRepo) List(ctx context.Context, params models.PaginationParams) (*models.PaginatedResult[models.Tenant], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting tenants: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, COALESCE(domain, ''), parent_id, settings, branding, created_at, updated_at
		FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2`, params.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("listing tenants: %w", err)
	}
	defer rows.Close()

	var tenants []models.Tenant
	for rows.Next() {
		var t models.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Domain, &t.ParentID, &t.Settings, &t.Branding, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning tenant: %w", err)
		}
		tenants = append(tenants, t)
	}
	if tenants == nil {
		tenants = []models.Tenant{}
	}

	return &models.PaginatedResult[models.Tenant]{
		Data:       tenants,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}
