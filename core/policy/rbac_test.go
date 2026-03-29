package policy

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// --- Mock Role Repository ---

type mockRoleRepo struct {
	roles     map[uuid.UUID]*models.Role
	userRoles map[uuid.UUID][]models.Role
	hierarchy map[uuid.UUID][]models.Role
}

func newMockRoleRepo() *mockRoleRepo {
	return &mockRoleRepo{
		roles:     make(map[uuid.UUID]*models.Role),
		userRoles: make(map[uuid.UUID][]models.Role),
		hierarchy: make(map[uuid.UUID][]models.Role),
	}
}

func (m *mockRoleRepo) Create(_ context.Context, role *models.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleRepo) GetByID(_ context.Context, _ uuid.UUID, id uuid.UUID) (*models.Role, error) {
	role, ok := m.roles[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return role, nil
}

func (m *mockRoleRepo) GetByName(_ context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	for _, role := range m.roles {
		if role.TenantID == tenantID && role.Name == name {
			return role, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *mockRoleRepo) Update(_ context.Context, role *models.Role) error {
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleRepo) Delete(_ context.Context, _ uuid.UUID, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

func (m *mockRoleRepo) List(_ context.Context, _ uuid.UUID, _ models.PaginationParams) (*models.PaginatedResult[models.Role], error) {
	var roles []models.Role
	for _, r := range m.roles {
		roles = append(roles, *r)
	}
	return &models.PaginatedResult[models.Role]{Data: roles, Total: int64(len(roles))}, nil
}

func (m *mockRoleRepo) GetRolesForUser(_ context.Context, userID uuid.UUID) ([]models.Role, error) {
	return m.userRoles[userID], nil
}

func (m *mockRoleRepo) GetRoleHierarchy(_ context.Context, roleID uuid.UUID) ([]models.Role, error) {
	return m.hierarchy[roleID], nil
}

func (m *mockRoleRepo) AssignRoleToUser(_ context.Context, userID, roleID, _ uuid.UUID) error {
	return nil
}

func (m *mockRoleRepo) RemoveRoleFromUser(_ context.Context, userID, roleID, _ uuid.UUID) error {
	return nil
}

// --- Tests ---

func TestNewRBACService(t *testing.T) {
	repo := newMockRoleRepo()
	logger := zap.NewNop()
	svc := NewRBACService(repo, logger)

	if svc == nil {
		t.Fatal("NewRBACService returned nil")
	}
}

func TestRBACService_CreateRole(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	tenantID := uuid.New()
	role := &models.Role{
		TenantID:    tenantID,
		Name:        "admin",
		Permissions: []string{"users:read", "users:write"},
	}

	err := svc.CreateRole(context.Background(), role)
	if err != nil {
		t.Fatalf("CreateRole returned error: %v", err)
	}
	if role.ID == uuid.Nil {
		t.Error("role ID should be assigned after creation")
	}
}

func TestRBACService_GetRole(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	tenantID := uuid.New()
	role := &models.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        "editor",
		Permissions: []string{"posts:write"},
	}
	repo.roles[role.ID] = role

	got, err := svc.GetRole(context.Background(), tenantID, role.ID)
	if err != nil {
		t.Fatalf("GetRole returned error: %v", err)
	}
	if got.Name != "editor" {
		t.Errorf("Name = %q, want %q", got.Name, "editor")
	}
}

func TestRBACService_GetRole_NotFound(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	_, err := svc.GetRole(context.Background(), uuid.New(), uuid.New())
	if !models.IsAppError(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestHasPermission(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	userID := uuid.New()
	repo.userRoles[userID] = []models.Role{
		{
			ID:          uuid.New(),
			Name:        "editor",
			Permissions: []string{"posts:read", "posts:write"},
		},
	}

	tests := []struct {
		name       string
		permission string
		want       bool
	}{
		{"has permission", "posts:read", true},
		{"has another permission", "posts:write", true},
		{"missing permission", "users:delete", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, err := svc.HasPermission(context.Background(), userID, tt.permission)
			if err != nil {
				t.Fatalf("HasPermission returned error: %v", err)
			}
			if has != tt.want {
				t.Errorf("HasPermission(%q) = %v, want %v", tt.permission, has, tt.want)
			}
		})
	}
}

func TestHasPermission_Wildcard(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	userID := uuid.New()
	repo.userRoles[userID] = []models.Role{
		{
			ID:          uuid.New(),
			Name:        "superadmin",
			Permissions: []string{"*"},
		},
	}

	has, err := svc.HasPermission(context.Background(), userID, "anything:at:all")
	if err != nil {
		t.Fatalf("HasPermission returned error: %v", err)
	}
	if !has {
		t.Error("wildcard permission should match any permission")
	}
}

func TestHasPermission_NoRoles(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	userID := uuid.New()
	// No roles set for this user

	has, err := svc.HasPermission(context.Background(), userID, "anything")
	if err != nil {
		t.Fatalf("HasPermission returned error: %v", err)
	}
	if has {
		t.Error("user with no roles should not have any permissions")
	}
}

func TestHasAnyPermission(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	userID := uuid.New()
	repo.userRoles[userID] = []models.Role{
		{
			ID:          uuid.New(),
			Name:        "viewer",
			Permissions: []string{"posts:read", "comments:read"},
		},
	}

	tests := []struct {
		name        string
		permissions []string
		want        bool
	}{
		{"has one of them", []string{"posts:read", "users:admin"}, true},
		{"has none", []string{"users:admin", "system:config"}, false},
		{"has both", []string{"posts:read", "comments:read"}, true},
		{"empty list", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, err := svc.HasAnyPermission(context.Background(), userID, tt.permissions)
			if err != nil {
				t.Fatalf("HasAnyPermission returned error: %v", err)
			}
			if has != tt.want {
				t.Errorf("HasAnyPermission(%v) = %v, want %v", tt.permissions, has, tt.want)
			}
		})
	}
}

func TestGetEffectivePermissions_DirectOnly(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	userID := uuid.New()
	repo.userRoles[userID] = []models.Role{
		{
			ID:          uuid.New(),
			Name:        "editor",
			Permissions: []string{"posts:read", "posts:write"},
		},
		{
			ID:          uuid.New(),
			Name:        "commenter",
			Permissions: []string{"comments:read", "comments:write"},
		},
	}

	perms, err := svc.GetEffectivePermissions(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetEffectivePermissions returned error: %v", err)
	}

	permSet := make(map[string]bool)
	for _, p := range perms {
		permSet[p] = true
	}

	expected := []string{"posts:read", "posts:write", "comments:read", "comments:write"}
	for _, exp := range expected {
		if !permSet[exp] {
			t.Errorf("expected permission %q not found in effective permissions", exp)
		}
	}

	if len(perms) != len(expected) {
		t.Errorf("expected %d permissions, got %d", len(expected), len(perms))
	}
}

func TestGetEffectivePermissions_Hierarchy(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	parentRoleID := uuid.New()
	childRoleID := uuid.New()

	// Child role has parent
	childRole := models.Role{
		ID:           childRoleID,
		Name:         "child",
		Permissions:  []string{"child:perm"},
		ParentRoleID: &parentRoleID,
	}

	parentRole := models.Role{
		ID:          parentRoleID,
		Name:        "parent",
		Permissions: []string{"parent:perm"},
	}

	userID := uuid.New()
	repo.userRoles[userID] = []models.Role{childRole}
	repo.hierarchy[childRoleID] = []models.Role{parentRole}

	perms, err := svc.GetEffectivePermissions(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetEffectivePermissions returned error: %v", err)
	}

	permSet := make(map[string]bool)
	for _, p := range perms {
		permSet[p] = true
	}

	if !permSet["child:perm"] {
		t.Error("expected child:perm in effective permissions")
	}
	if !permSet["parent:perm"] {
		t.Error("expected parent:perm (inherited) in effective permissions")
	}
}

func TestGetEffectivePermissions_DeduplicatesPermissions(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	userID := uuid.New()
	repo.userRoles[userID] = []models.Role{
		{
			ID:          uuid.New(),
			Name:        "role1",
			Permissions: []string{"read", "write"},
		},
		{
			ID:          uuid.New(),
			Name:        "role2",
			Permissions: []string{"read", "delete"},
		},
	}

	perms, err := svc.GetEffectivePermissions(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetEffectivePermissions returned error: %v", err)
	}

	// Should deduplicate "read"
	if len(perms) != 3 {
		t.Errorf("expected 3 unique permissions, got %d: %v", len(perms), perms)
	}
}

func TestRBACService_UpdateRole(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	role := &models.Role{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Name:        "editor",
		Permissions: []string{"read"},
	}
	repo.roles[role.ID] = role

	role.Permissions = append(role.Permissions, "write")
	err := svc.UpdateRole(context.Background(), role)
	if err != nil {
		t.Fatalf("UpdateRole returned error: %v", err)
	}

	updated := repo.roles[role.ID]
	if len(updated.Permissions) != 2 {
		t.Errorf("expected 2 permissions after update, got %d", len(updated.Permissions))
	}
}

func TestRBACService_DeleteRole(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	tenantID := uuid.New()
	roleID := uuid.New()
	repo.roles[roleID] = &models.Role{ID: roleID, TenantID: tenantID, Name: "temp"}

	err := svc.DeleteRole(context.Background(), tenantID, roleID)
	if err != nil {
		t.Fatalf("DeleteRole returned error: %v", err)
	}

	if _, ok := repo.roles[roleID]; ok {
		t.Error("role should have been deleted")
	}
}

func TestRBACService_ListRoles(t *testing.T) {
	repo := newMockRoleRepo()
	svc := NewRBACService(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.roles[uuid.New()] = &models.Role{ID: uuid.New(), TenantID: tenantID, Name: "admin"}
	repo.roles[uuid.New()] = &models.Role{ID: uuid.New(), TenantID: tenantID, Name: "editor"}

	result, err := svc.ListRoles(context.Background(), tenantID, models.PaginationParams{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("ListRoles returned error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
}
