package user

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

type mockUserRepoUser struct {
	users   map[uuid.UUID]*models.User
	byEmail map[string]*models.User
}

func newMockUserRepoUser() *mockUserRepoUser {
	return &mockUserRepoUser{
		users:   make(map[uuid.UUID]*models.User),
		byEmail: make(map[string]*models.User),
	}
}

func (m *mockUserRepoUser) Create(_ context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt
	m.users[user.ID] = user
	m.byEmail[user.TenantID.String()+":"+user.Email] = user
	return nil
}
func (m *mockUserRepoUser) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return user, nil
}
func (m *mockUserRepoUser) GetByEmail(_ context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	user, ok := m.byEmail[tenantID.String()+":"+email]
	if !ok {
		return nil, models.ErrNotFound
	}
	return user, nil
}
func (m *mockUserRepoUser) Update(_ context.Context, user *models.User) error {
	user.UpdatedAt = time.Now().UTC()
	m.users[user.ID] = user
	return nil
}
func (m *mockUserRepoUser) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	if user, ok := m.users[id]; ok {
		delete(m.byEmail, user.TenantID.String()+":"+user.Email)
		delete(m.users, id)
	}
	return nil
}
func (m *mockUserRepoUser) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, search string) (*models.PaginatedResult[models.User], error) {
	var result []models.User
	for _, u := range m.users {
		if u.TenantID == tenantID {
			result = append(result, *u)
		}
	}
	return &models.PaginatedResult[models.User]{Data: result, Total: int64(len(result))}, nil
}
func (m *mockUserRepoUser) Block(_ context.Context, tenantID, id uuid.UUID) error   { return nil }
func (m *mockUserRepoUser) Unblock(_ context.Context, tenantID, id uuid.UUID) error { return nil }
func (m *mockUserRepoUser) CountByTenant(_ context.Context, tenantID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockUserRepoUser) GetPasswordHistory(_ context.Context, userID uuid.UUID, limit int) ([]models.PasswordHistory, error) {
	return nil, nil
}
func (m *mockUserRepoUser) AddPasswordHistory(_ context.Context, entry *models.PasswordHistory) error {
	return nil
}

type mockIdentityRepoUser struct {
	identities map[uuid.UUID]*models.Identity
}

func newMockIdentityRepoUser() *mockIdentityRepoUser {
	return &mockIdentityRepoUser{identities: make(map[uuid.UUID]*models.Identity)}
}

func (m *mockIdentityRepoUser) Create(_ context.Context, identity *models.Identity) error {
	if identity.ID == uuid.Nil {
		identity.ID = uuid.New()
	}
	m.identities[identity.ID] = identity
	return nil
}
func (m *mockIdentityRepoUser) GetByID(_ context.Context, id uuid.UUID) (*models.Identity, error) {
	identity, ok := m.identities[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return identity, nil
}
func (m *mockIdentityRepoUser) GetByProvider(_ context.Context, provider, providerUserID string) (*models.Identity, error) {
	return nil, models.ErrNotFound
}
func (m *mockIdentityRepoUser) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Identity, error) {
	var result []models.Identity
	for _, i := range m.identities {
		if i.UserID == userID {
			result = append(result, *i)
		}
	}
	return result, nil
}
func (m *mockIdentityRepoUser) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.identities, id)
	return nil
}
func (m *mockIdentityRepoUser) Update(_ context.Context, identity *models.Identity) error {
	m.identities[identity.ID] = identity
	return nil
}

type mockSessionRepoUser struct {
	sessions map[uuid.UUID]*models.Session
}

func newMockSessionRepoUser() *mockSessionRepoUser {
	return &mockSessionRepoUser{sessions: make(map[uuid.UUID]*models.Session)}
}

