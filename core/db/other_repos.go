package db

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// nilIfEmpty returns nil if s is empty, otherwise returns &s.
// Used for nullable inet columns that can't accept empty strings.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// --- Identity Repository ---

type identityRepo struct{ pool *pgxpool.Pool }

func NewIdentityRepository(pool *pgxpool.Pool) models.IdentityRepository {
	return &identityRepo{pool: pool}
}

func (r *identityRepo) Create(ctx context.Context, identity *models.Identity) error {
	identity.ID = uuid.New()
	now := time.Now().UTC()
	identity.CreatedAt = now
	identity.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO identities (id, user_id, provider, provider_user_id, tokens_encrypted, profile, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		identity.ID, identity.UserID, identity.Provider, identity.ProviderUserID,
		identity.TokensEncrypted, identity.Profile, identity.CreatedAt, identity.UpdatedAt)
	return err
}

func (r *identityRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Identity, error) {
	var i models.Identity
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, provider_user_id, tokens_encrypted, profile, created_at, updated_at
		FROM identities WHERE id = $1`, id).
		Scan(&i.ID, &i.UserID, &i.Provider, &i.ProviderUserID, &i.TokensEncrypted, &i.Profile, &i.CreatedAt, &i.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &i, err
}

func (r *identityRepo) GetByProvider(ctx context.Context, provider, providerUserID string) (*models.Identity, error) {
	var i models.Identity
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, provider_user_id, tokens_encrypted, profile, created_at, updated_at
		FROM identities WHERE provider = $1 AND provider_user_id = $2`, provider, providerUserID).
		Scan(&i.ID, &i.UserID, &i.Provider, &i.ProviderUserID, &i.TokensEncrypted, &i.Profile, &i.CreatedAt, &i.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &i, err
}

func (r *identityRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Identity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, provider, provider_user_id, profile, created_at, updated_at
		FROM identities WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var identities []models.Identity
	for rows.Next() {
		var i models.Identity
		if err := rows.Scan(&i.ID, &i.UserID, &i.Provider, &i.ProviderUserID, &i.Profile, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		identities = append(identities, i)
	}
	return identities, nil
}

func (r *identityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM identities WHERE id = $1`, id)
	return err
}

func (r *identityRepo) Update(ctx context.Context, identity *models.Identity) error {
	identity.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE identities SET tokens_encrypted = $1, profile = $2, updated_at = $3 WHERE id = $4`,
		identity.TokensEncrypted, identity.Profile, identity.UpdatedAt, identity.ID)
	return err
}

// --- Organization Repository ---

type orgRepo struct{ pool *pgxpool.Pool }

func NewOrganizationRepository(pool *pgxpool.Pool) models.OrganizationRepository {
	return &orgRepo{pool: pool}
}

func (r *orgRepo) Create(ctx context.Context, org *models.Organization) error {
	org.ID = uuid.New()
	now := time.Now().UTC()
	org.CreatedAt = now
	org.UpdatedAt = now
	if org.Domains == nil {
		org.Domains = []string{}
	}
	if org.Settings == nil {
		org.Settings = json.RawMessage(`{}`)
	}
	if org.Branding == nil {
		org.Branding = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO organizations (id, tenant_id, name, slug, domains, settings, branding, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		org.ID, org.TenantID, org.Name, org.Slug, org.Domains, org.Settings, org.Branding, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *orgRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Organization, error) {
	var o models.Organization
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, slug, domains, settings, branding, created_at, updated_at
		FROM organizations WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&o.ID, &o.TenantID, &o.Name, &o.Slug, &o.Domains, &o.Settings, &o.Branding, &o.CreatedAt, &o.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &o, err
}

func (r *orgRepo) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*models.Organization, error) {
	var o models.Organization
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, slug, domains, settings, branding, created_at, updated_at
		FROM organizations WHERE slug = $1 AND tenant_id = $2`, slug, tenantID).
		Scan(&o.ID, &o.TenantID, &o.Name, &o.Slug, &o.Domains, &o.Settings, &o.Branding, &o.CreatedAt, &o.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &o, err
}

func (r *orgRepo) Update(ctx context.Context, org *models.Organization) error {
	org.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE organizations SET name = $1, slug = $2, domains = $3, settings = $4, branding = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8`,
		org.Name, org.Slug, org.Domains, org.Settings, org.Branding, org.UpdatedAt, org.ID, org.TenantID)
	return err
}

