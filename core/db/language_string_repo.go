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

type languageStringRepo struct {
	pool *pgxpool.Pool
}

func NewLanguageStringRepository(pool *pgxpool.Pool) models.LanguageStringRepository {
	return &languageStringRepo{pool: pool}
}

func (r *languageStringRepo) List(ctx context.Context, tenantID uuid.UUID, locale string) ([]models.LanguageString, error) {
	query := `SELECT id, tenant_id, string_key, locale, value, created_at, updated_at
		FROM template_language_strings WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if locale != "" {
		query += ` AND locale = $2`
		args = append(args, locale)
	}
	query += ` ORDER BY string_key, locale`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing language strings: %w", err)
	}
	defer rows.Close()

	var result []models.LanguageString
	for rows.Next() {
		var ls models.LanguageString
		if err := rows.Scan(&ls.ID, &ls.TenantID, &ls.StringKey, &ls.Locale, &ls.Value, &ls.CreatedAt, &ls.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning language string: %w", err)
		}
		result = append(result, ls)
	}
	return result, nil
}

func (r *languageStringRepo) Upsert(ctx context.Context, ls *models.LanguageString) error {
	now := time.Now().UTC()
	if ls.ID == uuid.Nil {
		ls.ID = uuid.New()
	}
	ls.CreatedAt = now
	ls.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO template_language_strings (id, tenant_id, string_key, locale, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, string_key, locale) DO UPDATE
		SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		ls.ID, ls.TenantID, ls.StringKey, ls.Locale, ls.Value, ls.CreatedAt, ls.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting language string: %w", err)
	}
	return nil
}

func (r *languageStringRepo) Delete(ctx context.Context, tenantID uuid.UUID, stringKey, locale string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM template_language_strings WHERE tenant_id = $1 AND string_key = $2 AND locale = $3`,
		tenantID, stringKey, locale,
	)
	if err != nil {
		return fmt.Errorf("deleting language string: %w", err)
	}
	return nil
}

func (r *languageStringRepo) GetByKeys(ctx context.Context, tenantID uuid.UUID, keys []string, locale string) (map[string]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT string_key, value FROM template_language_strings
		WHERE tenant_id = $1 AND locale = $2 AND string_key = ANY($3)`,
		tenantID, locale, keys,
	)
	if err != nil {
		return nil, fmt.Errorf("getting language strings by keys: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string, len(keys))
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scanning language string: %w", err)
		}
		result[key] = value
	}
	if err := rows.Err(); err != nil {
		if err == pgx.ErrNoRows {
			return result, nil
		}
		return nil, fmt.Errorf("iterating language strings: %w", err)
	}
	return result, nil
}