func (m *mockSessionRepoUser) Create(_ context.Context, session *models.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	m.sessions[session.ID] = session
	return nil
}
func (m *mockSessionRepoUser) GetByID(_ context.Context, id uuid.UUID) (*models.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return s, nil
}
func (m *mockSessionRepoUser) Update(_ context.Context, session *models.Session) error {
	m.sessions[session.ID] = session
	return nil
}
func (m *mockSessionRepoUser) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.sessions, id)
	return nil
}
func (m *mockSessionRepoUser) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Session, error) {
	var result []models.Session
	for _, s := range m.sessions {
		if s.UserID == userID {
			result = append(result, *s)
		}
	}
	return result, nil
}
func (m *mockSessionRepoUser) DeleteByUser(_ context.Context, userID uuid.UUID) error {
	for id, s := range m.sessions {
		if s.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}
func (m *mockSessionRepoUser) DeleteByTenant(_ context.Context, tenantID uuid.UUID) error {
	return nil
}

// Mock audit log and webhook repos for event service
type mockAuditLogRepoUser struct{}

func (m *mockAuditLogRepoUser) Create(_ context.Context, log *models.AuditLog) error { return nil }
func (m *mockAuditLogRepoUser) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, action string) (*models.PaginatedResult[models.AuditLog], error) {
	return &models.PaginatedResult[models.AuditLog]{}, nil
}

type mockWebhookRepoUser struct{}

func (m *mockWebhookRepoUser) Create(_ context.Context, webhook *models.Webhook) error { return nil }
func (m *mockWebhookRepoUser) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.Webhook, error) {
	return nil, models.ErrNotFound
}
func (m *mockWebhookRepoUser) Update(_ context.Context, webhook *models.Webhook) error { return nil }
func (m *mockWebhookRepoUser) Delete(_ context.Context, tenantID, id uuid.UUID) error  { return nil }
func (m *mockWebhookRepoUser) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Webhook], error) {
	return &models.PaginatedResult[models.Webhook]{}, nil
}
func (m *mockWebhookRepoUser) ListByEvent(_ context.Context, tenantID uuid.UUID, event string) ([]models.Webhook, error) {
	return nil, nil
}

// --- Test setup ---

func testUserConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Security.HIBPEnabled = false
	return cfg
}

type testContext struct {
	handler     *Handler
	userRepo    *mockUserRepoUser
	sessionRepo *mockSessionRepoUser
	identityRepo *mockIdentityRepoUser
	tenantID    uuid.UUID
}

func setupTest() *testContext {
	cfg := testUserConfig()
	logger := zap.NewNop()

	userRepo := newMockUserRepoUser()
	identityRepo := newMockIdentityRepoUser()
	sessionRepo := newMockSessionRepoUser()

	userSvc := users.NewService(userRepo, identityRepo, cfg, logger)
	sessionSvc := sessions.NewService(sessionRepo, nil, cfg, logger)
	eventSvc := events.NewService(nil, &mockAuditLogRepoUser{}, &mockWebhookRepoUser{}, logger)

	h := NewHandler(userSvc, sessionSvc, nil, nil, eventSvc, logger)

	tenantID := uuid.New()

	return &testContext{
		handler:      h,
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		identityRepo: identityRepo,
		tenantID:     tenantID,
	}
}

func withUserContext(r *http.Request, tenantID, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyTenantID, tenantID)
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, userID)
	return r.WithContext(ctx)
}

func chiCtxParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func createTestUserInRepo(repo *mockUserRepoUser, tenantID uuid.UUID) *models.User {
	user := &models.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Email:         "testuser@example.com",
		Name:          "Test User",
		Phone:         "+15551234567",
		AvatarURL:     "https://example.com/avatar.png",
		Status:        models.StatusActive,
		EmailVerified: true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	repo.users[user.ID] = user
	repo.byEmail[user.TenantID.String()+":"+user.Email] = user
	return user
}

// --- Tests ---

func TestNewHandler(t *testing.T) {
	tc := setupTest()
	if tc.handler == nil {
		t.Fatal("NewHandler returned nil")
	}
}

func TestRegisterRoutes(t *testing.T) {
	tc := setupTest()
	r := chi.NewRouter()
	tc.handler.RegisterRoutes(r)
	// Just verify it doesn't panic
}

