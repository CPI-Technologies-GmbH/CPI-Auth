package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// AutoMigrate runs all .up.sql migrations in order.
// It tracks applied migrations in a schema_migrations table.
func AutoMigrate(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) error {
	// Create tracking table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	// Read all migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Get already applied
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("querying applied migrations: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		rows.Scan(&v)
		applied[v] = true
	}
	rows.Close()

	// Apply pending
	for _, file := range files {
		if applied[file] {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", file, err)
		}

		logger.Info("applying migration", zap.String("file", file))
		_, err = pool.Exec(ctx, string(data))
		if err != nil {
			// Log but don't fail on non-critical errors (e.g. "already exists")
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate key") {
				logger.Warn("migration had conflicts (skipped)", zap.String("file", file), zap.Error(err))
			} else {
				return fmt.Errorf("applying migration %s: %w", file, err)
			}
		}

		pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", file)
	}

	logger.Info("migrations complete", zap.Int("total", len(files)), zap.Int("applied", len(files)-len(applied)))
	return nil
}
