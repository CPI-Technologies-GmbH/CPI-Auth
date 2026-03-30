package policy

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// --- Mock FGA Tuple Repository ---

type mockFGATupleRepo struct {
	tuples []models.FGATuple
}

func newMockFGATupleRepo() *mockFGATupleRepo {
	return &mockFGATupleRepo{
		tuples: []models.FGATuple{},
	}
}

func (m *mockFGATupleRepo) Create(_ context.Context, tuple *models.FGATuple) error {
	if tuple.ID == uuid.Nil {
		tuple.ID = uuid.New()
	}
	tuple.CreatedAt = time.Now()
	m.tuples = append(m.tuples, *tuple)
	return nil
}

func (m *mockFGATupleRepo) Delete(_ context.Context, id uuid.UUID) error {
	for i, t := range m.tuples {
		if t.ID == id {
			m.tuples = append(m.tuples[:i], m.tuples[i+1:]...)
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *mockFGATupleRepo) Check(_ context.Context, tenantID uuid.UUID, userType, userID, relation, objectType, objectID string) (bool, error) {
	for _, t := range m.tuples {
		if t.TenantID == tenantID &&
			t.UserType == userType &&
			t.UserID == userID &&
			t.Relation == relation &&
			t.ObjectType == objectType &&
			t.ObjectID == objectID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockFGATupleRepo) ListByObject(_ context.Context, tenantID uuid.UUID, objectType, objectID string) ([]models.FGATuple, error) {
	var result []models.FGATuple
	for _, t := range m.tuples {
		if t.TenantID == tenantID && t.ObjectType == objectType && t.ObjectID == objectID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockFGATupleRepo) ListByUser(_ context.Context, tenantID uuid.UUID, userType, userID string) ([]models.FGATuple, error) {
	var result []models.FGATuple
	for _, t := range m.tuples {
		if t.TenantID == tenantID && t.UserType == userType && t.UserID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}

// --- Tests ---

func TestNewFGAService(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	if svc == nil {
		t.Fatal("NewFGAService returned nil")
	}
}

func TestFGAService_WriteTuple(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenantID := uuid.New()
	tuple := &models.FGATuple{
		TenantID:   tenantID,
		UserType:   "user",
		UserID:     "user-123",
		Relation:   "editor",
		ObjectType: "document",
		ObjectID:   "doc-456",
	}

	err := svc.WriteTuple(context.Background(), tuple)
	if err != nil {
		t.Fatalf("WriteTuple returned error: %v", err)
	}

	if tuple.ID == uuid.Nil {
		t.Error("tuple ID should be assigned after creation")
	}
	if len(repo.tuples) != 1 {
		t.Errorf("expected 1 tuple in repo, got %d", len(repo.tuples))
	}
}

func TestFGAService_Check(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenantID := uuid.New()
	tuple := &models.FGATuple{
		TenantID:   tenantID,
		UserType:   "user",
		UserID:     "user-123",
		Relation:   "viewer",
		ObjectType: "document",
		ObjectID:   "doc-789",
	}
	_ = svc.WriteTuple(context.Background(), tuple)

	tests := []struct {
		name       string
		userType   string
		userID     string
		relation   string
		objectType string
		objectID   string
		want       bool
	}{
		{
			name:       "exact match",
			userType:   "user",
			userID:     "user-123",
			relation:   "viewer",
			objectType: "document",
			objectID:   "doc-789",
			want:       true,
		},
		{
			name:       "wrong user",
			userType:   "user",
			userID:     "user-999",
			relation:   "viewer",
			objectType: "document",
			objectID:   "doc-789",
			want:       false,
		},
		{
			name:       "wrong relation",
			userType:   "user",
			userID:     "user-123",
			relation:   "editor",
			objectType: "document",
			objectID:   "doc-789",
			want:       false,
		},
		{
			name:       "wrong object",
			userType:   "user",
			userID:     "user-123",
			relation:   "viewer",
			objectType: "document",
			objectID:   "doc-000",
			want:       false,
		},
		{
			name:       "wrong object type",
			userType:   "user",
			userID:     "user-123",
			relation:   "viewer",
			objectType: "folder",
			objectID:   "doc-789",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.Check(context.Background(), tenantID, tt.userType, tt.userID, tt.relation, tt.objectType, tt.objectID)
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if result != tt.want {
				t.Errorf("Check = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestFGAService_Check_Cache(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenantID := uuid.New()
	tuple := &models.FGATuple{
		TenantID:   tenantID,
		UserType:   "user",
		UserID:     "user-abc",
		Relation:   "owner",
		ObjectType: "project",
		ObjectID:   "proj-1",
	}
	_ = svc.WriteTuple(context.Background(), tuple)

	// First call should query the repo and cache
	result1, err := svc.Check(context.Background(), tenantID, "user", "user-abc", "owner", "project", "proj-1")
	if err != nil {
		t.Fatalf("first Check returned error: %v", err)
	}
	if !result1 {
		t.Error("first Check should return true")
	}

	// Second call should use cache (same result)
	result2, err := svc.Check(context.Background(), tenantID, "user", "user-abc", "owner", "project", "proj-1")
	if err != nil {
		t.Fatalf("second Check returned error: %v", err)
	}
	if !result2 {
		t.Error("second Check (cached) should return true")
	}
}

func TestFGAService_WriteTuple_InvalidatesCache(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenantID := uuid.New()

	// Check before tuple exists - should be false and cached
	result, err := svc.Check(context.Background(), tenantID, "user", "u1", "reader", "doc", "d1")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result {
		t.Error("Check should return false before tuple exists")
	}

	// Write the tuple - should invalidate cache
	tuple := &models.FGATuple{
		TenantID:   tenantID,
		UserType:   "user",
		UserID:     "u1",
		Relation:   "reader",
		ObjectType: "doc",
		ObjectID:   "d1",
	}
	_ = svc.WriteTuple(context.Background(), tuple)

	// Check again - should now return true (cache invalidated)
	result, err = svc.Check(context.Background(), tenantID, "user", "u1", "reader", "doc", "d1")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !result {
		t.Error("Check should return true after tuple is written and cache is invalidated")
	}
}

func TestFGAService_DeleteTuple(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenantID := uuid.New()
	tuple := &models.FGATuple{
		TenantID:   tenantID,
		UserType:   "user",
		UserID:     "user-1",
		Relation:   "admin",
		ObjectType: "org",
		ObjectID:   "org-1",
	}
	_ = svc.WriteTuple(context.Background(), tuple)

	err := svc.DeleteTuple(context.Background(), tuple.ID)
	if err != nil {
		t.Fatalf("DeleteTuple returned error: %v", err)
	}

	if len(repo.tuples) != 0 {
		t.Errorf("expected 0 tuples after delete, got %d", len(repo.tuples))
	}
}

func TestFGAService_DeleteTuple_NotFound(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	err := svc.DeleteTuple(context.Background(), uuid.New())
	if !models.IsAppError(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFGAService_ListByObject(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenantID := uuid.New()
	// Add two tuples for the same object
	_ = svc.WriteTuple(context.Background(), &models.FGATuple{
		TenantID: tenantID, UserType: "user", UserID: "u1", Relation: "viewer", ObjectType: "doc", ObjectID: "d1",
	})
	_ = svc.WriteTuple(context.Background(), &models.FGATuple{
		TenantID: tenantID, UserType: "user", UserID: "u2", Relation: "editor", ObjectType: "doc", ObjectID: "d1",
	})
	// Add one tuple for a different object
	_ = svc.WriteTuple(context.Background(), &models.FGATuple{
		TenantID: tenantID, UserType: "user", UserID: "u1", Relation: "viewer", ObjectType: "doc", ObjectID: "d2",
	})

	tuples, err := svc.ListByObject(context.Background(), tenantID, "doc", "d1")
	if err != nil {
		t.Fatalf("ListByObject returned error: %v", err)
	}
	if len(tuples) != 2 {
		t.Errorf("expected 2 tuples for doc:d1, got %d", len(tuples))
	}
}

func TestFGAService_ListByUser(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenantID := uuid.New()
	_ = svc.WriteTuple(context.Background(), &models.FGATuple{
		TenantID: tenantID, UserType: "user", UserID: "u1", Relation: "viewer", ObjectType: "doc", ObjectID: "d1",
	})
	_ = svc.WriteTuple(context.Background(), &models.FGATuple{
		TenantID: tenantID, UserType: "user", UserID: "u1", Relation: "editor", ObjectType: "doc", ObjectID: "d2",
	})
	_ = svc.WriteTuple(context.Background(), &models.FGATuple{
		TenantID: tenantID, UserType: "user", UserID: "u2", Relation: "viewer", ObjectType: "doc", ObjectID: "d1",
	})

	tuples, err := svc.ListByUser(context.Background(), tenantID, "user", "u1")
	if err != nil {
		t.Fatalf("ListByUser returned error: %v", err)
	}
	if len(tuples) != 2 {
		t.Errorf("expected 2 tuples for user:u1, got %d", len(tuples))
	}
}

func TestFGAService_Check_DifferentTenants(t *testing.T) {
	repo := newMockFGATupleRepo()
	svc := NewFGAService(repo, zap.NewNop())

	tenant1 := uuid.New()
	tenant2 := uuid.New()

	_ = svc.WriteTuple(context.Background(), &models.FGATuple{
		TenantID: tenant1, UserType: "user", UserID: "u1", Relation: "admin", ObjectType: "system", ObjectID: "s1",
	})

	// Same tuple params but different tenant should not match
	result, err := svc.Check(context.Background(), tenant2, "user", "u1", "admin", "system", "s1")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result {
		t.Error("Check should return false for a different tenant")
	}
}

// --- Cache unit tests ---

func TestFGACache_GetSet(t *testing.T) {
	c := newFGACache(1 * time.Minute)

	// Miss
	_, ok := c.Get("key1")
	if ok {
		t.Error("Get should return false for missing key")
	}

	// Set and hit
	c.Set("key1", true)
	val, ok := c.Get("key1")
	if !ok {
		t.Error("Get should return true for existing key")
	}
	if !val {
		t.Error("cached value should be true")
	}
}

func TestFGACache_Invalidate(t *testing.T) {
	c := newFGACache(1 * time.Minute)

	c.Set("key1", true)
	c.Invalidate("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("Get should return false after Invalidate")
	}
}

func TestFGACache_Expiration(t *testing.T) {
	c := newFGACache(1 * time.Millisecond)

	c.Set("key1", true)

	// Wait for expiry
	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("key1")
	if ok {
		t.Error("Get should return false for expired cache entry")
	}
}