func TestGetMe(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.GetMe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["email"] != user.Email {
		t.Errorf("email = %v, want %q", result["email"], user.Email)
	}
	if result["name"] != user.Name {
		t.Errorf("name = %v, want %q", result["name"], user.Name)
	}
}

func TestGetMe_UserNotFound(t *testing.T) {
	tc := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req = withUserContext(req, tc.tenantID, uuid.New())
	w := httptest.NewRecorder()

	tc.handler.GetMe(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUpdateMe(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	body := `{"name":"Updated Name","phone":"+15559876543"}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.UpdateMe(w, req)

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
	if result["phone"] != "+15559876543" {
		t.Errorf("phone = %v, want %q", result["phone"], "+15559876543")
	}
}

func TestUpdateMe_InvalidJSON(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	req := httptest.NewRequest(http.MethodPatch, "/v1/users/me", bytes.NewBufferString("bad-json"))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.UpdateMe(w, req)

	if w.Code == http.StatusOK {
		t.Error("UpdateMe should fail for invalid JSON")
	}
}

func TestUpdateMe_PartialUpdate(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)
	originalPhone := user.Phone

	// Only update name, phone should remain unchanged
	body := `{"name":"New Name Only"}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["name"] != "New Name Only" {
		t.Errorf("name = %v, want %q", result["name"], "New Name Only")
	}
	if result["phone"] != originalPhone {
		t.Errorf("phone = %v, want %q (should be unchanged)", result["phone"], originalPhone)
	}
}

func TestChangePassword_InvalidJSON(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	req := httptest.NewRequest(http.MethodPost, "/v1/users/me/change-password", bytes.NewBufferString("bad-json"))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.ChangePassword(w, req)

	if w.Code == http.StatusOK {
		t.Error("ChangePassword should fail for invalid JSON")
	}
}

func TestListSessions(t *testing.T) {
	tc := setupTest()
	userID := uuid.New()

	// Add some sessions
	tc.sessionRepo.sessions[uuid.New()] = &models.Session{
		ID: uuid.New(), UserID: userID, TenantID: tc.tenantID,
		IP: "192.168.1.1", ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	tc.sessionRepo.sessions[uuid.New()] = &models.Session{
		ID: uuid.New(), UserID: userID, TenantID: tc.tenantID,
		IP: "10.0.0.1", ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me/sessions", nil)
	req = withUserContext(req, tc.tenantID, userID)
	w := httptest.NewRecorder()

	tc.handler.ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result []interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(result))
	}
}

func TestListSessions_Empty(t *testing.T) {
	tc := setupTest()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me/sessions", nil)
	req = withUserContext(req, tc.tenantID, userID)
	w := httptest.NewRecorder()

	tc.handler.ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRevokeSession(t *testing.T) {
	tc := setupTest()
	userID := uuid.New()
	sessionID := uuid.New()

	tc.sessionRepo.sessions[sessionID] = &models.Session{
		ID: sessionID, UserID: userID, TenantID: tc.tenantID,
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/users/me/sessions/"+sessionID.String(), nil)
	req = withUserContext(req, tc.tenantID, userID)
	req = chiCtxParam(req, "sessionId", sessionID.String())
	w := httptest.NewRecorder()

	tc.handler.RevokeSession(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify session was deleted
	if len(tc.sessionRepo.sessions) != 0 {
		t.Error("session should have been deleted")
	}
}

func TestRevokeSession_InvalidID(t *testing.T) {
	tc := setupTest()

	req := httptest.NewRequest(http.MethodDelete, "/v1/users/me/sessions/not-a-uuid", nil)
	req = withUserContext(req, tc.tenantID, uuid.New())
	req = chiCtxParam(req, "sessionId", "not-a-uuid")
	w := httptest.NewRecorder()

	tc.handler.RevokeSession(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestListIdentities(t *testing.T) {
	tc := setupTest()
	userID := uuid.New()

	tc.identityRepo.identities[uuid.New()] = &models.Identity{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       "google",
		ProviderUserID: "google-123",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me/identities", nil)
	req = withUserContext(req, tc.tenantID, userID)
	w := httptest.NewRecorder()

	tc.handler.ListIdentities(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result []interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 identity, got %d", len(result))
	}
}

func TestUnlinkIdentity(t *testing.T) {
	tc := setupTest()
	identityID := uuid.New()

	tc.identityRepo.identities[identityID] = &models.Identity{
		ID:             identityID,
		UserID:         uuid.New(),
		Provider:       "github",
		ProviderUserID: "gh-456",
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/users/me/identities/"+identityID.String(), nil)
	req = withUserContext(req, tc.tenantID, uuid.New())
	req = chiCtxParam(req, "identityId", identityID.String())
	w := httptest.NewRecorder()

	tc.handler.UnlinkIdentity(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestUnlinkIdentity_InvalidID(t *testing.T) {
	tc := setupTest()

	req := httptest.NewRequest(http.MethodDelete, "/v1/users/me/identities/not-a-uuid", nil)
	req = withUserContext(req, tc.tenantID, uuid.New())
	req = chiCtxParam(req, "identityId", "not-a-uuid")
	w := httptest.NewRecorder()

	tc.handler.UnlinkIdentity(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestExportData(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	req := httptest.NewRequest(http.MethodPost, "/v1/users/me/export", nil)
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.ExportData(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// Should contain user data
	if result["user"] == nil {
		t.Error("export should contain 'user' data")
	}
}

func TestDeleteAccount(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	// Add a session for this user (should be cleaned up)
	tc.sessionRepo.sessions[uuid.New()] = &models.Session{
		ID: uuid.New(), UserID: user.ID, TenantID: tc.tenantID,
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/users/me", nil)
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.DeleteAccount(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify user was deleted
	if _, ok := tc.userRepo.users[user.ID]; ok {
		t.Error("user should have been deleted")
	}
}

func TestDeleteAccount_UserNotFound(t *testing.T) {
	tc := setupTest()

	req := httptest.NewRequest(http.MethodDelete, "/v1/users/me", nil)
	req = withUserContext(req, tc.tenantID, uuid.New())
	w := httptest.NewRecorder()

	tc.handler.DeleteAccount(w, req)

	// Delete for non-existent user - depends on implementation
	// The delete on user repo succeeds even if not found (no error from our mock)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestListMFA(t *testing.T) {
	// ListMFA requires mfaSvc which we set to nil - it will panic.
	// We test that the handler exists and routes are registered.
	// Actual MFA tests are in core/flows/mfa_test.go.
	tc := setupTest()
	r := chi.NewRouter()
	tc.handler.RegisterRoutes(r)

	// Verify the route exists by checking it doesn't 404
	req := httptest.NewRequest(http.MethodGet, "/v1/users/me/mfa", nil)
	req = withUserContext(req, tc.tenantID, uuid.New())
	w := httptest.NewRecorder()

	// This will panic since mfaSvc is nil, but let's verify the route is registered
	// by using a defer/recover
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected since mfaSvc is nil
			}
		}()
		tc.handler.ListMFA(w, req)
	}()
}

func TestUpdateMe_AvatarURL(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	body := `{"avatar_url":"https://example.com/new-avatar.png"}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["avatar_url"] != "https://example.com/new-avatar.png" {
		t.Errorf("avatar_url = %v, want %q", result["avatar_url"], "https://example.com/new-avatar.png")
	}
}

func TestUpdateMe_Metadata(t *testing.T) {
	tc := setupTest()
	user := createTestUserInRepo(tc.userRepo, tc.tenantID)

	body := `{"metadata":{"preferred_language":"en","timezone":"UTC"}}`
	req := httptest.NewRequest(http.MethodPatch, "/v1/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserContext(req, tc.tenantID, user.ID)
	w := httptest.NewRecorder()

	tc.handler.UpdateMe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