func (r *orgRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM organizations WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *orgRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Organization], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM organizations WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, slug, domains, settings, branding, created_at, updated_at
		FROM organizations WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orgs []models.Organization
	for rows.Next() {
		var o models.Organization
		if err := rows.Scan(&o.ID, &o.TenantID, &o.Name, &o.Slug, &o.Domains, &o.Settings, &o.Branding, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	if orgs == nil {
		orgs = []models.Organization{}
	}
	return &models.PaginatedResult[models.Organization]{
		Data: orgs, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

func (r *orgRepo) AddMember(ctx context.Context, member *models.OrganizationMember) error {
	member.CreatedAt = time.Now().UTC()
	if member.Role == "" {
		member.Role = "member"
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role, created_at) VALUES ($1, $2, $3, $4)`,
		member.OrgID, member.UserID, member.Role, member.CreatedAt)
	return err
}

func (r *orgRepo) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`, orgID, userID)
	return err
}

func (r *orgRepo) ListMembers(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.OrganizationMember], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM organization_members WHERE organization_id = $1`, orgID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT organization_id, user_id, role, created_at FROM organization_members WHERE organization_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		orgID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []models.OrganizationMember
	for rows.Next() {
		var m models.OrganizationMember
		if err := rows.Scan(&m.OrgID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if members == nil {
		members = []models.OrganizationMember{}
	}
	return &models.PaginatedResult[models.OrganizationMember]{
		Data: members, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

func (r *orgRepo) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error) {
	var m models.OrganizationMember
	err := r.pool.QueryRow(ctx, `
		SELECT organization_id, user_id, role, created_at FROM organization_members WHERE organization_id = $1 AND user_id = $2`,
		orgID, userID).Scan(&m.OrgID, &m.UserID, &m.Role, &m.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &m, err
}

// --- Role Repository ---

type roleRepo struct{ pool *pgxpool.Pool }

func NewRoleRepository(pool *pgxpool.Pool) models.RoleRepository {
	return &roleRepo{pool: pool}
}

func (r *roleRepo) Create(ctx context.Context, role *models.Role) error {
	role.ID = uuid.New()
	now := time.Now().UTC()
	role.CreatedAt = now
	role.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO roles (id, tenant_id, name, description, is_system, permissions, parent_role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		role.ID, role.TenantID, role.Name, role.Description, role.IsSystem, role.Permissions, role.ParentRoleID, role.CreatedAt, role.UpdatedAt)
	return err
}

func (r *roleRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, COALESCE(description, ''), is_system, permissions, parent_role_id, created_at, updated_at
		FROM roles WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.Permissions, &role.ParentRoleID, &role.CreatedAt, &role.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &role, err
}

func (r *roleRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	var role models.Role
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, COALESCE(description, ''), is_system, permissions, parent_role_id, created_at, updated_at
		FROM roles WHERE name = $1 AND tenant_id = $2`, name, tenantID).
		Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.Permissions, &role.ParentRoleID, &role.CreatedAt, &role.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &role, err
}

func (r *roleRepo) Update(ctx context.Context, role *models.Role) error {
	role.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE roles SET name = $1, description = $2, is_system = $3, permissions = $4, parent_role_id = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8`,
		role.Name, role.Description, role.IsSystem, role.Permissions, role.ParentRoleID, role.UpdatedAt, role.ID, role.TenantID)
	return err
}

func (r *roleRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *roleRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Role], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM roles WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, COALESCE(description, ''), is_system, permissions, parent_role_id, created_at, updated_at
		FROM roles WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.Permissions, &role.ParentRoleID, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if roles == nil {
		roles = []models.Role{}
	}
	return &models.PaginatedResult[models.Role]{
		Data: roles, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

func (r *roleRepo) GetRolesForUser(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.tenant_id, r.name, COALESCE(r.description, ''), r.is_system, r.permissions, r.parent_role_id, r.created_at, r.updated_at
		FROM roles r INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.Permissions, &role.ParentRoleID, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *roleRepo) GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]models.Role, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE role_tree AS (
			SELECT id, tenant_id, name, COALESCE(description, ''), is_system, permissions, parent_role_id, created_at, updated_at FROM roles WHERE id = $1
			UNION ALL
			SELECT r.id, r.tenant_id, r.name, COALESCE(r.description, ''), r.is_system, r.permissions, r.parent_role_id, r.created_at, r.updated_at
			FROM roles r INNER JOIN role_tree rt ON r.id = rt.parent_role_id
		)
		SELECT id, tenant_id, name, description, is_system, permissions, parent_role_id, created_at, updated_at FROM role_tree`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsSystem, &role.Permissions, &role.ParentRoleID, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *roleRepo) AssignRoleToUser(ctx context.Context, userID, roleID, organizationID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, organization_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`,
		userID, roleID, organizationID)
	return err
}

func (r *roleRepo) RemoveRoleFromUser(ctx context.Context, userID, roleID, organizationID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2 AND organization_id = $3`,
		userID, roleID, organizationID)
	return err
}

// --- Permission Repository ---

type permissionRepo struct{ pool *pgxpool.Pool }

func NewPermissionRepository(pool *pgxpool.Pool) models.PermissionRepository {
	return &permissionRepo{pool: pool}
}

