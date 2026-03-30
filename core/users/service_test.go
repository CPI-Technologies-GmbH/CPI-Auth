package users

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// --- Mock User Repository ---

type mockUserRepo struct {
	users           map[uuid.UUID]*models.User
	passwordHistory map[uuid.UUID][]models.PasswordHistory
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:           make(map[uuid.UUID]*models.User),
		passwordHistory: make(map[uuid.UUID][]models.PasswordHistory),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	u, ok := m.users[id]
	if !ok || u.TenantID != tenantID {
		return nil, models.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.TenantID == tenantID && u.Email == email {
			return u, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *mockUserRepo) Update(_ context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	if u, ok := m.users[id]; ok && u.TenantID == tenantID {
		delete(m.users, id)
		return nil
	}
	return models.ErrNotFound
}

func (m *mockUserRepo) List(_ context.Context, tenantID uuid.UUID, params models.PaginationParams, search string) (*models.PaginatedResult[models.User], error) {
	var result []models.User
	for _, u := range m.users {
		if u.TenantID == tenantID {
			result = append(result, *u)
		}
	}
	return &models.PaginatedResult[models.User]{Data: result, Total: int64(len(result)), Page: params.Page, PerPage: params.PerPage}, nil
}

func (m *mockUserRepo) Block(_ context.Context, tenantID, id uuid.UUID) error {
	if u, ok := m.users[id]; ok && u.TenantID == tenantID {
		u.Status = models.StatusBlocked
		return nil
	}
	return models.ErrNotFound
}

func (m *mockUserRepo) Unblock(_ context.Context, tenantID, id uuid.UUID) error {
	if u, ok := m.users[id]; ok && u.TenantID == tenantID {
		u.Status = models.StatusActive
		return nil
	}
	return models.ErrNotFound
}

func (m *mockUserRepo) CountByTenant(_ context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	for _, u := range m.users {
		if u.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

func (m *mockUserRepo) GetPasswordHistory(_ context.Context, userID uuid.UUID, limit int) ([]models.PasswordHistory, error) {
	history := m.passwordHistory[userID]
	if len(history) > limit {
		return history[:limit], nil
	}
	return history, nil
}

func (m *mockUserRepo) AddPasswordHistory(_ context.Context, entry *models.PasswordHistory) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	entry.CreatedAt = time.Now()
	m.passwordHistory[entry.UserID] = append(m.passwordHistory[entry.UserID], *entry)
	return nil
}

// --- Mock Identity Repository ---

type mockIdentityRepo struct {
	identities map[uuid.UUID]*models.Identity
}

func newMockIdentityRepo() *mockIdentityRepo {
	return &mockIdentityRepo{
		identities: make(map[uuid.UUID]*models.Identity),
	}
}

func (m *mockIdentityRepo) Create(_ context.Context, identity *models.Identity) error {
	if identity.ID == uuid.Nil {
		identity.ID = uuid.New()
	}
	m.identities[identity.ID] = identity
	return nil
}

func (m *mockIdentityRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Identity, error) {
	i, ok := m.identities[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return i, nil
}

func (m *mockIdentityRepo) GetByProvider(_ context.Context, provider, providerUserID string) (*models.Identity, error) {
	for _, i := range m.identities {
		if i.Provider == provider && i.ProviderUserID == providerUserID {
			return i, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *mockIdentityRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Identity, error) {
	var result []models.Identity
	for _, i := range m.identities {
		if i.UserID == userID {
			result = append(result, *i)
		}
	}
	return result, nil
}

func (m *mockIdentityRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.identities, id)
	return nil
}

func (m *mockIdentityRepo) Update(_ context.Context, identity *models.Identity) error {
	m.identities[identity.ID] = identity
	return nil
}

// --- Test Helpers ---

func testUserService() (*Service, *mockUserRepo, *mockIdentityRepo) {
	userRepo := newMockUserRepo()
	identityRepo := newMockIdentityRepo()
	cfg := config.DefaultConfig()
	cfg.Security.HIBPEnabled = false // Disable HIBP in tests
	logger := zap.NewNop()

	svc := NewService(userRepo, identityRepo, cfg, logger)
	return svc, userRepo, identityRepo
}

// --- Tests ---

func TestNewUserService(t *testing.T) {
	svc, _, _ := testUserService()
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestRegister(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "test@example.com",
		Password: "SecureP@ss1",
		Name:     "Test User",
		Phone:    "+1234567890",
	}

	user, err := svc.Register(context.Background(), tenantID, input)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if user.ID == uuid.Nil {
		t.Error("user ID should be assigned")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "test@example.com")
	}
	if user.Name != "Test User" {
		t.Errorf("Name = %q, want %q", user.Name, "Test User")
	}
	if user.Status != models.StatusActive {
		t.Errorf("Status = %q, want %q", user.Status, models.StatusActive)
	}
	if user.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", user.TenantID, tenantID)
	}
	if user.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}
	if user.PasswordHash == input.Password {
		t.Error("PasswordHash should not be the raw password")
	}

	// Verify stored in repo
	if len(userRepo.users) != 1 {
		t.Errorf("expected 1 user in repo, got %d", len(userRepo.users))
	}

	// Verify password history was stored
	if len(userRepo.passwordHistory[user.ID]) != 1 {
		t.Errorf("expected 1 password history entry, got %d", len(userRepo.passwordHistory[user.ID]))
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "dupe@example.com",
		Password: "SecureP@ss1",
	}

	_, err := svc.Register(context.Background(), tenantID, input)
	if err != nil {
		t.Fatalf("first Register returned error: %v", err)
	}

	_, err = svc.Register(context.Background(), tenantID, input)
	if err == nil {
		t.Error("second Register should fail for duplicate email")
	}
	if !models.IsAppError(err, models.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestRegister_NormalizesEmail(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "  TEST@Example.COM  ",
		Password: "SecureP@ss1",
	}

	user, err := svc.Register(context.Background(), tenantID, input)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q (lowered and trimmed)", user.Email, "test@example.com")
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	tests := []struct {
		name     string
		password string
	}{
		{"too short", "Ab1"},
		{"no uppercase", "secure123password"},
		{"no lowercase", "SECURE123PASSWORD"},
		{"no digit", "SecurePassword"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RegisterInput{
				Email:    tt.name + "@example.com",
				Password: tt.password,
			}

			_, err := svc.Register(context.Background(), tenantID, input)
			if err == nil {
				t.Error("Register should fail for weak password")
			}
			if !models.IsAppError(err, models.ErrPasswordPolicy) {
				t.Errorf("expected ErrPasswordPolicy, got %v", err)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	password := "SecureP@ss1"
	input := RegisterInput{
		Email:    "login@example.com",
		Password: password,
	}

	registered, err := svc.Register(context.Background(), tenantID, input)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	user, err := svc.Authenticate(context.Background(), tenantID, "login@example.com", password)
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	if user.ID != registered.ID {
		t.Errorf("user ID = %v, want %v", user.ID, registered.ID)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "wrongpass@example.com",
		Password: "SecureP@ss1",
	}

	_, _ = svc.Register(context.Background(), tenantID, input)

	_, err := svc.Authenticate(context.Background(), tenantID, "wrongpass@example.com", "WrongPassword1")
	if err == nil {
		t.Error("Authenticate should fail with wrong password")
	}
	if !models.IsAppError(err, models.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticate_NonExistentUser(t *testing.T) {
	svc, _, _ := testUserService()

	_, err := svc.Authenticate(context.Background(), uuid.New(), "nonexistent@example.com", "password")
	if err == nil {
		t.Error("Authenticate should fail for non-existent user")
	}
	if !models.IsAppError(err, models.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticate_BlockedUser(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "blocked@example.com",
		Password: "SecureP@ss1",
	}

	user, _ := svc.Register(context.Background(), tenantID, input)

	// Block the user
	userRepo.users[user.ID].Status = models.StatusBlocked

	_, err := svc.Authenticate(context.Background(), tenantID, "blocked@example.com", "SecureP@ss1")
	if err == nil {
		t.Error("Authenticate should fail for blocked user")
	}
	if !models.IsAppError(err, models.ErrAccountBlocked) {
		t.Errorf("expected ErrAccountBlocked, got %v", err)
	}
}

func TestAuthenticate_DeletedUser(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "deleted@example.com",
		Password: "SecureP@ss1",
	}

	user, _ := svc.Register(context.Background(), tenantID, input)
	userRepo.users[user.ID].Status = models.StatusDeleted

	_, err := svc.Authenticate(context.Background(), tenantID, "deleted@example.com", "SecureP@ss1")
	if err == nil {
		t.Error("Authenticate should fail for deleted user")
	}
	if !models.IsAppError(err, models.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	oldPassword := "OldSecure1!"
	newPassword := "NewSecure2@"

	input := RegisterInput{
		Email:    "change@example.com",
		Password: oldPassword,
	}

	user, _ := svc.Register(context.Background(), tenantID, input)
	oldHash := user.PasswordHash

	err := svc.ChangePassword(context.Background(), tenantID, user.ID, oldPassword, newPassword)
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}

	// Verify password was changed
	updated := userRepo.users[user.ID]
	if updated.PasswordHash == oldHash {
		t.Error("password hash should have changed")
	}

	// Verify can authenticate with new password
	_, err = svc.Authenticate(context.Background(), tenantID, "change@example.com", newPassword)
	if err != nil {
		t.Errorf("should be able to authenticate with new password: %v", err)
	}
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "changefail@example.com",
		Password: "OldSecure1!",
	}

	user, _ := svc.Register(context.Background(), tenantID, input)

	err := svc.ChangePassword(context.Background(), tenantID, user.ID, "WrongOld1!", "NewSecure2@")
	if err == nil {
		t.Error("ChangePassword should fail with wrong old password")
	}
	if !models.IsAppError(err, models.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePassword_WeakNewPassword(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "weaknew@example.com",
		Password: "OldSecure1!",
	}

	user, _ := svc.Register(context.Background(), tenantID, input)

	err := svc.ChangePassword(context.Background(), tenantID, user.ID, "OldSecure1!", "weak")
	if err == nil {
		t.Error("ChangePassword should fail for weak new password")
	}
}

func TestChangePassword_PasswordReuse(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	password := "SecureP@ss1"
	input := RegisterInput{
		Email:    "reuse@example.com",
		Password: password,
	}

	user, _ := svc.Register(context.Background(), tenantID, input)

	// Change to a new password first
	newPass := "NewSecure2@"
	err := svc.ChangePassword(context.Background(), tenantID, user.ID, password, newPass)
	if err != nil {
		t.Fatalf("first ChangePassword returned error: %v", err)
	}

	// Try to change back to the original password (reuse)
	err = svc.ChangePassword(context.Background(), tenantID, user.ID, newPass, password)
	if err == nil {
		t.Error("ChangePassword should detect password reuse")
	}
	if !models.IsAppError(err, models.ErrPasswordReused) {
		t.Errorf("expected ErrPasswordReused, got %v", err)
	}
}

func TestResetPassword(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "reset@example.com",
		Password: "OldSecure1!",
	}

	user, _ := svc.Register(context.Background(), tenantID, input)

	newPassword := "ResetSecure2@"
	err := svc.ResetPassword(context.Background(), tenantID, user.ID, newPassword)
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}

	// Should be able to authenticate with new password
	_, err = svc.Authenticate(context.Background(), tenantID, "reset@example.com", newPassword)
	if err != nil {
		t.Errorf("should authenticate with reset password: %v", err)
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "resetweak@example.com",
		Password: "OldSecure1!",
	}

	user, _ := svc.Register(context.Background(), tenantID, input)

	err := svc.ResetPassword(context.Background(), tenantID, user.ID, "weak")
	if err == nil {
		t.Error("ResetPassword should fail for weak password")
	}
}

func TestVerifyEmail(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "verify@example.com",
		Password: "SecureP@ss1",
	}

	user, _ := svc.Register(context.Background(), tenantID, input)

	if userRepo.users[user.ID].EmailVerified {
		t.Error("email should not be verified initially")
	}

	err := svc.VerifyEmail(context.Background(), tenantID, user.ID)
	if err != nil {
		t.Fatalf("VerifyEmail returned error: %v", err)
	}

	if !userRepo.users[user.ID].EmailVerified {
		t.Error("email should be verified after VerifyEmail")
	}
}

func TestGeneratePasswordResetToken(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "resettoken@example.com",
		Password: "SecureP@ss1",
	}
	_, _ = svc.Register(context.Background(), tenantID, input)

	token, err := svc.GeneratePasswordResetToken(context.Background(), tenantID, "resettoken@example.com")
	if err != nil {
		t.Fatalf("GeneratePasswordResetToken returned error: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty for existing user")
	}
}

func TestGeneratePasswordResetToken_NonExistentEmail(t *testing.T) {
	svc, _, _ := testUserService()

	// Should not reveal that the email doesn't exist
	token, err := svc.GeneratePasswordResetToken(context.Background(), uuid.New(), "nonexistent@example.com")
	if err != nil {
		t.Fatalf("GeneratePasswordResetToken returned error: %v", err)
	}
	// Returns empty token but no error (security: don't reveal user existence)
	if token != "" {
		t.Error("token should be empty for non-existent email")
	}
}

func TestCheckPasswordPolicy(t *testing.T) {
	settings := models.DefaultTenantSettings()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "SecureP@ss1", false},
		{"too short", "Ab1", true},
		{"no uppercase", "securepassword1", true},
		{"no lowercase", "SECUREPASSWORD1", true},
		{"no digit", "SecurePassword", true},
		{"exactly min length", "Secure1a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordPolicy(tt.password, settings)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCheckPasswordPolicy_SymbolRequired(t *testing.T) {
	settings := models.DefaultTenantSettings()
	settings.PasswordRequireSymbol = true

	err := validatePasswordPolicy("SecurePass1", settings)
	if err == nil {
		t.Error("should require symbol when PasswordRequireSymbol is true")
	}

	err = validatePasswordPolicy("SecureP@ss1", settings)
	if err != nil {
		t.Errorf("password with symbol should pass: %v", err)
	}
}

func TestGetByID(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "getbyid@example.com",
		Password: "SecureP@ss1",
	}

	created, _ := svc.Register(context.Background(), tenantID, input)

	user, err := svc.GetByID(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if user.Email != "getbyid@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "getbyid@example.com")
	}
}

func TestGetByEmail(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "getbyemail@example.com",
		Password: "SecureP@ss1",
	}
	_, _ = svc.Register(context.Background(), tenantID, input)

	user, err := svc.GetByEmail(context.Background(), tenantID, "getbyemail@example.com")
	if err != nil {
		t.Fatalf("GetByEmail returned error: %v", err)
	}
	if user.Email != "getbyemail@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "getbyemail@example.com")
	}
}

func TestDelete(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "delete@example.com",
		Password: "SecureP@ss1",
	}
	user, _ := svc.Register(context.Background(), tenantID, input)

	err := svc.Delete(context.Background(), tenantID, user.ID)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if len(userRepo.users) != 0 {
		t.Error("user should have been deleted")
	}
}

