package flows

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// --- Mock MFA Enrollment Repository ---

type mockMFAEnrollmentRepo struct {
	enrollments map[uuid.UUID]*models.MFAEnrollment
}

func newMockMFAEnrollmentRepo() *mockMFAEnrollmentRepo {
	return &mockMFAEnrollmentRepo{
		enrollments: make(map[uuid.UUID]*models.MFAEnrollment),
	}
}

func (m *mockMFAEnrollmentRepo) Create(_ context.Context, enrollment *models.MFAEnrollment) error {
	if enrollment.ID == uuid.Nil {
		enrollment.ID = uuid.New()
	}
	m.enrollments[enrollment.ID] = enrollment
	return nil
}

func (m *mockMFAEnrollmentRepo) GetByID(_ context.Context, id uuid.UUID) (*models.MFAEnrollment, error) {
	e, ok := m.enrollments[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return e, nil
}

func (m *mockMFAEnrollmentRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]models.MFAEnrollment, error) {
	var result []models.MFAEnrollment
	for _, e := range m.enrollments {
		if e.UserID == userID {
			result = append(result, *e)
		}
	}
	return result, nil
}

func (m *mockMFAEnrollmentRepo) Update(_ context.Context, enrollment *models.MFAEnrollment) error {
	m.enrollments[enrollment.ID] = enrollment
	return nil
}

func (m *mockMFAEnrollmentRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.enrollments, id)
	return nil
}

// --- Mock Recovery Code Repository ---

type mockRecoveryCodeRepo struct {
	codes map[uuid.UUID]*models.RecoveryCode
}

func newMockRecoveryCodeRepo() *mockRecoveryCodeRepo {
	return &mockRecoveryCodeRepo{
		codes: make(map[uuid.UUID]*models.RecoveryCode),
	}
}

func (m *mockRecoveryCodeRepo) Create(_ context.Context, code *models.RecoveryCode) error {
	if code.ID == uuid.Nil {
		code.ID = uuid.New()
	}
	m.codes[code.ID] = code
	return nil
}

