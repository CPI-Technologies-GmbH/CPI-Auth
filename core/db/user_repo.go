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

type userRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL-backed user repository.
func NewUserRepository(pool *pgxpool.Pool) models.UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	user.ID = uuid.New()
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Status == "" {
		user.Status = models.StatusActive
	}
	if user.Metadata == nil {
		user.Metadata = json.RawMessage(`{}`)
	}
	if user.AppMetadata == nil {
		user.AppMetadata = json.RawMessage(`{}`)
	}

	if user.Locale == "" {
		user.Locale = "en"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, phone, name, avatar_url, password_hash, locale, metadata, app_metadata, status, email_verified, phone_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		user.ID, user.TenantID, user.Email, user.Phone, user.Name, user.AvatarURL,
		user.PasswordHash, user.Locale, user.Metadata, user.AppMetadata, user.Status,
		user.EmailVerified, user.PhoneVerified, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, COALESCE(phone, ''), COALESCE(name, ''), COALESCE(avatar_url, ''), COALESCE(password_hash, ''), COALESCE(locale, 'en'), metadata, app_metadata, status, email_verified, phone_verified, created_at, updated_at
		FROM users WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&u.ID, &u.TenantID, &u.Email, &u.Phone, &u.Name, &u.AvatarURL,
			&u.PasswordHash, &u.Locale, &u.Metadata, &u.AppMetadata, &u.Status,
			&u.EmailVerified, &u.PhoneVerified, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return &u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, COALESCE(phone, ''), COALESCE(name, ''), COALESCE(avatar_url, ''), COALESCE(password_hash, ''), COALESCE(locale, 'en'), metadata, app_metadata, status, email_verified, phone_verified, created_at, updated_at
		FROM users WHERE email = $1 AND tenant_id = $2`, email, tenantID).
		Scan(&u.ID, &u.TenantID, &u.Email, &u.Phone, &u.Name, &u.AvatarURL,
			&u.PasswordHash, &u.Locale, &u.Metadata, &u.AppMetadata, &u.Status,
			&u.EmailVerified, &u.PhoneVerified, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return &u, nil
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now().UTC()
	if user.Locale == "" {
		user.Locale = "en"
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET email = $1, phone = $2, name = $3, avatar_url = $4, password_hash = $5,
		locale = $6, metadata = $7, app_metadata = $8, status = $9, email_verified = $10, phone_verified = $11, updated_at = $12
		WHERE id = $13 AND tenant_id = $14`,
		user.Email, user.Phone, user.Name, user.AvatarURL, user.PasswordHash,
		user.Locale, user.Metadata, user.AppMetadata, user.Status, user.EmailVerified, user.PhoneVerified,
		user.UpdatedAt, user.ID, user.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func (r *userRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

func (r *userRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams, search string) (*models.PaginatedResult[models.User], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage

	var total int64
	var rows pgx.Rows
	var err error

	if search != "" {
		searchPattern := "%" + search + "%"
		err = r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND (email ILIKE $2 OR name ILIKE $2)`,
			tenantID, searchPattern).Scan(&total)
		if err != nil {
			return nil, fmt.Errorf("counting users: %w", err)
		}
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, email, COALESCE(phone, ''), COALESCE(name, ''), COALESCE(avatar_url, ''), COALESCE(locale, 'en'), metadata, app_metadata, status, email_verified, phone_verified, created_at, updated_at
			FROM users WHERE tenant_id = $1 AND (email ILIKE $2 OR name ILIKE $2)
			ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
			tenantID, searchPattern, params.PerPage, offset)
	} else {
		err = r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE tenant_id = $1`, tenantID).Scan(&total)
		if err != nil {
			return nil, fmt.Errorf("counting users: %w", err)
		}
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, email, COALESCE(phone, ''), COALESCE(name, ''), COALESCE(avatar_url, ''), COALESCE(locale, 'en'), metadata, app_metadata, status, email_verified, phone_verified, created_at, updated_at
			FROM users WHERE tenant_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			tenantID, params.PerPage, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.Phone, &u.Name, &u.AvatarURL,
			&u.Locale, &u.Metadata, &u.AppMetadata, &u.Status, &u.EmailVerified, &u.PhoneVerified,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}

	if users == nil {
		users = []models.User{}
	}

	return &models.PaginatedResult[models.User]{
		Data:       users,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

func (r *userRepo) Block(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET status = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4`,
		models.StatusBlocked, time.Now().UTC(), id, tenantID)
	if err != nil {
		return fmt.Errorf("blocking user: %w", err)
	}
	return nil
}

func (r *userRepo) Unblock(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET status = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4`,
		models.StatusActive, time.Now().UTC(), id, tenantID)
	if err != nil {
		return fmt.Errorf("unblocking user: %w", err)
	}
	return nil
}

func (r *userRepo) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE tenant_id = $1`, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return count, nil
}

func (r *userRepo) GetPasswordHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.PasswordHistory, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, hash, created_at FROM password_history WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("querying password history: %w", err)
	}
	defer rows.Close()

	var history []models.PasswordHistory
	for rows.Next() {
		var ph models.PasswordHistory
		if err := rows.Scan(&ph.ID, &ph.UserID, &ph.Hash, &ph.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning password history: %w", err)
		}
		history = append(history, ph)
	}
	return history, nil
}

func (r *userRepo) AddPasswordHistory(ctx context.Context, entry *models.PasswordHistory) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_history (id, user_id, password_hash, created_at) VALUES ($1, $2, $3, $4)`,
		entry.ID, entry.UserID, entry.Hash, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting password history: %w", err)
	}
	return nil
}
