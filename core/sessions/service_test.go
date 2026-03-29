package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// --- Mock Session Repository ---

type mockSessionRepo struct {
	sessions map[uuid.UUID]*models.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[uuid.UUID]*models.Session),
	}
}

func (m *mockSessionRepo) Create(_ context.Context, session *models.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	session.CreatedAt = time.Now().UTC()
	session.LastActiveAt = session.CreatedAt
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) Update(_ context.Context, session *models.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.sessions, id)
	return nil
}

func (m *mockSessionRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Session, error) {
	var result []models.Session
	for _, s := range m.sessions {
		if s.UserID == userID {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *mockSessionRepo) DeleteByUser(_ context.Context, userID uuid.UUID) error {
	for id, s := range m.sessions {
		if s.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepo) DeleteByTenant(_ context.Context, tenantID uuid.UUID) error {
	for id, s := range m.sessions {
		if s.TenantID == tenantID {
			delete(m.sessions, id)
		}
	}
	return nil
}

// --- Mock Redis Client ---
// We implement a minimal in-memory store that mimics the Redis operations used by the session service.

type mockRedisClient struct {
	data map[string][]byte
	sets map[string]map[string]bool
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string][]byte),
		sets: make(map[string]map[string]bool),
	}
}

// redisClientAdapter wraps mockRedisClient to satisfy the *redis.Client interface usage patterns.
// Since the session service uses *redis.Client directly, we cannot easily mock it.
// Instead, we test only the repository layer (which we can mock) and accept that
// Redis operations will fail gracefully in tests (the service logs warnings but doesn't fail).

func testSessionService() (*Service, *mockSessionRepo) {
	repo := newMockSessionRepo()
	cfg := config.DefaultConfig()
	cfg.Security.SessionLifetime = 24 * time.Hour
	cfg.Security.InactivityTimeout = 1 * time.Hour
	logger := zap.NewNop()

	// Pass nil for Redis - the service handles this gracefully
	svc := &Service{
		sessions: repo,
		redis:    nil, // Will cause Redis ops to be skipped/error gracefully
		cfg:      cfg,
		logger:   logger,
	}

	return svc, repo
}

// --- Tests ---

func TestNewService(t *testing.T) {
	repo := newMockSessionRepo()
	cfg := config.DefaultConfig()
	logger := zap.NewNop()

	svc := NewService(repo, nil, cfg, logger)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestCreateSession(t *testing.T) {
	svc, repo := testSessionService()

	userID := uuid.New()
	tenantID := uuid.New()

	input := CreateSessionInput{
		UserID:     userID,
		TenantID:   tenantID,
		IP:         "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		DeviceInfo: map[string]interface{}{"os": "linux"},
	}

	session, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if session.ID == uuid.Nil {
		t.Error("session ID should not be nil")
	}
	if session.UserID != userID {
		t.Errorf("UserID = %v, want %v", session.UserID, userID)
	}
	if session.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", session.TenantID, tenantID)
	}
	if session.IP != "192.168.1.1" {
		t.Errorf("IP = %q, want %q", session.IP, "192.168.1.1")
	}
	if session.UserAgent != "Mozilla/5.0" {
		t.Errorf("UserAgent = %q, want %q", session.UserAgent, "Mozilla/5.0")
	}
	if session.ExpiresAt.Before(time.Now()) {
		t.Error("session should not already be expired")
	}

	// Verify device info JSON
	var deviceInfo map[string]interface{}
	if err := json.Unmarshal(session.DeviceInfo, &deviceInfo); err != nil {
		t.Fatalf("failed to unmarshal DeviceInfo: %v", err)
	}
	if deviceInfo["os"] != "linux" {
		t.Errorf("DeviceInfo.os = %q, want %q", deviceInfo["os"], "linux")
	}

	// Verify stored in repo
	if len(repo.sessions) != 1 {
		t.Errorf("expected 1 session in repo, got %d", len(repo.sessions))
	}
}

