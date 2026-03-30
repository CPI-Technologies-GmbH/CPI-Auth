package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/events"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/sessions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/users"
)

// --- Mock Repositories ---

type mockUserRepoAdmin struct {
	users   map[uuid.UUID]*models.User
	byEmail map[string]*models.User
}

func newMockUserRepoAdmin() *mockUserRepoAdmin {
	return &mockUserRepoAdmin{
		users:   make(map[uuid.UUID]*models.User),
		byEmail: make(map[string]*models.User),
	}
}

func (m *mockUserRepoAdmin) Create(_ context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt
	m.users[user.ID] = user
	m.byEmail[user.TenantID.String()+":"+user.Email] = user
	return nil
}
func (m *mockUserRepoAdmin) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return user, nil
}
func (m *mockUserRepoAdmin) GetByEmail(_ context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	user, ok := m.byEmail[tenantID.String()+":"+email]
	if !ok {
		return nil, models.ErrNotFound
	}
	return user, nil
}
func (m *mockUserRepoAdmin) Update(_ context.Context, user *models.User) error {
	user.UpdatedAt = time.Now().UTC()
	m.users[user.ID] = user
	return nil
}
func (m *mockUserRepoAdmin) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		delete(m.byEmail, user.TenantID.String()+":"+user.Email)
		delete(m.users, id)
	}
	return nil
}
func (m *mockUserRepoAdmin) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, search string) (*models.PaginatedResult[models.User], error) {
	var result []models.User
	for _, u := range m.users {
		if u.TenantID == tenantID {
			result = append(result, *u)
		}
	}
	return &models.PaginatedResult[models.User]{
		Data:    result,
		Total:   int64(len(result)),
		Page:    params.Page,
		PerPage: params.PerPage,
	}, nil
}
func (m *mockUserRepoAdmin) Block(_ context.Context, tenantID, id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		user.Status = models.StatusBlocked
	}
	return nil
}
func (m *mockUserRepoAdmin) Unblock(_ context.Context, tenantID, id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		user.Status = models.StatusActive
	}
	return nil
}
func (m *mockUserRepoAdmin) CountByTenant(_ context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	for _, u := range m.users {
		if u.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}
func (m *mockUserRepoAdmin) GetPasswordHistory(_ context.Context, userID uuid.UUID, limit int) ([]models.PasswordHistory, error) {
	return nil, nil
}
func (m *mockUserRepoAdmin) AddPasswordHistory(_ context.Context, entry *models.PasswordHistory) error {
	return nil
}

type mockIdentityRepoAdmin struct{}

func (m *mockIdentityRepoAdmin) Create(_ context.Context, identity *models.Identity) error { return nil }
func (m *mockIdentityRepoAdmin) GetByID(_ context.Context, id uuid.UUID) (*models.Identity, error) {
	return nil, models.ErrNotFound
}
func (m *mockIdentityRepoAdmin) GetByProvider(_ context.Context, provider, providerUserID string) (*models.Identity, error) {
	return nil, models.ErrNotFound
}
func (m *mockIdentityRepoAdmin) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Identity, error) {
	return nil, nil
}
func (m *mockIdentityRepoAdmin) Delete(_ context.Context, id uuid.UUID) error { return nil }
func (m *mockIdentityRepoAdmin) Update(_ context.Context, identity *models.Identity) error {
	return nil
}

type mockTenantRepo struct {
	tenants map[uuid.UUID]*models.Tenant
}

func newMockTenantRepo() *mockTenantRepo {
	return &mockTenantRepo{tenants: make(map[uuid.UUID]*models.Tenant)}
}

