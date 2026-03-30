package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

type deviceCodeRepo struct{ pool *pgxpool.Pool }

// NewDeviceCodeRepository creates a new device code repository.
func NewDeviceCodeRepository(pool *pgxpool.Pool) models.DeviceCodeRepository {
	return &deviceCodeRepo{pool: pool}
}

func (r *deviceCodeRepo) Create(ctx context.Context, dc *models.DeviceCode) error {
	if dc.ID == uuid.Nil {
		dc.ID = uuid.New()
	}
	dc.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO device_codes (id, tenant_id, device_code, user_code, client_id, scopes, status, expires_at, poll_interval, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		dc.ID, dc.TenantID, dc.DeviceCode, dc.UserCode, dc.ClientID,
		dc.Scopes, dc.Status, dc.ExpiresAt, dc.PollInterval, dc.CreatedAt)
	return err
}

func (r *deviceCodeRepo) GetByDeviceCode(ctx context.Context, deviceCode string) (*models.DeviceCode, error) {
	var dc models.DeviceCode
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, device_code, user_code, client_id, scopes, status, user_id, expires_at, poll_interval, created_at
		FROM device_codes WHERE device_code = $1`, deviceCode).
		Scan(&dc.ID, &dc.TenantID, &dc.DeviceCode, &dc.UserCode, &dc.ClientID,
			&dc.Scopes, &dc.Status, &dc.UserID, &dc.ExpiresAt, &dc.PollInterval, &dc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &dc, err
}

func (r *deviceCodeRepo) GetByUserCode(ctx context.Context, userCode string) (*models.DeviceCode, error) {
	var dc models.DeviceCode
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, device_code, user_code, client_id, scopes, status, user_id, expires_at, poll_interval, created_at
		FROM device_codes WHERE user_code = $1`, userCode).
		Scan(&dc.ID, &dc.TenantID, &dc.DeviceCode, &dc.UserCode, &dc.ClientID,
			&dc.Scopes, &dc.Status, &dc.UserID, &dc.ExpiresAt, &dc.PollInterval, &dc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &dc, err
}

func (r *deviceCodeRepo) Authorize(ctx context.Context, userCode string, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE device_codes SET status = 'authorized', user_id = $1
		WHERE user_code = $2 AND status = 'pending'`, userID, userCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *deviceCodeRepo) Deny(ctx context.Context, userCode string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE device_codes SET status = 'denied'
		WHERE user_code = $1 AND status = 'pending'`, userCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *deviceCodeRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM device_codes WHERE expires_at < NOW()`)
	return err
}
