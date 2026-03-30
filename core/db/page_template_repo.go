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

type pageTemplateRepo struct {
	pool *pgxpool.Pool
}

func NewPageTemplateRepository(pool *pgxpool.Pool) models.PageTemplateRepository {
	return &pageTemplateRepo{pool: pool}
}

const ptSelectCols = `id, tenant_id, page_type, name, html_content, css_content, is_active, is_default, created_at, updated_at`

func (r *pageTemplateRepo) scanPT(row pgx.Row) (*models.PageTemplate, error) {
	var pt models.PageTemplate
	err := row.Scan(&pt.ID, &pt.TenantID, &pt.PageType, &pt.Name,
		&pt.HTMLContent, &pt.CSSContent, &pt.IsActive, &pt.IsDefault, &pt.CreatedAt, &pt.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &pt, nil
}

func (r *pageTemplateRepo) Create(ctx context.Context, tmpl *models.PageTemplate) error {
	tmpl.ID = uuid.New()
	now := time.Now().UTC()
	tmpl.CreatedAt = now
	tmpl.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO page_templates (id, tenant_id, page_type, name, html_content, css_content, is_active, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		tmpl.ID, tmpl.TenantID, tmpl.PageType, tmpl.Name,
		tmpl.HTMLContent, tmpl.CSSContent, tmpl.IsActive, tmpl.IsDefault, tmpl.CreatedAt, tmpl.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting page template: %w", err)
	}
	return nil
}

func (r *pageTemplateRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.PageTemplate, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+ptSelectCols+` FROM page_templates WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	pt, err := r.scanPT(row)
	if err != nil {
		return nil, fmt.Errorf("querying page template by id: %w", err)
	}
	return pt, nil
}

func (r *pageTemplateRepo) GetByType(ctx context.Context, tenantID uuid.UUID, pageType string) (*models.PageTemplate, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+ptSelectCols+` FROM page_templates WHERE tenant_id = $1 AND page_type = $2`,
		tenantID, pageType)
	pt, err := r.scanPT(row)
	if err != nil {
		return nil, fmt.Errorf("querying page template by type: %w", err)
	}
	return pt, nil
}

func (r *pageTemplateRepo) Update(ctx context.Context, tmpl *models.PageTemplate) error {
	if tmpl.IsDefault {
		return fmt.Errorf("updating page template: %w", models.ErrForbidden.WithMessage("Default templates cannot be modified."))
	}
	tmpl.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE page_templates SET page_type = $1, name = $2,
			html_content = $3, css_content = $4, is_active = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8`,
		tmpl.PageType, tmpl.Name,
		tmpl.HTMLContent, tmpl.CSSContent, tmpl.IsActive, tmpl.UpdatedAt,
		tmpl.ID, tmpl.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating page template: %w", err)
	}
	return nil
}

func (r *pageTemplateRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	// Check if the template is a default template
	var isDefault bool
	err := r.pool.QueryRow(ctx, `SELECT is_default FROM page_templates WHERE id = $1 AND tenant_id = $2`, id, tenantID).Scan(&isDefault)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("deleting page template: %w", models.ErrNotFound)
		}
		return fmt.Errorf("deleting page template: %w", err)
	}
	if isDefault {
		return fmt.Errorf("deleting page template: %w", models.ErrForbidden.WithMessage("Default templates cannot be deleted."))
	}
	_, err = r.pool.Exec(ctx, `DELETE FROM page_templates WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return fmt.Errorf("deleting page template: %w", err)
	}
	return nil
}

func (r *pageTemplateRepo) List(ctx context.Context, tenantID uuid.UUID) ([]models.PageTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+ptSelectCols+` FROM page_templates WHERE tenant_id = $1 ORDER BY is_default DESC, page_type, name`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing page templates: %w", err)
	}
	defer rows.Close()

	var templates []models.PageTemplate
	for rows.Next() {
		var pt models.PageTemplate
		if err := rows.Scan(&pt.ID, &pt.TenantID, &pt.PageType, &pt.Name,
			&pt.HTMLContent, &pt.CSSContent, &pt.IsActive, &pt.IsDefault, &pt.CreatedAt, &pt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning page template: %w", err)
		}
		templates = append(templates, pt)
	}
	return templates, nil
}

func (r *pageTemplateRepo) Duplicate(ctx context.Context, tenantID, sourceID uuid.UUID, newName string) (*models.PageTemplate, error) {
	src, err := r.GetByID(ctx, tenantID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("duplicating page template: %w", err)
	}
	dup := &models.PageTemplate{
		TenantID:    tenantID,
		PageType:    src.PageType,
		Name:        newName,
		HTMLContent: src.HTMLContent,
		CSSContent:  src.CSSContent,
		IsActive:    false,
		IsDefault:   false,
	}
	if err := r.Create(ctx, dup); err != nil {
		return nil, fmt.Errorf("duplicating page template: %w", err)
	}
	return dup, nil
}
