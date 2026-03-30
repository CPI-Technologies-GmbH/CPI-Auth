package flows

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"image/png"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// MFAService handles multi-factor authentication flows.
type MFAService struct {
	enrollments   models.MFAEnrollmentRepository
	recoveryCodes models.RecoveryCodeRepository
	cfg           *config.Config
	logger        *zap.Logger
}

// NewMFAService creates a new MFA service.
func NewMFAService(enrollments models.MFAEnrollmentRepository, codes models.RecoveryCodeRepository, cfg *config.Config, logger *zap.Logger) *MFAService {
	return &MFAService{
		enrollments:   enrollments,
		recoveryCodes: codes,
		cfg:           cfg,
		logger:        logger,
	}
}

// TOTPEnrollment holds the result of a TOTP enrollment initiation.
type TOTPEnrollment struct {
	EnrollmentID uuid.UUID `json:"enrollment_id"`
	Secret       string    `json:"secret"`
	URI          string    `json:"uri"`
	QRCode       string    `json:"qr_code"` // base64-encoded PNG
}

// EnrollTOTP begins the TOTP enrollment process.
func (s *MFAService) EnrollTOTP(ctx context.Context, userID uuid.UUID, email string) (*TOTPEnrollment, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CPI Auth",
		AccountName: email,
		Period:      30,
		SecretSize:  32,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("generating TOTP key: %w", err)
	}

	// Encrypt the secret for storage
	encrypted, err := crypto.Encrypt([]byte(key.Secret()), s.cfg.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting TOTP secret: %w", err)
	}

	enrollment := &models.MFAEnrollment{
		UserID:          userID,
		Method:          models.MFAMethodTOTP,
		SecretEncrypted: encrypted,
		Verified:        false,
	}

	if err := s.enrollments.Create(ctx, enrollment); err != nil {
		return nil, fmt.Errorf("creating MFA enrollment: %w", err)
	}

	// Generate QR code
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("generating QR code: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encoding QR code: %w", err)
	}
	qrBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	return &TOTPEnrollment{
		EnrollmentID: enrollment.ID,
		Secret:       key.Secret(),
		URI:          key.URL(),
		QRCode:       qrBase64,
	}, nil
}

// VerifyTOTP verifies a TOTP code and marks the enrollment as verified.
func (s *MFAService) VerifyTOTP(ctx context.Context, enrollmentID uuid.UUID, code string) error {
	enrollment, err := s.enrollments.GetByID(ctx, enrollmentID)
	if err != nil {
		return err
	}

	if enrollment.Method != models.MFAMethodTOTP {
		return models.ErrBadRequest.WithMessage("Enrollment is not TOTP.")
	}

	// Decrypt secret
	secret, err := crypto.Decrypt(enrollment.SecretEncrypted, s.cfg.Security.EncryptionKey)
	if err != nil {
		return fmt.Errorf("decrypting TOTP secret: %w", err)
	}

	// Validate the code
	valid := totp.Validate(code, string(secret))
	if !valid {
		return models.ErrMFAInvalidCode
	}

	// Mark as verified
	enrollment.Verified = true
	return s.enrollments.Update(ctx, enrollment)
}

// ValidateTOTP validates a TOTP code for an already-verified enrollment.
func (s *MFAService) ValidateTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	enrollments, err := s.enrollments.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, enrollment := range enrollments {
		if enrollment.Method != models.MFAMethodTOTP || !enrollment.Verified {
			continue
		}

		secret, err := crypto.Decrypt(enrollment.SecretEncrypted, s.cfg.Security.EncryptionKey)
		if err != nil {
			continue
		}

		if totp.Validate(code, string(secret)) {
			return nil
		}
	}

	return models.ErrMFAInvalidCode
}

// GenerateEmailOTP generates a 6-digit OTP for email-based MFA.
func (s *MFAService) GenerateEmailOTP(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate a 6-digit code using TOTP with a short period
	secret := make([]byte, 20)
	if _, err := crypto.GenerateRandomBytes(20); err != nil {
		return "", err
	}
	copy(secret, []byte(base32.StdEncoding.EncodeToString(secret)[:20]))

	code, err := totp.GenerateCodeCustom(base32.StdEncoding.EncodeToString(secret), time.Now(), totp.ValidateOpts{
		Period:    300, // 5 minute validity
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", fmt.Errorf("generating email OTP: %w", err)
	}

	// Store encrypted secret for later validation
	encrypted, err := crypto.Encrypt(secret, s.cfg.Security.EncryptionKey)
	if err != nil {
		return "", err
	}

	enrollment := &models.MFAEnrollment{
		UserID:          userID,
		Method:          models.MFAMethodEmail,
		SecretEncrypted: encrypted,
		Verified:        false,
	}
	if err := s.enrollments.Create(ctx, enrollment); err != nil {
		return "", err
	}

	return code, nil
}

// ListEnrollments returns all MFA enrollments for a user.
func (s *MFAService) ListEnrollments(ctx context.Context, userID uuid.UUID) ([]models.MFAEnrollment, error) {
	return s.enrollments.ListByUser(ctx, userID)
}

// DeleteEnrollment removes an MFA enrollment.
func (s *MFAService) DeleteEnrollment(ctx context.Context, enrollmentID uuid.UUID) error {
	return s.enrollments.Delete(ctx, enrollmentID)
}

// --- Recovery Codes ---

// GenerateRecoveryCodes creates a set of one-time recovery codes for a user.
func (s *MFAService) GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Delete existing codes first
	if err := s.recoveryCodes.DeleteByUser(ctx, userID); err != nil {
		return nil, err
	}

	codes := make([]string, 10)
	for i := range codes {
		code, err := crypto.GenerateRandomString(10)
		if err != nil {
			return nil, err
		}
		codes[i] = code

		rc := &models.RecoveryCode{
			UserID:   userID,
			CodeHash: crypto.HashToken(code),
			Used:     false,
		}
		if err := s.recoveryCodes.Create(ctx, rc); err != nil {
			return nil, err
		}
	}

	return codes, nil
}

// VerifyRecoveryCode verifies and consumes a recovery code.
func (s *MFAService) VerifyRecoveryCode(ctx context.Context, userID uuid.UUID, code string) error {
	hash := crypto.HashToken(code)
	rc, err := s.recoveryCodes.GetByUserAndHash(ctx, userID, hash)
	if err != nil {
		return models.ErrMFAInvalidCode
	}
	if rc.Used {
		return models.ErrMFAInvalidCode.WithMessage("Recovery code has already been used.")
	}
	return s.recoveryCodes.MarkUsed(ctx, rc.ID)
}

// ListRecoveryCodes returns remaining unused recovery codes.
func (s *MFAService) ListRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]models.RecoveryCode, error) {
	return s.recoveryCodes.ListByUser(ctx, userID)
}

// HasVerifiedMFA checks if a user has any verified MFA enrollments.
func (s *MFAService) HasVerifiedMFA(ctx context.Context, userID uuid.UUID) (bool, error) {
	enrollments, err := s.enrollments.ListByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, e := range enrollments {
		if e.Verified {
			return true, nil
		}
	}
	return false, nil
}