func (m *mockRecoveryCodeRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]models.RecoveryCode, error) {
	var result []models.RecoveryCode
	for _, c := range m.codes {
		if c.UserID == userID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockRecoveryCodeRepo) MarkUsed(_ context.Context, id uuid.UUID) error {
	if c, ok := m.codes[id]; ok {
		c.Used = true
		return nil
	}
	return models.ErrNotFound
}

func (m *mockRecoveryCodeRepo) DeleteByUser(_ context.Context, userID uuid.UUID) error {
	for id, c := range m.codes {
		if c.UserID == userID {
			delete(m.codes, id)
		}
	}
	return nil
}

func (m *mockRecoveryCodeRepo) GetByUserAndHash(_ context.Context, userID uuid.UUID, codeHash string) (*models.RecoveryCode, error) {
	for _, c := range m.codes {
		if c.UserID == userID && c.CodeHash == codeHash {
			return c, nil
		}
	}
	return nil, models.ErrNotFound
}

// --- Test Helpers ---

// validEncryptionKey returns a 64 hex char string (32 bytes)
func validEncryptionKey() string {
	return "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}

func testConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Security.EncryptionKey = validEncryptionKey()
	return cfg
}

func testMFAService() (*MFAService, *mockMFAEnrollmentRepo, *mockRecoveryCodeRepo) {
	enrollmentRepo := newMockMFAEnrollmentRepo()
	recoveryRepo := newMockRecoveryCodeRepo()
	cfg := testConfig()
	svc := NewMFAService(enrollmentRepo, recoveryRepo, cfg, zap.NewNop())
	return svc, enrollmentRepo, recoveryRepo
}

// --- Tests ---

func TestNewMFAService(t *testing.T) {
	svc, _, _ := testMFAService()
	if svc == nil {
		t.Fatal("NewMFAService returned nil")
	}
}

func TestEnrollTOTP(t *testing.T) {
	svc, enrollmentRepo, _ := testMFAService()

	userID := uuid.New()
	email := "user@example.com"

	enrollment, err := svc.EnrollTOTP(context.Background(), userID, email)
	if err != nil {
		t.Fatalf("EnrollTOTP returned error: %v", err)
	}

	if enrollment.EnrollmentID == uuid.Nil {
		t.Error("EnrollmentID should not be nil")
	}
	if enrollment.Secret == "" {
		t.Error("Secret should not be empty")
	}
	if enrollment.URI == "" {
		t.Error("URI should not be empty")
	}
	if enrollment.QRCode == "" {
		t.Error("QRCode should not be empty")
	}

	// Verify enrollment was stored
	if len(enrollmentRepo.enrollments) != 1 {
		t.Errorf("expected 1 enrollment in repo, got %d", len(enrollmentRepo.enrollments))
	}

	// Check stored enrollment is not verified yet
	stored := enrollmentRepo.enrollments[enrollment.EnrollmentID]
	if stored.Verified {
		t.Error("new enrollment should not be verified")
	}
	if stored.Method != models.MFAMethodTOTP {
		t.Errorf("Method = %q, want %q", stored.Method, models.MFAMethodTOTP)
	}
	if stored.UserID != userID {
		t.Error("UserID mismatch")
	}
}

func TestVerifyTOTP_InvalidCode(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	enrollment, err := svc.EnrollTOTP(context.Background(), userID, "user@test.com")
	if err != nil {
		t.Fatalf("EnrollTOTP returned error: %v", err)
	}

	// Try with an invalid code
	err = svc.VerifyTOTP(context.Background(), enrollment.EnrollmentID, "000000")
	if err == nil {
		t.Error("VerifyTOTP should return error for invalid code")
	}
	if !models.IsAppError(err, models.ErrMFAInvalidCode) {
		t.Errorf("expected ErrMFAInvalidCode, got %v", err)
	}
}

func TestVerifyTOTP_NonExistentEnrollment(t *testing.T) {
	svc, _, _ := testMFAService()

	err := svc.VerifyTOTP(context.Background(), uuid.New(), "123456")
	if err == nil {
		t.Error("VerifyTOTP should return error for non-existent enrollment")
	}
}

func TestVerifyTOTP_WrongMethod(t *testing.T) {
	svc, enrollmentRepo, _ := testMFAService()

	// Create an enrollment with email method
	enrollmentID := uuid.New()
	enrollmentRepo.enrollments[enrollmentID] = &models.MFAEnrollment{
		ID:              enrollmentID,
		UserID:          uuid.New(),
		Method:          models.MFAMethodEmail,
		SecretEncrypted: []byte("some-encrypted-data"),
		Verified:        false,
	}

	err := svc.VerifyTOTP(context.Background(), enrollmentID, "123456")
	if err == nil {
		t.Error("VerifyTOTP should return error for non-TOTP enrollment")
	}
}

func TestValidateTOTP_NoVerifiedEnrollments(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	// Enroll but don't verify
	_, _ = svc.EnrollTOTP(context.Background(), userID, "user@test.com")

	err := svc.ValidateTOTP(context.Background(), userID, "123456")
	if err == nil {
		t.Error("ValidateTOTP should fail when no verified enrollments exist")
	}
	if !models.IsAppError(err, models.ErrMFAInvalidCode) {
		t.Errorf("expected ErrMFAInvalidCode, got %v", err)
	}
}

func TestGenerateRecoveryCodes(t *testing.T) {
	svc, _, recoveryRepo := testMFAService()

	userID := uuid.New()
	codes, err := svc.GenerateRecoveryCodes(context.Background(), userID)
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes returned error: %v", err)
	}

	if len(codes) != 10 {
		t.Errorf("expected 10 recovery codes, got %d", len(codes))
	}

	// Ensure all codes are unique
	codeSet := make(map[string]bool)
	for _, code := range codes {
		if code == "" {
			t.Error("recovery code should not be empty")
		}
		if len(code) != 10 {
			t.Errorf("recovery code length = %d, want 10", len(code))
		}
		if codeSet[code] {
			t.Errorf("duplicate recovery code: %s", code)
		}
		codeSet[code] = true
	}

	// Verify codes were stored
	if len(recoveryRepo.codes) != 10 {
		t.Errorf("expected 10 codes in repo, got %d", len(recoveryRepo.codes))
	}
}

func TestGenerateRecoveryCodes_DeletesExisting(t *testing.T) {
	svc, _, recoveryRepo := testMFAService()

	userID := uuid.New()

	// Generate first batch
	_, err := svc.GenerateRecoveryCodes(context.Background(), userID)
	if err != nil {
		t.Fatalf("first GenerateRecoveryCodes returned error: %v", err)
	}
	if len(recoveryRepo.codes) != 10 {
		t.Fatalf("expected 10 codes after first generation, got %d", len(recoveryRepo.codes))
	}

	// Generate second batch - should delete old ones
	_, err = svc.GenerateRecoveryCodes(context.Background(), userID)
	if err != nil {
		t.Fatalf("second GenerateRecoveryCodes returned error: %v", err)
	}
	if len(recoveryRepo.codes) != 10 {
		t.Errorf("expected 10 codes after regeneration, got %d", len(recoveryRepo.codes))
	}
}

func TestVerifyRecoveryCode(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	codes, err := svc.GenerateRecoveryCodes(context.Background(), userID)
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes returned error: %v", err)
	}

	// Verify valid code
	err = svc.VerifyRecoveryCode(context.Background(), userID, codes[0])
	if err != nil {
		t.Fatalf("VerifyRecoveryCode returned error: %v", err)
	}
}