func TestBlock(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "block@example.com",
		Password: "SecureP@ss1",
	}
	user, _ := svc.Register(context.Background(), tenantID, input)

	err := svc.Block(context.Background(), tenantID, user.ID)
	if err != nil {
		t.Fatalf("Block returned error: %v", err)
	}
	if userRepo.users[user.ID].Status != models.StatusBlocked {
		t.Error("user should be blocked")
	}
}

func TestUnblock(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "unblock@example.com",
		Password: "SecureP@ss1",
	}
	user, _ := svc.Register(context.Background(), tenantID, input)

	_ = svc.Block(context.Background(), tenantID, user.ID)
	err := svc.Unblock(context.Background(), tenantID, user.ID)
	if err != nil {
		t.Fatalf("Unblock returned error: %v", err)
	}
	if userRepo.users[user.ID].Status != models.StatusActive {
		t.Error("user should be active after unblock")
	}
}

func TestList(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	for i := 0; i < 3; i++ {
		input := RegisterInput{
			Email:    crypto.HashToken(uuid.New().String())[:10] + "@example.com",
			Password: "SecureP@ss1",
		}
		_, _ = svc.Register(context.Background(), tenantID, input)
	}

	result, err := svc.List(context.Background(), tenantID, models.PaginationParams{Page: 1, PerPage: 10}, "")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Total = %d, want 3", result.Total)
	}
}