func (m *mockTenantRepo) Create(_ context.Context, tenant *models.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	tenant.CreatedAt = time.Now().UTC()
	tenant.UpdatedAt = tenant.CreatedAt
	m.tenants[tenant.ID] = tenant
	return nil
}
func (m *mockTenantRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Tenant, error) {
	tenant, ok := m.tenants[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return tenant, nil
}
func (m *mockTenantRepo) GetBySlug(_ context.Context, slug string) (*models.Tenant, error) {
	for _, t := range m.tenants {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, models.ErrNotFound
}
func (m *mockTenantRepo) GetByDomain(_ context.Context, domain string) (*models.Tenant, error) {
	for _, t := range m.tenants {
		if t.Domain == domain {
			return t, nil
		}
	}
	return nil, models.ErrNotFound
}
func (m *mockTenantRepo) Update(_ context.Context, tenant *models.Tenant) error {
	m.tenants[tenant.ID] = tenant
	return nil
}
func (m *mockTenantRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.tenants, id)
	return nil
}
func (m *mockTenantRepo) List(_ context.Context, params models.PaginationParams) (*models.PaginatedResult[models.Tenant], error) {
	var result []models.Tenant
	for _, t := range m.tenants {
		result = append(result, *t)
	}
	return &models.PaginatedResult[models.Tenant]{Data: result, Total: int64(len(result)), Page: params.Page, PerPage: params.PerPage}, nil
}

type mockAppRepoAdmin struct {
	apps map[uuid.UUID]*models.Application
}

func newMockAppRepoAdmin() *mockAppRepoAdmin {
	return &mockAppRepoAdmin{apps: make(map[uuid.UUID]*models.Application)}
}

func (m *mockAppRepoAdmin) Create(_ context.Context, app *models.Application) error {
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	app.CreatedAt = time.Now().UTC()
	app.UpdatedAt = app.CreatedAt
	m.apps[app.ID] = app
	return nil
}
func (m *mockAppRepoAdmin) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.Application, error) {
	app, ok := m.apps[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return app, nil
}
func (m *mockAppRepoAdmin) GetByClientID(_ context.Context, clientID string) (*models.Application, error) {
	for _, app := range m.apps {
		if app.ClientID == clientID {
			return app, nil
		}
	}
	return nil, models.ErrNotFound
}
func (m *mockAppRepoAdmin) Update(_ context.Context, app *models.Application) error {
	m.apps[app.ID] = app
	return nil
}
func (m *mockAppRepoAdmin) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	delete(m.apps, id)
	return nil
}
func (m *mockAppRepoAdmin) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Application], error) {
	var result []models.Application
	for _, app := range m.apps {
		if app.TenantID == tenantID {
			result = append(result, *app)
		}
	}
	return &models.PaginatedResult[models.Application]{Data: result, Total: int64(len(result))}, nil
}

type mockSessionRepoAdmin struct {
	sessionsByUser map[uuid.UUID][]models.Session
}

func newMockSessionRepoAdmin() *mockSessionRepoAdmin {
	return &mockSessionRepoAdmin{
		sessionsByUser: make(map[uuid.UUID][]models.Session),
	}
}

func (m *mockSessionRepoAdmin) Create(_ context.Context, session *models.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	m.sessionsByUser[session.UserID] = append(m.sessionsByUser[session.UserID], *session)
	return nil
}
func (m *mockSessionRepoAdmin) GetByID(_ context.Context, id uuid.UUID) (*models.Session, error) {
	return nil, models.ErrNotFound
}
func (m *mockSessionRepoAdmin) Update(_ context.Context, session *models.Session) error { return nil }
func (m *mockSessionRepoAdmin) Delete(_ context.Context, id uuid.UUID) error            { return nil }
func (m *mockSessionRepoAdmin) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Session, error) {
	return m.sessionsByUser[userID], nil
}
func (m *mockSessionRepoAdmin) DeleteByUser(_ context.Context, userID uuid.UUID) error {
	delete(m.sessionsByUser, userID)
	return nil
}
func (m *mockSessionRepoAdmin) DeleteByTenant(_ context.Context, tenantID uuid.UUID) error {
	return nil
}

// Mock audit log and webhook repos for event service
type mockAuditLogRepoAdmin struct{}

func (m *mockAuditLogRepoAdmin) Create(_ context.Context, log *models.AuditLog) error { return nil }
func (m *mockAuditLogRepoAdmin) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, action string) (*models.PaginatedResult[models.AuditLog], error) {
	return &models.PaginatedResult[models.AuditLog]{Data: []models.AuditLog{}, Total: 0}, nil
}

type mockWebhookRepoAdmin struct{}

func (m *mockWebhookRepoAdmin) Create(_ context.Context, webhook *models.Webhook) error { return nil }
func (m *mockWebhookRepoAdmin) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.Webhook, error) {
	return nil, models.ErrNotFound
}
func (m *mockWebhookRepoAdmin) Update(_ context.Context, webhook *models.Webhook) error { return nil }
func (m *mockWebhookRepoAdmin) Delete(_ context.Context, tenantID, id uuid.UUID) error  { return nil }
func (m *mockWebhookRepoAdmin) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Webhook], error) {
	return &models.PaginatedResult[models.Webhook]{Data: []models.Webhook{}, Total: 0}, nil
}
func (m *mockWebhookRepoAdmin) ListByEvent(_ context.Context, tenantID uuid.UUID, event string) ([]models.Webhook, error) {
	return nil, nil
}

// --- Test setup ---

func testAdminConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Security.HIBPEnabled = false
	return cfg
}