// systemPermissions defines the default permissions seeded for each tenant.
var systemPermissions = []struct {
	Name        string
	DisplayName string
	Description string
	GroupName   string
}{
	{"users:read", "Read Users", "View user profiles and details", "Users"},
	{"users:write", "Write Users", "Create, update, and delete users", "Users"},
	{"users:block", "Block Users", "Block and unblock user accounts", "Users"},
	{"applications:read", "Read Applications", "View application configurations", "Applications"},
	{"applications:write", "Write Applications", "Create, update, and delete applications", "Applications"},
	{"organizations:read", "Read Organizations", "View organizations and members", "Organizations"},
	{"organizations:write", "Write Organizations", "Create, update, and delete organizations", "Organizations"},
	{"roles:read", "Read Roles", "View roles and permissions", "Roles"},
	{"roles:write", "Write Roles", "Create, update, and delete roles", "Roles"},
	{"tenants:read", "Read Tenants", "View tenant configurations", "Tenants"},
	{"tenants:write", "Write Tenants", "Create, update, and delete tenants", "Tenants"},
	{"webhooks:read", "Read Webhooks", "View webhook configurations", "Webhooks"},
	{"webhooks:write", "Write Webhooks", "Create, update, and delete webhooks", "Webhooks"},
	{"logs:read", "Read Logs", "View audit logs", "Logs"},
	{"settings:read", "Read Settings", "View platform settings", "Settings"},
	{"settings:write", "Write Settings", "Modify platform settings", "Settings"},
	{"admin:access", "Admin Access", "Access admin dashboard", "Admin"},
	{"*", "Super Admin", "Full access to all resources", "Admin"},
}

func (r *permissionRepo) Create(ctx context.Context, perm *models.Permission) error {
	perm.ID = uuid.New()
	now := time.Now().UTC()
	perm.CreatedAt = now
	perm.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO permissions (id, tenant_id, name, display_name, description, group_name, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		perm.ID, perm.TenantID, perm.Name, perm.DisplayName, perm.Description, perm.GroupName, perm.IsSystem, perm.CreatedAt, perm.UpdatedAt)
	return err
}

func (r *permissionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Permission, error) {
	var p models.Permission
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, display_name, description, group_name, is_system, created_at, updated_at
		FROM permissions WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&p.ID, &p.TenantID, &p.Name, &p.DisplayName, &p.Description, &p.GroupName, &p.IsSystem, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &p, err
}

func (r *permissionRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Permission, error) {
	var p models.Permission
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, display_name, description, group_name, is_system, created_at, updated_at
		FROM permissions WHERE name = $1 AND tenant_id = $2`, name, tenantID).
		Scan(&p.ID, &p.TenantID, &p.Name, &p.DisplayName, &p.Description, &p.GroupName, &p.IsSystem, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &p, err
}

func (r *permissionRepo) Update(ctx context.Context, perm *models.Permission) error {
	perm.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE permissions SET name = $1, display_name = $2, description = $3, group_name = $4, is_system = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8`,
		perm.Name, perm.DisplayName, perm.Description, perm.GroupName, perm.IsSystem, perm.UpdatedAt, perm.ID, perm.TenantID)
	return err
}

