package db

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

type appRepo struct {
	pool *pgxpool.Pool
}

func NewApplicationRepository(pool *pgxpool.Pool) models.ApplicationRepository {
	return &appRepo{pool: pool}
}

const appSelectCols = `id, tenant_id, name, COALESCE(description, ''), type, client_id, COALESCE(client_secret_hash, ''),
	COALESCE(logo_url, ''), redirect_uris, allowed_origins, post_logout_redirect_uris, grant_types,
	access_token_ttl, refresh_token_ttl, id_token_ttl, is_active, settings, created_at, updated_at`

func scanApp(row pgx.Row) (*models.Application, error) {
	var a models.Application
	err := row.Scan(&a.ID, &a.TenantID, &a.Name, &a.Description, &a.Type, &a.ClientID, &a.ClientSecretHash,
		&a.LogoURL, &a.RedirectURIs, &a.AllowedOrigins, &a.AllowedLogoutURLs, &a.GrantTypes,
		&a.AccessTokenTTL, &a.RefreshTokenTTL, &a.IDTokenTTL, &a.IsActive, &a.Settings, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *appRepo) Create(ctx context.Context, app *models.Application) error {
	app.ID = uuid.New()
	now := time.Now().UTC()
	app.CreatedAt = now
	app.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO applications (id, tenant_id, name, description, type, client_id, client_secret_hash,
			logo_url, redirect_uris, allowed_origins, post_logout_redirect_uris, grant_types,
			access_token_ttl, refresh_token_ttl, id_token_ttl, is_active, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		app.ID, app.TenantID, app.Name, app.Description, app.Type, app.ClientID, app.ClientSecretHash,
		app.LogoURL, app.RedirectURIs, app.AllowedOrigins, app.AllowedLogoutURLs, app.GrantTypes,
		app.AccessTokenTTL, app.RefreshTokenTTL, app.IDTokenTTL, app.IsActive, app.Settings, app.CreatedAt, app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting application: %w", err)
	}
	return nil
}

func (r *appRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Application, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+appSelectCols+` FROM applications WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	a, err := scanApp(row)
	if err != nil {
		return nil, fmt.Errorf("querying application: %w", err)
	}
	return a, nil
}

func (r *appRepo) GetByClientID(ctx context.Context, clientID string) (*models.Application, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+appSelectCols+` FROM applications WHERE client_id = $1`, clientID)
	a, err := scanApp(row)
	if err != nil {
		return nil, fmt.Errorf("querying application by client_id: %w", err)
	}
	return a, nil
}

func (r *appRepo) Update(ctx context.Context, app *models.Application) error {
	app.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE applications SET name = $1, description = $2, type = $3, logo_url = $4,
			redirect_uris = $5, allowed_origins = $6, post_logout_redirect_uris = $7, grant_types = $8,
			access_token_ttl = $9, refresh_token_ttl = $10, id_token_ttl = $11, is_active = $12,
			settings = $13, client_secret_hash = $14, updated_at = $15
		WHERE id = $16 AND tenant_id = $17`,
		app.Name, app.Description, app.Type, app.LogoURL,
		app.RedirectURIs, app.AllowedOrigins, app.AllowedLogoutURLs, app.GrantTypes,
		app.AccessTokenTTL, app.RefreshTokenTTL, app.IDTokenTTL, app.IsActive,
		app.Settings, app.ClientSecretHash, app.UpdatedAt, app.ID, app.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	return nil
}

func (r *appRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM applications WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return nil
}

func (r *appRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Application], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications WHERE tenant_id = $1`, tenantID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting applications: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT `+appSelectCols+`
		FROM applications WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("listing applications: %w", err)
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Name, &a.Description, &a.Type, &a.ClientID, &a.ClientSecretHash,
			&a.LogoURL, &a.RedirectURIs, &a.AllowedOrigins, &a.AllowedLogoutURLs, &a.GrantTypes,
			&a.AccessTokenTTL, &a.RefreshTokenTTL, &a.IDTokenTTL, &a.IsActive, &a.Settings, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning application: %w", err)
		}
		apps = append(apps, a)
	}
	if apps == nil {
		apps = []models.Application{}
	}

	return &models.PaginatedResult[models.Application]{
		Data:       apps,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}