func buildHandler(userRepo *mockUserRepoAdmin, tenantRepo *mockTenantRepo, appRepo *mockAppRepoAdmin) *Handler {
	cfg := testAdminConfig()
	logger := zap.NewNop()

	userSvc := users.NewService(userRepo, &mockIdentityRepoAdmin{}, cfg, logger)
	sessionSvc := sessions.NewService(newMockSessionRepoAdmin(), nil, cfg, logger)
	eventSvc := events.NewService(nil, &mockAuditLogRepoAdmin{}, &mockWebhookRepoAdmin{}, logger)

	return NewHandler(userSvc, sessionSvc, eventSvc, nil, nil, nil, tenantRepo, appRepo, nil, nil, &mockWebhookRepoAdmin{}, nil, nil, newMockPermRepoAdmin(), newMockAppPermRepoAdmin(), nil, nil, nil, nil, nil, logger)
}

func buildHandlerWithPerms(
	userRepo *mockUserRepoAdmin,
	tenantRepo *mockTenantRepo,
	appRepo *mockAppRepoAdmin,
	roleRepo *mockRoleRepoAdmin,
	permRepo *mockPermRepoAdmin,
	appPermRepo *mockAppPermRepoAdmin,
) *Handler {
	cfg := testAdminConfig()
	logger := zap.NewNop()

	userSvc := users.NewService(userRepo, &mockIdentityRepoAdmin{}, cfg, logger)
	sessionSvc := sessions.NewService(newMockSessionRepoAdmin(), nil, cfg, logger)
	eventSvc := events.NewService(nil, &mockAuditLogRepoAdmin{}, &mockWebhookRepoAdmin{}, logger)

	return NewHandler(userSvc, sessionSvc, eventSvc, nil, nil, nil, tenantRepo, appRepo, nil, roleRepo, &mockWebhookRepoAdmin{}, nil, nil, permRepo, appPermRepo, nil, nil, nil, nil, nil, logger)
}

// --- Mock Permission Repositories ---

type mockRoleRepoAdmin struct {
	roles     map[uuid.UUID]*models.Role
	userRoles map[uuid.UUID][]models.Role
}

func newMockRoleRepoAdmin() *mockRoleRepoAdmin {
	return &mockRoleRepoAdmin{
		roles:     make(map[uuid.UUID]*models.Role),
		userRoles: make(map[uuid.UUID][]models.Role),
	}
}