func (r *permissionRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM permissions WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *permissionRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Permission], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 100
	}
	offset := (params.Page - 1) * params.PerPage
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM permissions WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, display_name, description, group_name, is_system, created_at, updated_at
		FROM permissions WHERE tenant_id = $1 ORDER BY group_name, name LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.DisplayName, &p.Description, &p.GroupName, &p.IsSystem, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	if perms == nil {
		perms = []models.Permission{}
	}
	return &models.PaginatedResult[models.Permission]{
		Data: perms, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

func (r *permissionRepo) ListAll(ctx context.Context, tenantID uuid.UUID) ([]models.Permission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, display_name, description, group_name, is_system, created_at, updated_at
		FROM permissions WHERE tenant_id = $1 ORDER BY group_name, name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.DisplayName, &p.Description, &p.GroupName, &p.IsSystem, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *permissionRepo) EnsureSystemDefaults(ctx context.Context, tenantID uuid.UUID) error {
	for _, sp := range systemPermissions {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO permissions (id, tenant_id, name, display_name, description, group_name, is_system, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW(), NOW())
			ON CONFLICT (tenant_id, name) DO NOTHING`,
			uuid.New(), tenantID, sp.Name, sp.DisplayName, sp.Description, sp.GroupName)
		if err != nil {
			return err
		}
	}
	return nil
}

// --- Application Permission Repository ---

type appPermissionRepo struct{ pool *pgxpool.Pool }

func NewApplicationPermissionRepository(pool *pgxpool.Pool) models.ApplicationPermissionRepository {
	return &appPermissionRepo{pool: pool}
}

func (r *appPermissionRepo) SetPermissions(ctx context.Context, appID, tenantID uuid.UUID, permissions []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM application_permissions WHERE application_id = $1`, appID)
	if err != nil {
		return err
	}

	for _, perm := range permissions {
		_, err = tx.Exec(ctx, `
			INSERT INTO application_permissions (id, application_id, tenant_id, permission_name, created_at)
			VALUES ($1, $2, $3, $4, NOW())`,
			uuid.New(), appID, tenantID, perm)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *appPermissionRepo) GetPermissions(ctx context.Context, appID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT permission_name FROM application_permissions WHERE application_id = $1 ORDER BY permission_name`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, nil
}

// --- OAuth Grant Repository ---

type oauthGrantRepo struct{ pool *pgxpool.Pool }

func NewOAuthGrantRepository(pool *pgxpool.Pool) models.OAuthGrantRepository {
	return &oauthGrantRepo{pool: pool}
}

func (r *oauthGrantRepo) Create(ctx context.Context, grant *models.OAuthGrant) error {
	grant.ID = uuid.New()
	grant.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO oauth_grants (id, user_id, application_id, tenant_id, scopes, code, code_challenge, code_challenge_method, redirect_uri, nonce, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		grant.ID, grant.UserID, grant.ApplicationID, grant.TenantID, grant.Scopes, grant.Code,
		grant.CodeChallenge, grant.CodeChallengeMethod, grant.RedirectURI, grant.Nonce, grant.ExpiresAt, grant.CreatedAt)
	return err
}

func (r *oauthGrantRepo) GetByCode(ctx context.Context, code string) (*models.OAuthGrant, error) {
	var g models.OAuthGrant
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, application_id, scopes, code, COALESCE(code_challenge, ''), COALESCE(code_challenge_method, ''), redirect_uri, COALESCE(nonce, ''), expires_at, created_at
		FROM oauth_grants WHERE code = $1`, code).
		Scan(&g.ID, &g.UserID, &g.ApplicationID, &g.Scopes, &g.Code,
			&g.CodeChallenge, &g.CodeChallengeMethod, &g.RedirectURI, &g.Nonce, &g.ExpiresAt, &g.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &g, err
}

func (r *oauthGrantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM oauth_grants WHERE id = $1`, id)
	return err
}

func (r *oauthGrantRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM oauth_grants WHERE expires_at < NOW()`)
	return err
}

// --- Refresh Token Repository ---

type refreshTokenRepo struct{ pool *pgxpool.Pool }

func NewRefreshTokenRepository(pool *pgxpool.Pool) models.RefreshTokenRepository {
	return &refreshTokenRepo{pool: pool}
}

func (r *refreshTokenRepo) Create(ctx context.Context, token *models.RefreshToken) error {
	token.ID = uuid.New()
	token.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, application_id, tenant_id, token_hash, family, revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		token.ID, token.UserID, token.ApplicationID, token.TenantID, token.TokenHash, token.Family, token.Revoked, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *refreshTokenRepo) GetByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, application_id, tenant_id, token_hash, family, revoked, expires_at, created_at
		FROM refresh_tokens WHERE token_hash = $1`, hash).
		Scan(&t.ID, &t.UserID, &t.ApplicationID, &t.TenantID, &t.TokenHash, &t.Family, &t.Revoked, &t.ExpiresAt, &t.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &t, err
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE id = $1`, id)
	return err
}

func (r *refreshTokenRepo) RevokeByFamily(ctx context.Context, family string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE family = $1`, family)
	return err
}

func (r *refreshTokenRepo) RevokeByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userID)
	return err
}

func (r *refreshTokenRepo) RevokeByApplication(ctx context.Context, appID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE application_id = $1`, appID)
	return err
}

func (r *refreshTokenRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	return err
}

// --- MFA Enrollment Repository ---

type mfaRepo struct{ pool *pgxpool.Pool }

func NewMFAEnrollmentRepository(pool *pgxpool.Pool) models.MFAEnrollmentRepository {
	return &mfaRepo{pool: pool}
}

func (r *mfaRepo) Create(ctx context.Context, enrollment *models.MFAEnrollment) error {
	enrollment.ID = uuid.New()
	enrollment.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO mfa_enrollments (id, user_id, method, secret_encrypted, verified, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		enrollment.ID, enrollment.UserID, enrollment.Method, enrollment.SecretEncrypted, enrollment.Verified, enrollment.CreatedAt)
	return err
}

func (r *mfaRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.MFAEnrollment, error) {
	var e models.MFAEnrollment
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, method, secret_encrypted, verified, created_at
		FROM mfa_enrollments WHERE id = $1`, id).
		Scan(&e.ID, &e.UserID, &e.Method, &e.SecretEncrypted, &e.Verified, &e.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &e, err
}

func (r *mfaRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.MFAEnrollment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, method, secret_encrypted, verified, created_at
		FROM mfa_enrollments WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var enrollments []models.MFAEnrollment
	for rows.Next() {
		var e models.MFAEnrollment
		if err := rows.Scan(&e.ID, &e.UserID, &e.Method, &e.SecretEncrypted, &e.Verified, &e.CreatedAt); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, e)
	}
	return enrollments, nil
}

func (r *mfaRepo) Update(ctx context.Context, enrollment *models.MFAEnrollment) error {
	_, err := r.pool.Exec(ctx, `UPDATE mfa_enrollments SET verified = $1, secret_encrypted = $2 WHERE id = $3`,
		enrollment.Verified, enrollment.SecretEncrypted, enrollment.ID)
	return err
}

func (r *mfaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM mfa_enrollments WHERE id = $1`, id)
	return err
}

// --- Recovery Code Repository ---

type recoveryCodeRepo struct{ pool *pgxpool.Pool }

func NewRecoveryCodeRepository(pool *pgxpool.Pool) models.RecoveryCodeRepository {
	return &recoveryCodeRepo{pool: pool}
}

