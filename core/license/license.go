package license

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// Community license limits
	MaxTenants = 10
	MaxUsers   = 50_000

	// Cache TTL — don't query DB on every request
	cacheTTL = 30 * time.Second
)

// Checker validates license limits at runtime.
type Checker struct {
	pool *pgxpool.Pool

	mu          sync.RWMutex
	tenantCount int
	userCount   int
	lastCheck   time.Time
}

// NewChecker creates a license checker backed by the database.
func NewChecker(pool *pgxpool.Pool) *Checker {
	return &Checker{pool: pool}
}

func (c *Checker) refresh(ctx context.Context) error {
	c.mu.RLock()
	if time.Since(c.lastCheck) < cacheTTL {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(c.lastCheck) < cacheTTL {
		return nil
	}

	var tenants, users int
	if err := c.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tenants").Scan(&tenants); err != nil {
		return err
	}
	if err := c.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE status != 'deleted'").Scan(&users); err != nil {
		return err
	}

	c.tenantCount = tenants
	c.userCount = users
	c.lastCheck = time.Now()
	return nil
}

// CanCreateTenant returns nil if creating a new tenant is allowed.
func (c *Checker) CanCreateTenant(ctx context.Context) error {
	if err := c.refresh(ctx); err != nil {
		return nil // fail open — don't block on DB errors
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.tenantCount >= MaxTenants {
		return fmt.Errorf("community license limit reached: maximum %d tenants (current: %d). Upgrade to a commercial license at https://cpi-technologies.com/auth", MaxTenants, c.tenantCount)
	}
	return nil
}

// CanCreateUser returns nil if creating a new user is allowed.
func (c *Checker) CanCreateUser(ctx context.Context) error {
	if err := c.refresh(ctx); err != nil {
		return nil // fail open
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.userCount >= MaxUsers {
		return fmt.Errorf("community license limit reached: maximum %d users (current: %d). Upgrade to a commercial license at https://cpi-technologies.com/auth", MaxUsers, c.userCount)
	}
	return nil
}

// Stats returns current counts for display.
func (c *Checker) Stats(ctx context.Context) (tenants, users int, err error) {
	if err := c.refresh(ctx); err != nil {
		return 0, 0, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tenantCount, c.userCount, nil
}