func TestLinkIdentity(t *testing.T) {
	svc, _, identityRepo := testUserService()

	userID := uuid.New()
	identity := &models.Identity{
		UserID:         userID,
		Provider:       "google",
		ProviderUserID: "google-123",
	}

	err := svc.LinkIdentity(context.Background(), identity)
	if err != nil {
		t.Fatalf("LinkIdentity returned error: %v", err)
	}
	if len(identityRepo.identities) != 1 {
		t.Error("identity should have been created")
	}
}

func TestUnlinkIdentity(t *testing.T) {
	svc, _, identityRepo := testUserService()

	identityID := uuid.New()
	identityRepo.identities[identityID] = &models.Identity{
		ID:             identityID,
		UserID:         uuid.New(),
		Provider:       "github",
		ProviderUserID: "gh-456",
	}

	err := svc.UnlinkIdentity(context.Background(), identityID)
	if err != nil {
		t.Fatalf("UnlinkIdentity returned error: %v", err)
	}
	if len(identityRepo.identities) != 0 {
		t.Error("identity should have been deleted")
	}
}

func TestListIdentities(t *testing.T) {
	svc, _, identityRepo := testUserService()

	userID := uuid.New()
	identityRepo.identities[uuid.New()] = &models.Identity{ID: uuid.New(), UserID: userID, Provider: "google"}
	identityRepo.identities[uuid.New()] = &models.Identity{ID: uuid.New(), UserID: userID, Provider: "github"}

	identities, err := svc.ListIdentities(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListIdentities returned error: %v", err)
	}
	if len(identities) != 2 {
		t.Errorf("expected 2 identities, got %d", len(identities))
	}
}

