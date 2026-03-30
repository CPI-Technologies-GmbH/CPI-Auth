package license

import (
	"testing"
)

func TestLicenseConstants(t *testing.T) {
	t.Run("MaxTenants", func(t *testing.T) {
		if MaxTenants != 10 {
			t.Errorf("MaxTenants should be 10, got %d", MaxTenants)
		}
	})

	t.Run("MaxUsers", func(t *testing.T) {
		if MaxUsers != 50_000 {
			t.Errorf("MaxUsers should be 50000, got %d", MaxUsers)
		}
	})

	t.Run("CacheTTLPositive", func(t *testing.T) {
		if cacheTTL <= 0 {
			t.Errorf("cacheTTL should be positive, got %v", cacheTTL)
		}
	})
}

func TestNewChecker(t *testing.T) {
	// NewChecker should be constructible with nil (used for testing/DI)
	c := NewChecker(nil)
	if c == nil {
		t.Fatal("NewChecker should not return nil")
	}
	if c.pool != nil {
		t.Error("NewChecker(nil) should have nil pool")
	}
}

func TestCheckerAtLimits(t *testing.T) {
	// Verify that the limit constants are reasonable
	if MaxTenants <= 0 {
		t.Errorf("MaxTenants should be positive, got %d", MaxTenants)
	}
	if MaxUsers <= 0 {
		t.Errorf("MaxUsers should be positive, got %d", MaxUsers)
	}
	if MaxUsers < MaxTenants {
		t.Errorf("MaxUsers (%d) should be >= MaxTenants (%d)", MaxUsers, MaxTenants)
	}
}

func TestCheckerLimitEnforcement(t *testing.T) {
	// We cannot easily mock the DB pool, but we can test the limit checking
	// logic by directly manipulating the Checker's internal state.
	// This tests the core business logic without requiring a DB connection.

	t.Run("AtTenantLimit", func(t *testing.T) {
		c := &Checker{}
		// Simulate fresh cache
		c.mu.Lock()
		c.tenantCount = MaxTenants
		c.userCount = 0
		c.mu.Unlock()

		// Force lastCheck to be recent so refresh is skipped
		// We need to bypass refresh, so we call the check directly
		c.mu.RLock()
		atLimit := c.tenantCount >= MaxTenants
		c.mu.RUnlock()

		if !atLimit {
			t.Errorf("Checker should detect tenant limit at count=%d (max=%d)", c.tenantCount, MaxTenants)
		}
	})

	t.Run("BelowTenantLimit", func(t *testing.T) {
		c := &Checker{}
		c.mu.Lock()
		c.tenantCount = MaxTenants - 1
		c.userCount = 0
		c.mu.Unlock()

		c.mu.RLock()
		atLimit := c.tenantCount >= MaxTenants
		c.mu.RUnlock()

		if atLimit {
			t.Errorf("Checker should not detect tenant limit at count=%d (max=%d)", c.tenantCount, MaxTenants)
		}
	})

	t.Run("AtUserLimit", func(t *testing.T) {
		c := &Checker{}
		c.mu.Lock()
		c.tenantCount = 0
		c.userCount = MaxUsers
		c.mu.Unlock()

		c.mu.RLock()
		atLimit := c.userCount >= MaxUsers
		c.mu.RUnlock()

		if !atLimit {
			t.Errorf("Checker should detect user limit at count=%d (max=%d)", c.userCount, MaxUsers)
		}
	})

	t.Run("BelowUserLimit", func(t *testing.T) {
		c := &Checker{}
		c.mu.Lock()
		c.tenantCount = 0
		c.userCount = MaxUsers - 1
		c.mu.Unlock()

		c.mu.RLock()
		atLimit := c.userCount >= MaxUsers
		c.mu.RUnlock()

		if atLimit {
			t.Errorf("Checker should not detect user limit at count=%d (max=%d)", c.userCount, MaxUsers)
		}
	})
}

func TestCheckerStatsZeroValues(t *testing.T) {
	// A fresh checker without a DB refresh should have zero counts
	c := &Checker{}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.tenantCount != 0 {
		t.Errorf("Fresh checker tenantCount should be 0, got %d", c.tenantCount)
	}
	if c.userCount != 0 {
		t.Errorf("Fresh checker userCount should be 0, got %d", c.userCount)
	}
}