func (m *mockRoleRepoAdmin) Create(_ context.Context, role *models.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	role.CreatedAt = time.Now().UTC()
	role.UpdatedAt = role.CreatedAt
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepoAdmin) GetByID(_ context.Context, _ uuid.UUID, id uuid.UUID) (*models.Role, error) {
	role, ok := m.roles[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return role, nil
}
func (m *mockRoleRepoAdmin) GetByName(_ context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	for _, r := range m.roles {
		if r.TenantID == tenantID && r.Name == name {
			return r, nil
		}
	}
	return nil, models.ErrNotFound
}
func (m *mockRoleRepoAdmin) Update(_ context.Context, role *models.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepoAdmin) Delete(_ context.Context, _ uuid.UUID, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}
func (m *mockRoleRepoAdmin) List(_ context.Context, _ uuid.UUID, _ models.PaginationParams) (*models.PaginatedResult[models.Role], error) {
	var roles []models.Role
	for _, r := range m.roles {
		roles = append(roles, *r)
	}
	return &models.PaginatedResult[models.Role]{Data: roles, Total: int64(len(roles))}, nil
}
func (m *mockRoleRepoAdmin) GetRolesForUser(_ context.Context, userID uuid.UUID) ([]models.Role, error) {
	return m.userRoles[userID], nil
}
func (m *mockRoleRepoAdmin) GetRoleHierarchy(_ context.Context, _ uuid.UUID) ([]models.Role, error) {
	return nil, nil
}
func (m *mockRoleRepoAdmin) AssignRoleToUser(_ context.Context, userID, roleID, _ uuid.UUID) error {
	role, ok := m.roles[roleID]
	if !ok {
		return models.ErrNotFound
	}
	m.userRoles[userID] = append(m.userRoles[userID], *role)
	return nil
}
func (m *mockRoleRepoAdmin) RemoveRoleFromUser(_ context.Context, userID, roleID, _ uuid.UUID) error {
	roles := m.userRoles[userID]
	for i, r := range roles {
		if r.ID == roleID {
			m.userRoles[userID] = append(roles[:i], roles[i+1:]...)
			return nil
		}
	}
	return nil
}

type mockPermRepoAdmin struct {
	perms map[uuid.UUID]*models.Permission
}

func newMockPermRepoAdmin() *mockPermRepoAdmin {
	return &mockPermRepoAdmin{perms: make(map[uuid.UUID]*models.Permission)}
}

func (m *mockPermRepoAdmin) Create(_ context.Context, perm *models.Permission) error {
	if perm.ID == uuid.Nil {
		perm.ID = uuid.New()
	}
	perm.CreatedAt = time.Now().UTC()
	perm.UpdatedAt = perm.CreatedAt
	m.perms[perm.ID] = perm
	return nil
}
func (m *mockPermRepoAdmin) GetByID(_ context.Context, _ uuid.UUID, id uuid.UUID) (*models.Permission, error) {
	p, ok := m.perms[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return p, nil
}
func (m *mockPermRepoAdmin) GetByName(_ context.Context, tenantID uuid.UUID, name string) (*models.Permission, error) {
	for _, p := range m.perms {
		if p.TenantID == tenantID && p.Name == name {
			return p, nil
		}
	}
	return nil, models.ErrNotFound
}
func (m *mockPermRepoAdmin) Update(_ context.Context, perm *models.Permission) error {
	m.perms[perm.ID] = perm
	return nil
}
func (m *mockPermRepoAdmin) Delete(_ context.Context, _ uuid.UUID, id uuid.UUID) error {
	delete(m.perms, id)
	return nil
}
func (m *mockPermRepoAdmin) List(_ context.Context, _ uuid.UUID, _ models.PaginationParams) (*models.PaginatedResult[models.Permission], error) {
	var perms []models.Permission
	for _, p := range m.perms {
		perms = append(perms, *p)
	}
	return &models.PaginatedResult[models.Permission]{Data: perms, Total: int64(len(perms))}, nil
}
func (m *mockPermRepoAdmin) ListAll(_ context.Context, _ uuid.UUID) ([]models.Permission, error) {
	var perms []models.Permission
	for _, p := range m.perms {
		perms = append(perms, *p)
	}
	return perms, nil
}
func (m *mockPermRepoAdmin) EnsureSystemDefaults(_ context.Context, _ uuid.UUID) error {
	return nil
}

type mockAppPermRepoAdmin struct {
	perms map[uuid.UUID][]string
}

func newMockAppPermRepoAdmin() *mockAppPermRepoAdmin {
	return &mockAppPermRepoAdmin{perms: make(map[uuid.UUID][]string)}
}

func (m *mockAppPermRepoAdmin) SetPermissions(_ context.Context, appID, _ uuid.UUID, permissions []string) error {
	m.perms[appID] = permissions
	return nil
}
func (m *mockAppPermRepoAdmin) GetPermissions(_ context.Context, appID uuid.UUID) ([]string, error) {
	return m.perms[appID], nil
}

func withAdminContext(r *http.Request, tenantID, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyTenantID, tenantID)
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, userID)
	return r.WithContext(ctx)
}

func chiCtx(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// --- Tests ---

func TestNewAdminHandler(t *testing.T) {
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
}

func TestRegisterRoutes(t *testing.T) {
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	// Verify it doesn't panic and some routes exist
}

func TestListUsers(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()

	user1 := &models.User{ID: uuid.New(), TenantID: tenantID, Email: "user1@test.com", Status: models.StatusActive}
	user2 := &models.User{ID: uuid.New(), TenantID: tenantID, Email: "user2@test.com", Status: models.StatusActive}
	userRepo.users[user1.ID] = user1
	userRepo.users[user2.ID] = user2

	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users?page=1&per_page=20", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("response should contain 'data' array")
	}
	if len(data) != 2 {
		t.Errorf("expected 2 users, got %d", len(data))
	}
}