func TestExportUserData(t *testing.T) {
	svc, _, identityRepo := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "export@example.com",
		Password: "SecureP@ss1",
		Name:     "Export User",
	}
	user, _ := svc.Register(context.Background(), tenantID, input)

	identityRepo.identities[uuid.New()] = &models.Identity{
		ID: uuid.New(), UserID: user.ID, Provider: "google", ProviderUserID: "g-123",
	}

	data, err := svc.ExportUserData(context.Background(), tenantID, user.ID)
	if err != nil {
		t.Fatalf("ExportUserData returned error: %v", err)
	}

	if data["user"] == nil {
		t.Error("export should include user data")
	}
	if data["identities"] == nil {
		t.Error("export should include identities")
	}
	if data["exported_at"] == nil {
		t.Error("export should include exported_at timestamp")
	}
}

func TestDeactivate(t *testing.T) {
	svc, userRepo, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "deactivate@example.com",
		Password: "SecureP@ss1",
	}
	user, _ := svc.Register(context.Background(), tenantID, input)

	err := svc.Deactivate(context.Background(), tenantID, user.ID)
	if err != nil {
		t.Fatalf("Deactivate returned error: %v", err)
	}
	if userRepo.users[user.ID].Status != models.StatusInactive {
		t.Error("user should be inactive after deactivation")
	}
}

func TestGenerateEmailVerificationToken(t *testing.T) {
	svc, _, _ := testUserService()

	tenantID := uuid.New()
	input := RegisterInput{
		Email:    "emailverify@example.com",
		Password: "SecureP@ss1",
	}
	user, _ := svc.Register(context.Background(), tenantID, input)

	token, err := svc.GenerateEmailVerificationToken(context.Background(), tenantID, user.ID)
	if err != nil {
		t.Fatalf("GenerateEmailVerificationToken returned error: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
}
