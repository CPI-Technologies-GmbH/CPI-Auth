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

type customFieldRepo struct {
	pool *pgxpool.Pool
}

func NewCustomFieldDefinitionRepository(pool *pgxpool.Pool) models.CustomFieldDefinitionRepository {
	return &customFieldRepo{pool: pool}
}

func (r *customFieldRepo) Create(ctx context.Context, field *models.CustomFieldDefinition) error {
	field.ID = uuid.New()
	now := time.Now().UTC()
	field.CreatedAt = now
	field.UpdatedAt = now
	if field.VisibleOn == "" {
		field.VisibleOn = "both"
	}
	field.IsActive = true

	_, err := r.pool.Exec(ctx, `
		INSERT INTO custom_field_definitions (id, tenant_id, name, label, field_type, placeholder, description, options, required, visible_on, position, validation_rules, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		field.ID, field.TenantID, field.Name, field.Label, field.FieldType,
		field.Placeholder, field.Description, field.Options, field.Required,
		field.VisibleOn, field.Position, field.ValidationRules, field.IsActive,
		field.CreatedAt, field.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting custom field definition: %w", err)
	}
	return nil
}

func (r *customFieldRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.CustomFieldDefinition, error) {
	var f models.CustomFieldDefinition
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, label, field_type, COALESCE(placeholder, ''), COALESCE(description, ''),
			options, required, visible_on, position, validation_rules, is_active, created_at, updated_at
		FROM custom_field_definitions WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&f.ID, &f.TenantID, &f.Name, &f.Label, &f.FieldType,
			&f.Placeholder, &f.Description, &f.Options, &f.Required,
			&f.VisibleOn, &f.Position, &f.ValidationRules, &f.IsActive,
			&f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("querying custom field definition: %w", err)
	}
	return &f, nil
}

func (r *customFieldRepo) Update(ctx context.Context, field *models.CustomFieldDefinition) error {
	field.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE custom_field_definitions SET label = $1, field_type = $2, placeholder = $3, description = $4,
			options = $5, required = $6, visible_on = $7, position = $8, validation_rules = $9, is_active = $10, updated_at = $11
		WHERE id = $12 AND tenant_id = $13`,
		field.Label, field.FieldType, field.Placeholder, field.Description,
		field.Options, field.Required, field.VisibleOn, field.Position,
		field.ValidationRules, field.IsActive, field.UpdatedAt, field.ID, field.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating custom field definition: %w", err)
	}
	return nil
}

func (r *customFieldRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM custom_field_definitions WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return fmt.Errorf("deleting custom field definition: %w", err)
	}
	return nil
}

func (r *customFieldRepo) List(ctx context.Context, tenantID uuid.UUID) ([]models.CustomFieldDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, label, field_type, COALESCE(placeholder, ''), COALESCE(description, ''),
			options, required, visible_on, position, validation_rules, is_active, created_at, updated_at
		FROM custom_field_definitions WHERE tenant_id = $1 ORDER BY position ASC, created_at ASC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing custom field definitions: %w", err)
	}
	defer rows.Close()

	var fields []models.CustomFieldDefinition
	for rows.Next() {
		var f models.CustomFieldDefinition
		if err := rows.Scan(&f.ID, &f.TenantID, &f.Name, &f.Label, &f.FieldType,
			&f.Placeholder, &f.Description, &f.Options, &f.Required,
			&f.VisibleOn, &f.Position, &f.ValidationRules, &f.IsActive,
			&f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning custom field definition: %w", err)
		}
		fields = append(fields, f)
	}
	if fields == nil {
		fields = []models.CustomFieldDefinition{}
	}
	return fields, nil
}