func TestGetSession(t *testing.T) {
	svc, repo := testSessionService()

	userID := uuid.New()
	sessionID := uuid.New()
	repo.sessions[sessionID] = &models.Session{
		ID:           sessionID,
		UserID:       userID,
		TenantID:     uuid.New(),
		IP:           "10.0.0.1",
		ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		LastActiveAt: time.Now().UTC(),
	}

	session, err := svc.Get(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("ID = %v, want %v", session.ID, sessionID)
	}
	if session.UserID != userID {
		t.Errorf("UserID = %v, want %v", session.UserID, userID)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	svc, _ := testSessionService()

	_, err := svc.Get(context.Background(), uuid.New())
	if err == nil {
		t.Error("Get should return error for non-existent session")
	}
}

func TestGetSession_Expired(t *testing.T) {
	svc, repo := testSessionService()

	sessionID := uuid.New()
	repo.sessions[sessionID] = &models.Session{
		ID:           sessionID,
		UserID:       uuid.New(),
		TenantID:     uuid.New(),
		ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour), // Expired
		LastActiveAt: time.Now().UTC().Add(-2 * time.Hour),
	}

	_, err := svc.Get(context.Background(), sessionID)
	if err == nil {
		t.Error("Get should return error for expired session")
	}
	if !models.IsAppError(err, models.ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestDeleteSession(t *testing.T) {
	svc, repo := testSessionService()

	sessionID := uuid.New()
	repo.sessions[sessionID] = &models.Session{
		ID:       sessionID,
		UserID:   uuid.New(),
		TenantID: uuid.New(),
	}

	err := svc.Revoke(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}

	if len(repo.sessions) != 0 {
		t.Error("session should have been deleted")
	}
}

func TestListUserSessions(t *testing.T) {
	svc, repo := testSessionService()

	userID := uuid.New()
	otherUserID := uuid.New()
	tenantID := uuid.New()

	// Create sessions for our user
	repo.sessions[uuid.New()] = &models.Session{
		ID: uuid.New(), UserID: userID, TenantID: tenantID,
	}
	repo.sessions[uuid.New()] = &models.Session{
		ID: uuid.New(), UserID: userID, TenantID: tenantID,
	}
	// Create session for another user
	repo.sessions[uuid.New()] = &models.Session{
		ID: uuid.New(), UserID: otherUserID, TenantID: tenantID,
	}

	sessions, err := svc.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListByUser returned error: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions for user, got %d", len(sessions))
	}
}

func TestRevokeAllForUser(t *testing.T) {
	svc, repo := testSessionService()

	userID := uuid.New()
	tenantID := uuid.New()

	s1 := uuid.New()
	s2 := uuid.New()
	s3 := uuid.New()
	repo.sessions[s1] = &models.Session{ID: s1, UserID: userID, TenantID: tenantID}
	repo.sessions[s2] = &models.Session{ID: s2, UserID: userID, TenantID: tenantID}
	repo.sessions[s3] = &models.Session{ID: s3, UserID: uuid.New(), TenantID: tenantID}

	err := svc.RevokeAllForUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("RevokeAllForUser returned error: %v", err)
	}

	// Only s3 should remain
	if len(repo.sessions) != 1 {
		t.Errorf("expected 1 session remaining, got %d", len(repo.sessions))
	}
	if _, ok := repo.sessions[s3]; !ok {
		t.Error("other user's session should not be deleted")
	}
}

func TestRevokeAllForTenant(t *testing.T) {
	svc, repo := testSessionService()

	tenantID := uuid.New()
	otherTenantID := uuid.New()

	s1 := uuid.New()
	s2 := uuid.New()
	s3 := uuid.New()
	repo.sessions[s1] = &models.Session{ID: s1, UserID: uuid.New(), TenantID: tenantID}
	repo.sessions[s2] = &models.Session{ID: s2, UserID: uuid.New(), TenantID: tenantID}
	repo.sessions[s3] = &models.Session{ID: s3, UserID: uuid.New(), TenantID: otherTenantID}

	err := svc.RevokeAllForTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("RevokeAllForTenant returned error: %v", err)
	}

	if len(repo.sessions) != 1 {
		t.Errorf("expected 1 session remaining, got %d", len(repo.sessions))
	}
}

func TestValidate_ValidSession(t *testing.T) {
	svc, repo := testSessionService()

	sessionID := uuid.New()
	repo.sessions[sessionID] = &models.Session{
		ID:           sessionID,
		UserID:       uuid.New(),
		TenantID:     uuid.New(),
		ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		LastActiveAt: time.Now().UTC(),
	}

	session, err := svc.Validate(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if session.ID != sessionID {
		t.Errorf("ID = %v, want %v", session.ID, sessionID)
	}
}

func TestValidate_Expired(t *testing.T) {
	svc, repo := testSessionService()

	sessionID := uuid.New()
	repo.sessions[sessionID] = &models.Session{
		ID:           sessionID,
		UserID:       uuid.New(),
		TenantID:     uuid.New(),
		ExpiresAt:    time.Now().UTC().Add(-1 * time.Hour),
		LastActiveAt: time.Now().UTC().Add(-2 * time.Hour),
	}

	_, err := svc.Validate(context.Background(), sessionID)
	if err == nil {
		t.Error("Validate should fail for expired session")
	}
}

func TestValidate_Inactive(t *testing.T) {
	svc, repo := testSessionService()

	sessionID := uuid.New()
	repo.sessions[sessionID] = &models.Session{
		ID:           sessionID,
		UserID:       uuid.New(),
		TenantID:     uuid.New(),
		ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		LastActiveAt: time.Now().UTC().Add(-2 * time.Hour), // Inactive for 2 hours (timeout is 1 hour)
	}

	_, err := svc.Validate(context.Background(), sessionID)
	if err == nil {
		t.Error("Validate should fail for inactive session")
	}
}

func TestSessionRedisKey(t *testing.T) {
	id := uuid.New()
	key := sessionRedisKey(id)
	expected := fmt.Sprintf("session:%s", id.String())
	if key != expected {
		t.Errorf("sessionRedisKey = %q, want %q", key, expected)
	}
}

func TestUserSessionsKey(t *testing.T) {
	id := uuid.New()
	key := userSessionsKey(id)
	expected := fmt.Sprintf("user_sessions:%s", id.String())
	if key != expected {
		t.Errorf("userSessionsKey = %q, want %q", key, expected)
	}
}