func TestCreateUser(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()
	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	body := `{"email":"newuser@test.com","password":"StrongP@ss123","name":"New User"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var user map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if user["email"] != "newuser@test.com" {
		t.Errorf("email = %v, want %q", user["email"], "newuser@test.com")
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodPost, "/admin/v1/users", bytes.NewBufferString("bad json"))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	if w.Code == http.StatusCreated {
		t.Error("CreateUser should fail for invalid JSON")
	}
}

func TestGetUser(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()

	user := &models.User{
		ID: uuid.New(), TenantID: tenantID, Email: "getuser@test.com",
		Name: "Get User", Status: models.StatusActive,
	}
	userRepo.users[user.ID] = user
	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users/"+user.ID.String(), nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.GetUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users/"+uuid.New().String(), nil)
	req = withAdminContext(req, uuid.New(), uuid.New())
	req = chiCtx(req, "id", uuid.New().String())
	w := httptest.NewRecorder()

	h.GetUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetUser_InvalidUUID(t *testing.T) {
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users/not-a-uuid", nil)
	req = withAdminContext(req, uuid.New(), uuid.New())
	req = chiCtx(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateUser(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()

	user := &models.User{
		ID: uuid.New(), TenantID: tenantID, Email: "updateuser@test.com",
		Name: "Original Name", Status: models.StatusActive,
	}
	userRepo.users[user.ID] = user
	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	body := `{"name":"Updated Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/admin/v1/users/"+user.ID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.UpdateUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["name"] != "Updated Name" {
		t.Errorf("name = %v, want %q", result["name"], "Updated Name")
	}
}

func TestDeleteUser(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()

	user := &models.User{
		ID: uuid.New(), TenantID: tenantID, Email: "deleteuser@test.com",
		Status: models.StatusActive,
	}
	userRepo.users[user.ID] = user
	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodDelete, "/admin/v1/users/"+user.ID.String(), nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.DeleteUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestListApplications(t *testing.T) {
	tenantID := uuid.New()
	appRepo := newMockAppRepoAdmin()

	app := &models.Application{
		ID: uuid.New(), TenantID: tenantID, Name: "Test App",
		Type: models.AppTypeSPA, ClientID: "test-client",
	}
	appRepo.apps[app.ID] = app
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), appRepo)

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/applications", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.ListApplications(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCreateApplication(t *testing.T) {
	tenantID := uuid.New()
	appRepo := newMockAppRepoAdmin()
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), appRepo)

	body := `{"name":"New App","type":"spa","client_id":"new-client","redirect_uris":["https://app.example.com/cb"]}`
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/applications", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.CreateApplication(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["name"] != "New App" {
		t.Errorf("name = %v, want %q", result["name"], "New App")
	}
	if result["tenant_id"] != tenantID.String() {
		t.Errorf("tenant_id = %v, want %v", result["tenant_id"], tenantID.String())
	}
}

