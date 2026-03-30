package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Bootstrap configures the instance with runtime settings on every startup.
// This ensures Helm-deployed instances work without manual DB edits.
func Bootstrap(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	bootstrapTenantDomain(ctx, pool, logger)
	bootstrapAdminPassword(ctx, pool, logger)
	bootstrapTenantName(ctx, pool, logger)
}

// bootstrapTenantDomain sets the default tenant's domain from AF_PUBLIC_DOMAIN.
func bootstrapTenantDomain(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	domain := os.Getenv("AF_PUBLIC_DOMAIN")
	if domain == "" {
		return
	}
	_, err := pool.Exec(ctx, `
		UPDATE tenants SET domain = $1, updated_at = NOW()
		WHERE id = (SELECT id FROM tenants ORDER BY created_at ASC LIMIT 1)
		AND (domain IS NULL OR domain = '' OR domain != $1)
	`, domain)
	if err != nil {
		logger.Warn("failed to set default tenant domain", zap.Error(err))
		return
	}
	logger.Info("bootstrap: tenant domain configured", zap.String("domain", domain))
}

// bootstrapAdminPassword sets the admin password from AF_ADMIN_PASSWORD on first run.
// Only applies if the current password is the default seed password.
func bootstrapAdminPassword(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	newPassword := os.Getenv("AF_ADMIN_PASSWORD")
	if newPassword == "" {
		return
	}

	// Only update if current hash matches the seed default
	var currentHash string
	err := pool.QueryRow(ctx, `
		SELECT password_hash FROM users
		WHERE email = (SELECT email FROM users ORDER BY created_at ASC LIMIT 1)
	`).Scan(&currentHash)
	if err != nil {
		return
	}

	// Check if current password is still the default "admin123!"
	if bcrypt.CompareHashAndPassword([]byte(currentHash), []byte("admin123!")) != nil {
		// Password already changed — don't override
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		logger.Warn("failed to hash admin password", zap.Error(err))
		return
	}

	_, err = pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = NOW()
		WHERE id = (SELECT id FROM users ORDER BY created_at ASC LIMIT 1)
	`, string(hash))
	if err != nil {
		logger.Warn("failed to set admin password", zap.Error(err))
		return
	}
	logger.Info("bootstrap: admin password updated from AF_ADMIN_PASSWORD")
}

// bootstrapTenantName sets the default tenant name from AF_TENANT_NAME.
func bootstrapTenantName(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	name := os.Getenv("AF_TENANT_NAME")
	if name == "" {
		return
	}
	pool.Exec(ctx, `
		UPDATE tenants SET name = $1, updated_at = NOW()
		WHERE id = (SELECT id FROM tenants ORDER BY created_at ASC LIMIT 1)
		AND name = 'Default Tenant'
	`, name)
	logger.Info("bootstrap: tenant name configured", zap.String("name", name))
}