func (r *recoveryCodeRepo) Create(ctx context.Context, code *models.RecoveryCode) error {
	code.ID = uuid.New()
	_, err := r.pool.Exec(ctx, `INSERT INTO recovery_codes (id, user_id, code_hash, used) VALUES ($1, $2, $3, $4)`,
		code.ID, code.UserID, code.CodeHash, code.Used)
	return err
}

func (r *recoveryCodeRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.RecoveryCode, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, user_id, code_hash, used FROM recovery_codes WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []models.RecoveryCode
	for rows.Next() {
		var c models.RecoveryCode
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &c.Used); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, nil
}

func (r *recoveryCodeRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE recovery_codes SET used = true WHERE id = $1`, id)
	return err
}

func (r *recoveryCodeRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM recovery_codes WHERE user_id = $1`, userID)
	return err
}

func (r *recoveryCodeRepo) GetByUserAndHash(ctx context.Context, userID uuid.UUID, codeHash string) (*models.RecoveryCode, error) {
	var c models.RecoveryCode
	err := r.pool.QueryRow(ctx, `SELECT id, user_id, code_hash, used FROM recovery_codes WHERE user_id = $1 AND code_hash = $2`,
		userID, codeHash).Scan(&c.ID, &c.UserID, &c.CodeHash, &c.Used)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &c, err
}

// --- WebAuthn Credential Repository ---

type webauthnCredRepo struct{ pool *pgxpool.Pool }

func NewWebAuthnCredentialRepository(pool *pgxpool.Pool) models.WebAuthnCredentialRepository {
	return &webauthnCredRepo{pool: pool}
}

func (r *webauthnCredRepo) Create(ctx context.Context, cred *models.WebAuthnCredential) error {
	cred.ID = uuid.New()
	cred.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO webauthn_credentials (id, user_id, credential_id, public_key, sign_count, aaguid, name, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		cred.ID, cred.UserID, cred.CredentialID, cred.PublicKey, cred.SignCount, cred.AAGUID, cred.Name, cred.CreatedAt)
	return err
}

func (r *webauthnCredRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.WebAuthnCredential, error) {
	var c models.WebAuthnCredential
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, credential_id, public_key, sign_count, aaguid, name, created_at
		FROM webauthn_credentials WHERE id = $1`, id).
		Scan(&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.SignCount, &c.AAGUID, &c.Name, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &c, err
}

func (r *webauthnCredRepo) GetByCredentialID(ctx context.Context, credentialID []byte) (*models.WebAuthnCredential, error) {
	var c models.WebAuthnCredential
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, credential_id, public_key, sign_count, aaguid, name, created_at
		FROM webauthn_credentials WHERE credential_id = $1`, credentialID).
		Scan(&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.SignCount, &c.AAGUID, &c.Name, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &c, err
}

func (r *webauthnCredRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.WebAuthnCredential, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, credential_id, public_key, sign_count, aaguid, name, created_at
		FROM webauthn_credentials WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []models.WebAuthnCredential
	for rows.Next() {
		var c models.WebAuthnCredential
		if err := rows.Scan(&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.SignCount, &c.AAGUID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return creds, nil
}

func (r *webauthnCredRepo) Update(ctx context.Context, cred *models.WebAuthnCredential) error {
	_, err := r.pool.Exec(ctx, `UPDATE webauthn_credentials SET sign_count = $1, name = $2 WHERE id = $3`,
		cred.SignCount, cred.Name, cred.ID)
	return err
}

func (r *webauthnCredRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM webauthn_credentials WHERE id = $1`, id)
	return err
}

// --- Session Repository (DB-backed) ---

type sessionRepo struct{ pool *pgxpool.Pool }

func NewSessionRepository(pool *pgxpool.Pool) models.SessionRepository {
	return &sessionRepo{pool: pool}
}

func (r *sessionRepo) Create(ctx context.Context, session *models.Session) error {
	session.ID = uuid.New()
	session.CreatedAt = time.Now().UTC()
	session.LastActiveAt = session.CreatedAt
	if session.DeviceInfo == nil {
		session.DeviceInfo = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, tenant_id, device_info, ip_address, user_agent, created_at, expires_at, last_active_at)
		VALUES ($1, $2, $3, $4, $5::inet, $6, $7, $8, $9)`,
		session.ID, session.UserID, session.TenantID, session.DeviceInfo, nilIfEmpty(session.IP), session.UserAgent,
		session.CreatedAt, session.ExpiresAt, session.LastActiveAt)
	return err
}

func (r *sessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var s models.Session
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, tenant_id, device_info, COALESCE(host(ip_address), ''), COALESCE(user_agent, ''), created_at, expires_at, last_active_at
		FROM sessions WHERE id = $1`, id).
		Scan(&s.ID, &s.UserID, &s.TenantID, &s.DeviceInfo, &s.IP, &s.UserAgent, &s.CreatedAt, &s.ExpiresAt, &s.LastActiveAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &s, err
}

func (r *sessionRepo) Update(ctx context.Context, session *models.Session) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sessions SET last_active_at = $1, expires_at = $2 WHERE id = $3`,
		session.LastActiveAt, session.ExpiresAt, session.ID)
	return err
}