func TestCreateTenant(t *testing.T) {
	tenantRepo := newMockTenantRepo()
	h := buildHandler(newMockUserRepoAdmin(), tenantRepo, newMockAppRepoAdmin())

	body := `{"name":"New Tenant","slug":"new-tenant","domain":"tenant.example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/tenants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateTenant(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestGetTenant(t *testing.T) {
	tenantRepo := newMockTenantRepo()
	tenant := &models.Tenant{ID: uuid.New(), Name: "Test Tenant", Slug: "test-tenant"}
	tenantRepo.tenants[tenant.ID] = tenant
	h := buildHandler(newMockUserRepoAdmin(), tenantRepo, newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/tenants/"+tenant.ID.String(), nil)
	req = chiCtx(req, "id", tenant.ID.String())
	w := httptest.NewRecorder()

	h.GetTenant(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetTenant_NotFound(t *testing.T) {
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/tenants/"+uuid.New().String(), nil)
	req = chiCtx(req, "id", uuid.New().String())
	w := httptest.NewRecorder()

	h.GetTenant(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDeleteTenant(t *testing.T) {
	tenantRepo := newMockTenantRepo()
	tenant := &models.Tenant{ID: uuid.New(), Name: "Delete Me", Slug: "delete-me"}
	tenantRepo.tenants[tenant.ID] = tenant
	h := buildHandler(newMockUserRepoAdmin(), tenantRepo, newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodDelete, "/admin/v1/tenants/"+tenant.ID.String(), nil)
	req = chiCtx(req, "id", tenant.ID.String())
	w := httptest.NewRecorder()

	h.DeleteTenant(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if len(tenantRepo.tenants) != 0 {
		t.Error("tenant should have been deleted")
	}
}

func TestGetStats(t *testing.T) {
	tenantID := uuid.New()
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/stats", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := stats["total_users"]; !ok {
		t.Error("stats should contain total_users")
	}
}

func TestListAuditLogs(t *testing.T) {
	tenantID := uuid.New()
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/logs?page=1&per_page=20", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.ListAuditLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetBranding(t *testing.T) {
	tenantID := uuid.New()
	tenantRepo := newMockTenantRepo()

	tenant := &models.Tenant{
		ID:       tenantID,
		Name:     "Test Tenant",
		Slug:     "test",
		Branding: json.RawMessage(`{"logo":"https://example.com/logo.png"}`),
	}
	tenantRepo.tenants[tenantID] = tenant
	h := buildHandler(newMockUserRepoAdmin(), tenantRepo, newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/branding", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.GetBranding(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["branding"] == nil {
		t.Error("response should contain branding")
	}
}

func TestUpdateBranding(t *testing.T) {
	tenantID := uuid.New()
	tenantRepo := newMockTenantRepo()

	tenant := &models.Tenant{
		ID:   tenantID,
		Name: "Test Tenant",
		Slug: "test",
	}
	tenantRepo.tenants[tenantID] = tenant
	h := buildHandler(newMockUserRepoAdmin(), tenantRepo, newMockAppRepoAdmin())

	body := `{"branding":{"logo":"https://example.com/new-logo.png","primary_color":"#ff0000"}}`
	req := httptest.NewRequest(http.MethodPatch, "/admin/v1/branding", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.UpdateBranding(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestBlockUser(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()

	user := &models.User{
		ID: uuid.New(), TenantID: tenantID, Email: "block@test.com",
		Status: models.StatusActive,
	}
	userRepo.users[user.ID] = user
	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodPost, "/admin/v1/users/"+user.ID.String()+"/block", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.BlockUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["status"] != "blocked" {
		t.Errorf("status = %v, want %q", result["status"], "blocked")
	}
}

func TestUnblockUser(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()

	user := &models.User{
		ID: uuid.New(), TenantID: tenantID, Email: "unblock@test.com",
		Status: models.StatusBlocked,
	}
	userRepo.users[user.ID] = user
	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodPost, "/admin/v1/users/"+user.ID.String()+"/unblock", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.UnblockUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["status"] != "active" {
		t.Errorf("status = %v, want %q", result["status"], "active")
	}
}

func TestGetUserSessions(t *testing.T) {
	userID := uuid.New()
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users/"+userID.String()+"/sessions", nil)
	req = withAdminContext(req, uuid.New(), uuid.New())
	req = chiCtx(req, "id", userID.String())
	w := httptest.NewRecorder()

	h.GetUserSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestForceLogoutUser(t *testing.T) {
	userID := uuid.New()
	h := buildHandler(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodDelete, "/admin/v1/users/"+userID.String()+"/sessions", nil)
	req = withAdminContext(req, uuid.New(), uuid.New())
	req = chiCtx(req, "id", userID.String())
	w := httptest.NewRecorder()

	h.ForceLogoutUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

// --- Pagination Helper ---

func TestGetPagination(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantPage int
		wantPer  int
	}{
		{"defaults", "", 1, 20},
		{"valid", "?page=2&per_page=50", 2, 50},
		{"zero page", "?page=0&per_page=10", 1, 10},
		{"negative page", "?page=-1&per_page=10", 1, 10},
		{"over max per_page", "?page=1&per_page=200", 1, 20},
		{"invalid values", "?page=abc&per_page=xyz", 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			params := getPagination(req)
			if params.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.wantPage)
			}
			if params.PerPage != tt.wantPer {
				t.Errorf("PerPage = %d, want %d", params.PerPage, tt.wantPer)
			}
		})
	}
}

func TestListTenants(t *testing.T) {
	tenantRepo := newMockTenantRepo()
	tenant := &models.Tenant{ID: uuid.New(), Name: "Tenant 1", Slug: "t1"}
	tenantRepo.tenants[tenant.ID] = tenant
	h := buildHandler(newMockUserRepoAdmin(), tenantRepo, newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/tenants", nil)
	w := httptest.NewRecorder()

	h.ListTenants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestExportUsers(t *testing.T) {
	userRepo := newMockUserRepoAdmin()
	tenantID := uuid.New()

	user := &models.User{ID: uuid.New(), TenantID: tenantID, Email: "export@test.com", Status: models.StatusActive}
	userRepo.users[user.ID] = user
	h := buildHandler(userRepo, newMockTenantRepo(), newMockAppRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users/export", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.ExportUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	disposition := w.Header().Get("Content-Disposition")
	if disposition != "attachment; filename=users.json" {
		t.Errorf("Content-Disposition = %q, want %q", disposition, "attachment; filename=users.json")
	}
}

// --- Permission CRUD Tests ---

func TestListPermissions(t *testing.T) {
	tenantID := uuid.New()
	permRepo := newMockPermRepoAdmin()
	perm := &models.Permission{
		ID: uuid.New(), TenantID: tenantID, Name: "users:read",
		DisplayName: "Read Users", GroupName: "Users", IsSystem: true,
	}
	permRepo.perms[perm.ID] = perm

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), permRepo, newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/permissions", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.ListPermissions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestCreatePermission(t *testing.T) {
	tenantID := uuid.New()
	permRepo := newMockPermRepoAdmin()
	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), permRepo, newMockAppPermRepoAdmin())

	body := `{"name":"billing:manage","display_name":"Manage Billing","group_name":"Billing","description":"Full billing access"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/permissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	w := httptest.NewRecorder()

	h.CreatePermission(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["name"] != "billing:manage" {
		t.Errorf("name = %v, want %q", result["name"], "billing:manage")
	}
}

func TestCreatePermission_InvalidJSON(t *testing.T) {
	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), newMockPermRepoAdmin(), newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodPost, "/admin/v1/permissions", bytes.NewBufferString("bad json"))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, uuid.New(), uuid.New())
	w := httptest.NewRecorder()

	h.CreatePermission(w, req)

	if w.Code == http.StatusCreated {
		t.Error("CreatePermission should fail for invalid JSON")
	}
}