func TestVerifyRecoveryCode_UsedCode(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	codes, _ := svc.GenerateRecoveryCodes(context.Background(), userID)

	// Use the code
	err := svc.VerifyRecoveryCode(context.Background(), userID, codes[0])
	if err != nil {
		t.Fatalf("first VerifyRecoveryCode returned error: %v", err)
	}

	// Try to use it again
	err = svc.VerifyRecoveryCode(context.Background(), userID, codes[0])
	if err == nil {
		t.Error("VerifyRecoveryCode should fail for already-used code")
	}
}

func TestVerifyRecoveryCode_InvalidCode(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	_, _ = svc.GenerateRecoveryCodes(context.Background(), userID)

	err := svc.VerifyRecoveryCode(context.Background(), userID, "invalid-code-xxx")
	if err == nil {
		t.Error("VerifyRecoveryCode should fail for invalid code")
	}
	if !models.IsAppError(err, models.ErrMFAInvalidCode) {
		t.Errorf("expected ErrMFAInvalidCode, got %v", err)
	}
}

func TestVerifyRecoveryCode_WrongUser(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	codes, _ := svc.GenerateRecoveryCodes(context.Background(), userID)

	// Try with a different user
	err := svc.VerifyRecoveryCode(context.Background(), uuid.New(), codes[0])
	if err == nil {
		t.Error("VerifyRecoveryCode should fail for wrong user")
	}
}

func TestListEnrollments(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	_, _ = svc.EnrollTOTP(context.Background(), userID, "user@test.com")

	enrollments, err := svc.ListEnrollments(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListEnrollments returned error: %v", err)
	}
	if len(enrollments) != 1 {
		t.Errorf("expected 1 enrollment, got %d", len(enrollments))
	}
}

func TestListEnrollments_Empty(t *testing.T) {
	svc, _, _ := testMFAService()

	enrollments, err := svc.ListEnrollments(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListEnrollments returned error: %v", err)
	}
	if len(enrollments) != 0 {
		t.Errorf("expected 0 enrollments, got %d", len(enrollments))
	}
}

func TestDeleteEnrollment(t *testing.T) {
	svc, enrollmentRepo, _ := testMFAService()

	userID := uuid.New()
	enrollment, _ := svc.EnrollTOTP(context.Background(), userID, "user@test.com")

	err := svc.DeleteEnrollment(context.Background(), enrollment.EnrollmentID)
	if err != nil {
		t.Fatalf("DeleteEnrollment returned error: %v", err)
	}

	if len(enrollmentRepo.enrollments) != 0 {
		t.Error("enrollment should have been deleted")
	}
}

func TestHasVerifiedMFA_NoEnrollments(t *testing.T) {
	svc, _, _ := testMFAService()

	has, err := svc.HasVerifiedMFA(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("HasVerifiedMFA returned error: %v", err)
	}
	if has {
		t.Error("should return false when no enrollments exist")
	}
}

func TestHasVerifiedMFA_UnverifiedOnly(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	_, _ = svc.EnrollTOTP(context.Background(), userID, "user@test.com")

	has, err := svc.HasVerifiedMFA(context.Background(), userID)
	if err != nil {
		t.Fatalf("HasVerifiedMFA returned error: %v", err)
	}
	if has {
		t.Error("should return false when only unverified enrollments exist")
	}
}

func TestHasVerifiedMFA_WithVerified(t *testing.T) {
	svc, enrollmentRepo, _ := testMFAService()

	userID := uuid.New()

	// Create a verified enrollment directly
	secret := []byte("test-secret-12345678901234567890")
	encrypted, err := crypto.Encrypt(secret, validEncryptionKey())
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	enrollmentID := uuid.New()
	enrollmentRepo.enrollments[enrollmentID] = &models.MFAEnrollment{
		ID:              enrollmentID,
		UserID:          userID,
		Method:          models.MFAMethodTOTP,
		SecretEncrypted: encrypted,
		Verified:        true,
	}

	has, err := svc.HasVerifiedMFA(context.Background(), userID)
	if err != nil {
		t.Fatalf("HasVerifiedMFA returned error: %v", err)
	}
	if !has {
		t.Error("should return true when a verified enrollment exists")
	}
}

func TestListRecoveryCodes(t *testing.T) {
	svc, _, _ := testMFAService()

	userID := uuid.New()
	_, _ = svc.GenerateRecoveryCodes(context.Background(), userID)

	codes, err := svc.ListRecoveryCodes(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListRecoveryCodes returned error: %v", err)
	}
	if len(codes) != 10 {
		t.Errorf("expected 10 recovery codes, got %d", len(codes))
	}
}

func TestListRecoveryCodes_Empty(t *testing.T) {
	svc, _, _ := testMFAService()

	codes, err := svc.ListRecoveryCodes(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListRecoveryCodes returned error: %v", err)
	}
	if len(codes) != 0 {
		t.Errorf("expected 0 recovery codes, got %d", len(codes))
	}
}
