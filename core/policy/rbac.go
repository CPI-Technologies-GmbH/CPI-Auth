package policy

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// RBACService handles role-based access control with hierarchical roles.
type RBACService struct {
	roles  models.RoleRepository
	logger *zap.Logger
}

// NewRBACService creates a new RBAC service.
func NewRBACService(roles models.RoleRepository, logger *zap.Logger) *RBACService {
	return &RBACService{
		roles:  roles,
		logger: logger,
	}
}

// CreateRole creates a new role.
func (s *RBACService) CreateRole(ctx context.Context, role *models.Role) error {
	return s.roles.Create(ctx, role)
}

// GetRole retrieves a role by ID.
func (s *RBACService) GetRole(ctx context.Context, tenantID, roleID uuid.UUID) (*models.Role, error) {
	return s.roles.GetByID(ctx, tenantID, roleID)
}

// UpdateRole updates an existing role.
func (s *RBACService) UpdateRole(ctx context.Context, role *models.Role) error {
	return s.roles.Update(ctx, role)
}

// DeleteRole removes a role.
func (s *RBACService) DeleteRole(ctx context.Context, tenantID, roleID uuid.UUID) error {
	return s.roles.Delete(ctx, tenantID, roleID)
}

// ListRoles returns all roles for a tenant.
func (s *RBACService) ListRoles(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Role], error) {
	return s.roles.List(ctx, tenantID, params)
}

// GetEffectivePermissions returns all permissions for a user including inherited ones.
func (s *RBACService) GetEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	roles, err := s.roles.GetRolesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]bool)
	for _, role := range roles {
		// Get direct permissions
		for _, perm := range role.Permissions {
			permSet[perm] = true
		}

		// Get inherited permissions from parent roles
		if role.ParentRoleID != nil {
			hierarchy, err := s.roles.GetRoleHierarchy(ctx, role.ID)
			if err != nil {
				s.logger.Warn("failed to get role hierarchy",
					zap.String("role_id", role.ID.String()),
					zap.Error(err),
				)
				continue
			}
			for _, parentRole := range hierarchy {
				for _, perm := range parentRole.Permissions {
					permSet[perm] = true
				}
			}
		}
	}

	perms := make([]string, 0, len(permSet))
	for perm := range permSet {
		perms = append(perms, perm)
	}
	return perms, nil
}

// HasPermission checks if a user has a specific permission.
func (s *RBACService) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	perms, err := s.GetEffectivePermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p == permission || p == "*" {
			return true, nil
		}
	}
	return false, nil
}

// HasAnyPermission checks if a user has any of the given permissions.
func (s *RBACService) HasAnyPermission(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error) {
	for _, perm := range permissions {
		has, err := s.HasPermission(ctx, userID, perm)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}