func (r *sessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *sessionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, tenant_id, device_info, COALESCE(host(ip_address), ''), COALESCE(user_agent, ''), created_at, expires_at, last_active_at
		FROM sessions WHERE user_id = $1 ORDER BY last_active_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.TenantID, &s.DeviceInfo, &s.IP, &s.UserAgent, &s.CreatedAt, &s.ExpiresAt, &s.LastActiveAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *sessionRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (r *sessionRepo) DeleteByTenant(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE tenant_id = $1`, tenantID)
	return err
}

// --- Audit Log Repository ---

type auditLogRepo struct{ pool *pgxpool.Pool }

func NewAuditLogRepository(pool *pgxpool.Pool) models.AuditLogRepository {
	return &auditLogRepo{pool: pool}
}

func (r *auditLogRepo) Create(ctx context.Context, log *models.AuditLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now().UTC()
	if log.Metadata == nil {
		log.Metadata = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (id, tenant_id, actor_id, action, target_type, target_id, metadata, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::inet, $9)`,
		log.ID, log.TenantID, log.ActorID, log.Action, nilIfEmpty(log.TargetType), nilIfEmpty(log.TargetID), log.Metadata, nilIfEmpty(log.IP), log.CreatedAt)
	return err
}

func (r *auditLogRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams, action string) (*models.PaginatedResult[models.AuditLog], error) {
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

	if action != "" {
		err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1 AND action = $2`, tenantID, action).Scan(&total)
		if err != nil {
			return nil, err
		}
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, actor_id, action, COALESCE(target_type, ''), COALESCE(target_id::text, ''), metadata, COALESCE(host(ip_address), ''), created_at
			FROM audit_logs WHERE tenant_id = $1 AND action = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
			tenantID, action, params.PerPage, offset)
	} else {
		err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`, tenantID).Scan(&total)
		if err != nil {
			return nil, err
		}
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, actor_id, action, COALESCE(target_type, ''), COALESCE(target_id::text, ''), metadata, COALESCE(host(ip_address), ''), created_at
			FROM audit_logs WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			tenantID, params.PerPage, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ActorID, &l.Action, &l.TargetType, &l.TargetID, &l.Metadata, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []models.AuditLog{}
	}
	return &models.PaginatedResult[models.AuditLog]{
		Data: logs, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

// --- Webhook Repository ---

type webhookRepo struct{ pool *pgxpool.Pool }

func NewWebhookRepository(pool *pgxpool.Pool) models.WebhookRepository {
	return &webhookRepo{pool: pool}
}

func (r *webhookRepo) Create(ctx context.Context, webhook *models.Webhook) error {
	webhook.ID = uuid.New()
	now := time.Now().UTC()
	webhook.CreatedAt = now
	webhook.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO webhooks (id, tenant_id, url, events, secret, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		webhook.ID, webhook.TenantID, webhook.URL, webhook.Events, webhook.Secret, webhook.Active, webhook.CreatedAt, webhook.UpdatedAt)
	return err
}

func (r *webhookRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Webhook, error) {
	var w models.Webhook
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, url, events, secret, active, created_at, updated_at
		FROM webhooks WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&w.ID, &w.TenantID, &w.URL, &w.Events, &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &w, err
}

func (r *webhookRepo) Update(ctx context.Context, webhook *models.Webhook) error {
	webhook.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE webhooks SET url = $1, events = $2, active = $3, updated_at = $4 WHERE id = $5 AND tenant_id = $6`,
		webhook.URL, webhook.Events, webhook.Active, webhook.UpdatedAt, webhook.ID, webhook.TenantID)
	return err
}

func (r *webhookRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM webhooks WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *webhookRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Webhook], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM webhooks WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, url, events, secret, active, created_at, updated_at
		FROM webhooks WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		if err := rows.Scan(&w.ID, &w.TenantID, &w.URL, &w.Events, &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	if webhooks == nil {
		webhooks = []models.Webhook{}
	}
	return &models.PaginatedResult[models.Webhook]{
		Data: webhooks, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

func (r *webhookRepo) ListByEvent(ctx context.Context, tenantID uuid.UUID, event string) ([]models.Webhook, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, url, events, secret, active, created_at, updated_at
		FROM webhooks WHERE tenant_id = $1 AND active = true AND $2 = ANY(events)`,
		tenantID, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		if err := rows.Scan(&w.ID, &w.TenantID, &w.URL, &w.Events, &w.Secret, &w.Active, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

// --- Action Repository ---

type actionRepo struct{ pool *pgxpool.Pool }

func NewActionRepository(pool *pgxpool.Pool) models.ActionRepository {
	return &actionRepo{pool: pool}
}

func (r *actionRepo) Create(ctx context.Context, action *models.Action) error {
	action.ID = uuid.New()
	now := time.Now().UTC()
	action.CreatedAt = now
	action.UpdatedAt = now
	if action.TimeoutMs == 0 {
		action.TimeoutMs = 5000
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO actions (id, tenant_id, trigger, name, code, enabled, execution_order, timeout_ms, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		action.ID, action.TenantID, action.Trigger, action.Name, action.Code, action.Enabled, action.Order, action.TimeoutMs, action.CreatedAt, action.UpdatedAt)
	return err
}

func (r *actionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Action, error) {
	var a models.Action
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, trigger, name, code, enabled, execution_order, timeout_ms, created_at, updated_at
		FROM actions WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&a.ID, &a.TenantID, &a.Trigger, &a.Name, &a.Code, &a.Enabled, &a.Order, &a.TimeoutMs, &a.CreatedAt, &a.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &a, err
}

func (r *actionRepo) Update(ctx context.Context, action *models.Action) error {
	action.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE actions SET trigger = $1, name = $2, code = $3, enabled = $4, execution_order = $5, timeout_ms = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9`,
		action.Trigger, action.Name, action.Code, action.Enabled, action.Order, action.TimeoutMs, action.UpdatedAt, action.ID, action.TenantID)
	return err
}

func (r *actionRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM actions WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *actionRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Action], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM actions WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, trigger, name, code, enabled, execution_order, timeout_ms, created_at, updated_at
		FROM actions WHERE tenant_id = $1 ORDER BY execution_order ASC, created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var actions []models.Action
	for rows.Next() {
		var a models.Action
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Trigger, &a.Name, &a.Code, &a.Enabled, &a.Order, &a.TimeoutMs, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	if actions == nil {
		actions = []models.Action{}
	}
	return &models.PaginatedResult[models.Action]{
		Data: actions, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

func (r *actionRepo) ListByTrigger(ctx context.Context, tenantID uuid.UUID, trigger string) ([]models.Action, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, trigger, name, code, enabled, execution_order, timeout_ms, created_at, updated_at
		FROM actions WHERE tenant_id = $1 AND trigger = $2 AND enabled = true ORDER BY execution_order ASC`,
		tenantID, trigger)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var actions []models.Action
	for rows.Next() {
		var a models.Action
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Trigger, &a.Name, &a.Code, &a.Enabled, &a.Order, &a.TimeoutMs, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, nil
}

// --- Email Template Repository ---

type emailTemplateRepo struct{ pool *pgxpool.Pool }

func NewEmailTemplateRepository(pool *pgxpool.Pool) models.EmailTemplateRepository {
	return &emailTemplateRepo{pool: pool}
}

func (r *emailTemplateRepo) Create(ctx context.Context, tmpl *models.EmailTemplate) error {
	tmpl.ID = uuid.New()
	now := time.Now().UTC()
	tmpl.CreatedAt = now
	tmpl.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO email_templates (id, tenant_id, type, locale, subject, body_mjml, body_html, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		tmpl.ID, tmpl.TenantID, tmpl.Type, tmpl.Locale, tmpl.Subject, tmpl.BodyMJML, tmpl.BodyHTML, tmpl.CreatedAt, tmpl.UpdatedAt)
	return err
}

func (r *emailTemplateRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.EmailTemplate, error) {
	var t models.EmailTemplate
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, type, locale, subject, COALESCE(body_mjml, ''), body_html, created_at, updated_at
		FROM email_templates WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&t.ID, &t.TenantID, &t.Type, &t.Locale, &t.Subject, &t.BodyMJML, &t.BodyHTML, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &t, err
}

func (r *emailTemplateRepo) GetByTypeAndLocale(ctx context.Context, tenantID uuid.UUID, typ, locale string) (*models.EmailTemplate, error) {
	var t models.EmailTemplate
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, type, locale, subject, COALESCE(body_mjml, ''), body_html, created_at, updated_at
		FROM email_templates WHERE tenant_id = $1 AND type = $2 AND locale = $3`, tenantID, typ, locale).
		Scan(&t.ID, &t.TenantID, &t.Type, &t.Locale, &t.Subject, &t.BodyMJML, &t.BodyHTML, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &t, err
}

func (r *emailTemplateRepo) Update(ctx context.Context, tmpl *models.EmailTemplate) error {
	tmpl.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE email_templates SET type = $1, locale = $2, subject = $3, body_mjml = $4, body_html = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8`,
		tmpl.Type, tmpl.Locale, tmpl.Subject, tmpl.BodyMJML, tmpl.BodyHTML, tmpl.UpdatedAt, tmpl.ID, tmpl.TenantID)
	return err
}

func (r *emailTemplateRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM email_templates WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *emailTemplateRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.EmailTemplate], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM email_templates WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, type, locale, subject, COALESCE(body_mjml, ''), body_html, created_at, updated_at
		FROM email_templates WHERE tenant_id = $1 ORDER BY type, locale LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var templates []models.EmailTemplate
	for rows.Next() {
		var t models.EmailTemplate
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Type, &t.Locale, &t.Subject, &t.BodyMJML, &t.BodyHTML, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	if templates == nil {
		templates = []models.EmailTemplate{}
	}
	return &models.PaginatedResult[models.EmailTemplate]{
		Data: templates, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

// --- API Key Repository ---

type apiKeyRepo struct{ pool *pgxpool.Pool }

func NewAPIKeyRepository(pool *pgxpool.Pool) models.APIKeyRepository {
	return &apiKeyRepo{pool: pool}
}

func (r *apiKeyRepo) Create(ctx context.Context, key *models.APIKey) error {
	key.ID = uuid.New()
	key.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO api_keys (id, tenant_id, name, key_prefix, key_hash, scopes, rate_limit, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		key.ID, key.TenantID, key.Name, key.KeyPrefix, key.KeyHash, key.Scopes, key.RateLimit, key.ExpiresAt, key.CreatedAt)
	return err
}

func (r *apiKeyRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.APIKey, error) {
	var k models.APIKey
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, key_prefix, key_hash, scopes, rate_limit, expires_at, created_at
		FROM api_keys WHERE id = $1 AND tenant_id = $2`, id, tenantID).
		Scan(&k.ID, &k.TenantID, &k.Name, &k.KeyPrefix, &k.KeyHash, &k.Scopes, &k.RateLimit, &k.ExpiresAt, &k.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &k, err
}

func (r *apiKeyRepo) GetByKeyHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	var k models.APIKey
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, key_prefix, key_hash, scopes, rate_limit, expires_at, created_at
		FROM api_keys WHERE key_hash = $1`, keyHash).
		Scan(&k.ID, &k.TenantID, &k.Name, &k.KeyPrefix, &k.KeyHash, &k.Scopes, &k.RateLimit, &k.ExpiresAt, &k.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &k, err
}

func (r *apiKeyRepo) Update(ctx context.Context, key *models.APIKey) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE api_keys SET name = $1, scopes = $2, rate_limit = $3, expires_at = $4 WHERE id = $5 AND tenant_id = $6`,
		key.Name, key.Scopes, key.RateLimit, key.ExpiresAt, key.ID, key.TenantID)
	return err
}

func (r *apiKeyRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM api_keys WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *apiKeyRepo) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.APIKey], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM api_keys WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, key_prefix, key_hash, scopes, rate_limit, expires_at, created_at
		FROM api_keys WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, params.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.TenantID, &k.Name, &k.KeyPrefix, &k.KeyHash, &k.Scopes, &k.RateLimit, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []models.APIKey{}
	}
	return &models.PaginatedResult[models.APIKey]{
		Data: keys, Total: total, Page: params.Page, PerPage: params.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PerPage))),
	}, nil
}

// --- FGA Tuple Repository ---

type fgaTupleRepo struct{ pool *pgxpool.Pool }

func NewFGATupleRepository(pool *pgxpool.Pool) models.FGATupleRepository {
	return &fgaTupleRepo{pool: pool}
}

func (r *fgaTupleRepo) Create(ctx context.Context, tuple *models.FGATuple) error {
	tuple.ID = uuid.New()
	tuple.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO fga_tuples (id, tenant_id, user_type, user_id, relation, object_type, object_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tuple.ID, tuple.TenantID, tuple.UserType, tuple.UserID, tuple.Relation, tuple.ObjectType, tuple.ObjectID, tuple.CreatedAt)
	return err
}

func (r *fgaTupleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM fga_tuples WHERE id = $1`, id)
	return err
}

func (r *fgaTupleRepo) Check(ctx context.Context, tenantID uuid.UUID, userType, userID, relation, objectType, objectID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM fga_tuples
			WHERE tenant_id = $1 AND user_type = $2 AND user_id = $3 AND relation = $4 AND object_type = $5 AND object_id = $6
		)`, tenantID, userType, userID, relation, objectType, objectID).Scan(&exists)
	return exists, err
}

func (r *fgaTupleRepo) ListByObject(ctx context.Context, tenantID uuid.UUID, objectType, objectID string) ([]models.FGATuple, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, user_type, user_id, relation, object_type, object_id, created_at
		FROM fga_tuples WHERE tenant_id = $1 AND object_type = $2 AND object_id = $3`,
		tenantID, objectType, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tuples []models.FGATuple
	for rows.Next() {
		var t models.FGATuple
		if err := rows.Scan(&t.ID, &t.TenantID, &t.UserType, &t.UserID, &t.Relation, &t.ObjectType, &t.ObjectID, &t.CreatedAt); err != nil {
			return nil, err
		}
		tuples = append(tuples, t)
	}
	return tuples, nil
}

func (r *fgaTupleRepo) ListByUser(ctx context.Context, tenantID uuid.UUID, userType, userID string) ([]models.FGATuple, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, user_type, user_id, relation, object_type, object_id, created_at
		FROM fga_tuples WHERE tenant_id = $1 AND user_type = $2 AND user_id = $3`,
		tenantID, userType, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tuples []models.FGATuple
	for rows.Next() {
		var t models.FGATuple
		if err := rows.Scan(&t.ID, &t.TenantID, &t.UserType, &t.UserID, &t.Relation, &t.ObjectType, &t.ObjectID, &t.CreatedAt); err != nil {
			return nil, err
		}
		tuples = append(tuples, t)
	}
	return tuples, nil
}