func TestGetPermission(t *testing.T) {
	tenantID := uuid.New()
	permRepo := newMockPermRepoAdmin()
	perm := &models.Permission{
		ID: uuid.New(), TenantID: tenantID, Name: "users:write",
		DisplayName: "Write Users", GroupName: "Users",
	}
	permRepo.perms[perm.ID] = perm

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), permRepo, newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/permissions/"+perm.ID.String(), nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", perm.ID.String())
	w := httptest.NewRecorder()

	h.GetPermission(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetPermission_NotFound(t *testing.T) {
	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), newMockPermRepoAdmin(), newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/permissions/"+uuid.New().String(), nil)
	req = withAdminContext(req, uuid.New(), uuid.New())
	req = chiCtx(req, "id", uuid.New().String())
	w := httptest.NewRecorder()

	h.GetPermission(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUpdatePermission(t *testing.T) {
	tenantID := uuid.New()
	permRepo := newMockPermRepoAdmin()
	perm := &models.Permission{
		ID: uuid.New(), TenantID: tenantID, Name: "billing:read",
		DisplayName: "Read Billing", GroupName: "Billing",
	}
	permRepo.perms[perm.ID] = perm

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), permRepo, newMockAppPermRepoAdmin())

	body := `{"display_name":"View Billing"}`
	req := httptest.NewRequest(http.MethodPatch, "/admin/v1/permissions/"+perm.ID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", perm.ID.String())
	w := httptest.NewRecorder()

	h.UpdatePermission(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestDeletePermission(t *testing.T) {
	tenantID := uuid.New()
	permRepo := newMockPermRepoAdmin()
	perm := &models.Permission{
		ID: uuid.New(), TenantID: tenantID, Name: "custom:perm",
		DisplayName: "Custom", GroupName: "Custom", IsSystem: false,
	}
	permRepo.perms[perm.ID] = perm

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), permRepo, newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodDelete, "/admin/v1/permissions/"+perm.ID.String(), nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", perm.ID.String())
	w := httptest.NewRecorder()

	h.DeletePermission(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusNoContent, w.Body.String())
	}

	if len(permRepo.perms) != 0 {
		t.Error("permission should have been deleted")
	}
}

func TestDeletePermission_System(t *testing.T) {
	tenantID := uuid.New()
	permRepo := newMockPermRepoAdmin()
	perm := &models.Permission{
		ID: uuid.New(), TenantID: tenantID, Name: "users:read",
		DisplayName: "Read Users", GroupName: "Users", IsSystem: true,
	}
	permRepo.perms[perm.ID] = perm

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), newMockRoleRepoAdmin(), permRepo, newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodDelete, "/admin/v1/permissions/"+perm.ID.String(), nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", perm.ID.String())
	w := httptest.NewRecorder()

	h.DeletePermission(w, req)

	if w.Code == http.StatusNoContent {
		t.Error("should not delete system permission")
	}
}

// --- User Role Assignment Tests ---

func TestGetUserRoles(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	roleRepo := newMockRoleRepoAdmin()
	role := &models.Role{
		ID: uuid.New(), TenantID: tenantID, Name: "admin",
		Permissions: []string{"users:read"}, Description: "Admin role",
	}
	roleRepo.roles[role.ID] = role
	roleRepo.userRoles[userID] = []models.Role{*role}

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), roleRepo, newMockPermRepoAdmin(), newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/users/"+userID.String()+"/roles", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", userID.String())
	w := httptest.NewRecorder()

	h.GetUserRoles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var roles []interface{}
	if err := json.NewDecoder(w.Body).Decode(&roles); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(roles))
	}
}

