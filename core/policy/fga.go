package policy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// FGAService handles fine-grained authorization using the Zanzibar (ReBAC) model.
type FGAService struct {
	tuples models.FGATupleRepository
	cache  *fgaCache
	logger *zap.Logger
}

// fgaCache provides sub-ms latency for repeated checks.
type fgaCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	result    bool
	expiresAt time.Time
}

func newFGACache(ttl time.Duration) *fgaCache {
	return &fgaCache{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

func (c *fgaCache) Get(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return false, false
	}
	return entry.result, true
}

func (c *fgaCache) Set(key string, result bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *fgaCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// NewFGAService creates a new FGA service.
func NewFGAService(tuples models.FGATupleRepository, logger *zap.Logger) *FGAService {
	return &FGAService{
		tuples: tuples,
		cache:  newFGACache(30 * time.Second),
		logger: logger,
	}
}

// WriteTuple creates a new authorization tuple.
func (s *FGAService) WriteTuple(ctx context.Context, tuple *models.FGATuple) error {
	if err := s.tuples.Create(ctx, tuple); err != nil {
		return err
	}
	// Invalidate cache for this check
	cacheKey := s.cacheKey(tuple.TenantID, tuple.UserType, tuple.UserID, tuple.Relation, tuple.ObjectType, tuple.ObjectID)
	s.cache.Invalidate(cacheKey)
	return nil
}

// DeleteTuple removes an authorization tuple.
func (s *FGAService) DeleteTuple(ctx context.Context, id uuid.UUID) error {
	return s.tuples.Delete(ctx, id)
}

// Check performs a fine-grained authorization check: can user X do Y on object Z?
func (s *FGAService) Check(ctx context.Context, tenantID uuid.UUID, userType, userID, relation, objectType, objectID string) (bool, error) {
	cacheKey := s.cacheKey(tenantID, userType, userID, relation, objectType, objectID)

	// Check cache first for sub-ms latency
	if result, ok := s.cache.Get(cacheKey); ok {
		return result, nil
	}

	// Query database
	result, err := s.tuples.Check(ctx, tenantID, userType, userID, relation, objectType, objectID)
	if err != nil {
		return false, err
	}

	// Cache the result
	s.cache.Set(cacheKey, result)
	return result, nil
}

// ListByObject returns all tuples for a given object.
func (s *FGAService) ListByObject(ctx context.Context, tenantID uuid.UUID, objectType, objectID string) ([]models.FGATuple, error) {
	return s.tuples.ListByObject(ctx, tenantID, objectType, objectID)
}

// ListByUser returns all tuples for a given user.
func (s *FGAService) ListByUser(ctx context.Context, tenantID uuid.UUID, userType, userID string) ([]models.FGATuple, error) {
	return s.tuples.ListByUser(ctx, tenantID, userType, userID)
}

func (s *FGAService) cacheKey(tenantID uuid.UUID, userType, userID, relation, objectType, objectID string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", tenantID, userType, userID, relation, objectType, objectID)
}