func TestAssignUserRole(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	roleRepo := newMockRoleRepoAdmin()
	role := &models.Role{
		ID: uuid.New(), TenantID: tenantID, Name: "editor",
		Permissions: []string{"content:write"},
	}
	roleRepo.roles[role.ID] = role

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), roleRepo, newMockPermRepoAdmin(), newMockAppPermRepoAdmin())

	body := `{"role_id":"` + role.ID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/v1/users/"+userID.String()+"/roles", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", userID.String())
	w := httptest.NewRecorder()

	h.AssignUserRole(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusNoContent, w.Body.String())
	}

	if len(roleRepo.userRoles[userID]) != 1 {
		t.Errorf("expected 1 assigned role, got %d", len(roleRepo.userRoles[userID]))
	}
}

func TestRemoveUserRole(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	roleRepo := newMockRoleRepoAdmin()
	role := &models.Role{
		ID: uuid.New(), TenantID: tenantID, Name: "editor",
		Permissions: []string{"content:write"},
	}
	roleRepo.roles[role.ID] = role
	roleRepo.userRoles[userID] = []models.Role{*role}

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), newMockAppRepoAdmin(), roleRepo, newMockPermRepoAdmin(), newMockAppPermRepoAdmin())

	req := httptest.NewRequest(http.MethodDelete, "/admin/v1/users/"+userID.String()+"/roles/"+role.ID.String(), nil)
	req = withAdminContext(req, tenantID, uuid.New())
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID.String())
	rctx.URLParams.Add("roleId", role.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.RemoveUserRole(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusNoContent, w.Body.String())
	}

	if len(roleRepo.userRoles[userID]) != 0 {
		t.Errorf("expected 0 roles after removal, got %d", len(roleRepo.userRoles[userID]))
	}
}

// --- Application Permission Tests ---

func TestGetApplicationPermissions(t *testing.T) {
	tenantID := uuid.New()
	appRepo := newMockAppRepoAdmin()
	appPermRepo := newMockAppPermRepoAdmin()

	app := &models.Application{
		ID: uuid.New(), TenantID: tenantID, Name: "Test App",
		Type: models.AppTypeSPA, ClientID: "test-client",
	}
	appRepo.apps[app.ID] = app
	appPermRepo.perms[app.ID] = []string{"users:read", "users:write"}

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), appRepo, newMockRoleRepoAdmin(), newMockPermRepoAdmin(), appPermRepo)

	req := httptest.NewRequest(http.MethodGet, "/admin/v1/applications/"+app.ID.String()+"/permissions", nil)
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", app.ID.String())
	w := httptest.NewRecorder()

	h.GetApplicationPermissions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	perms, ok := result["permissions"].([]interface{})
	if !ok {
		t.Fatal("response should contain 'permissions' array")
	}
	if len(perms) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(perms))
	}
}

func TestSetApplicationPermissions(t *testing.T) {
	tenantID := uuid.New()
	appRepo := newMockAppRepoAdmin()
	appPermRepo := newMockAppPermRepoAdmin()

	app := &models.Application{
		ID: uuid.New(), TenantID: tenantID, Name: "Test App",
		Type: models.AppTypeSPA, ClientID: "test-client",
	}
	appRepo.apps[app.ID] = app

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), appRepo, newMockRoleRepoAdmin(), newMockPermRepoAdmin(), appPermRepo)

	body := `{"permissions":["users:read","billing:manage"]}`
	req := httptest.NewRequest(http.MethodPut, "/admin/v1/applications/"+app.ID.String()+"/permissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", app.ID.String())
	w := httptest.NewRecorder()

	h.SetApplicationPermissions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	stored := appPermRepo.perms[app.ID]
	if len(stored) != 2 {
		t.Errorf("expected 2 stored permissions, got %d", len(stored))
	}
}

func TestSetApplicationPermissions_Empty(t *testing.T) {
	tenantID := uuid.New()
	appRepo := newMockAppRepoAdmin()
	appPermRepo := newMockAppPermRepoAdmin()

	app := &models.Application{
		ID: uuid.New(), TenantID: tenantID, Name: "Test App",
		Type: models.AppTypeSPA, ClientID: "test-client",
	}
	appRepo.apps[app.ID] = app
	appPermRepo.perms[app.ID] = []string{"users:read"}

	h := buildHandlerWithPerms(newMockUserRepoAdmin(), newMockTenantRepo(), appRepo, newMockRoleRepoAdmin(), newMockPermRepoAdmin(), appPermRepo)

	body := `{"permissions":[]}`
	req := httptest.NewRequest(http.MethodPut, "/admin/v1/applications/"+app.ID.String()+"/permissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req, tenantID, uuid.New())
	req = chiCtx(req, "id", app.ID.String())
	w := httptest.NewRecorder()

	h.SetApplicationPermissions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	stored := appPermRepo.perms[app.ID]
	if len(stored) != 0 {
		t.Errorf("expected 0 stored permissions after clear, got %d", len(stored))
	}
}
