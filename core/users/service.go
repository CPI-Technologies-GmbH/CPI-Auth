package users

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Service handles user management operations.
type Service struct {
	users    models.UserRepository
	identities models.IdentityRepository
	cfg      *config.Config
	logger   *zap.Logger
	argon2   crypto.Argon2Params
}

// NewService creates a new user service.
func NewService(users models.UserRepository, identities models.IdentityRepository, cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		users:    users,
		identities: identities,
		cfg:      cfg,
		logger:   logger,
		argon2: crypto.Argon2Params{
			Time:    cfg.Security.Argon2Time,
			Memory:  cfg.Security.Argon2Memory,
			Threads: cfg.Security.Argon2Threads,
			KeyLen:  cfg.Security.Argon2KeyLen,
			SaltLen: cfg.Security.Argon2SaltLen,
		},
	}
}

// RegisterInput holds registration request data.
type RegisterInput struct {
	Email    string          `json:"email" validate:"required,email"`
	Password string          `json:"password" validate:"required,min=8"`
	Name     string          `json:"name"`
	Phone    string          `json:"phone"`
	Locale   string          `json:"locale"`
	Metadata json.RawMessage `json:"metadata"`
}

// Register creates a new user account with password validation.
func (s *Service) Register(ctx context.Context, tenantID uuid.UUID, input RegisterInput) (*models.User, error) {
	// Check if user already exists
	existing, err := s.users.GetByEmail(ctx, tenantID, input.Email)
	if err != nil && !models.IsAppError(err, models.ErrNotFound) {
		return nil, models.ErrInternal.Wrap(err)
	}
	if existing != nil {
		return nil, models.ErrConflict.WithMessage("A user with this email already exists.")
	}

	// Validate password policy
	settings := s.getTenantSettings(ctx, tenantID)
	if err := validatePasswordPolicy(input.Password, settings); err != nil {
		return nil, err
	}

	// Check HIBP if enabled
	if s.cfg.Security.HIBPEnabled {
		breached, checkErr := checkHIBP(input.Password)
		if checkErr != nil {
			s.logger.Warn("HIBP check failed", zap.Error(checkErr))
		} else if breached {
			return nil, models.ErrPasswordBreached
		}
	}

	// Hash password
	hash, err := crypto.HashPassword(input.Password, s.argon2)
	if err != nil {
		return nil, models.ErrInternal.Wrap(err)
	}

	user := &models.User{
		TenantID:     tenantID,
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		Phone:        input.Phone,
		Name:         input.Name,
		PasswordHash: hash,
		Locale:       input.Locale,
		Metadata:     input.Metadata,
		Status:       models.StatusActive,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, models.ErrInternal.Wrap(err)
	}

	// Store initial password in history
	_ = s.users.AddPasswordHistory(ctx, &models.PasswordHistory{
		UserID: user.ID,
		Hash:   hash,
	})

	s.logger.Info("user registered",
		zap.String("user_id", user.ID.String()),
		zap.String("tenant_id", tenantID.String()),
		zap.String("email", user.Email),
	)

	return user, nil
}

// Authenticate verifies email/password credentials and returns the user.
func (s *Service) Authenticate(ctx context.Context, tenantID uuid.UUID, email, password string) (*models.User, error) {
	user, err := s.users.GetByEmail(ctx, tenantID, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if models.IsAppError(err, models.ErrNotFound) {
			return nil, models.ErrInvalidCredentials
		}
		return nil, models.ErrInternal.Wrap(err)
	}

	if user.Status == models.StatusBlocked {
		return nil, models.ErrAccountBlocked
	}

	if user.Status == models.StatusDeleted {
		return nil, models.ErrInvalidCredentials
	}

	match, err := crypto.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, models.ErrInternal.Wrap(err)
	}
	if !match {
		return nil, models.ErrInvalidCredentials
	}

	return user, nil
}

// ChangePassword changes a user's password after verifying the old one.
func (s *Service) ChangePassword(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	// Verify old password
	match, err := crypto.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil {
		return models.ErrInternal.Wrap(err)
	}
	if !match {
		return models.ErrInvalidCredentials
	}

	// Validate new password
	settings := s.getTenantSettings(ctx, tenantID)
	if err := validatePasswordPolicy(newPassword, settings); err != nil {
		return err
	}

	// Check password history for reuse
	if settings.PasswordHistoryCount > 0 {
		history, err := s.users.GetPasswordHistory(ctx, userID, settings.PasswordHistoryCount)
		if err != nil {
			return models.ErrInternal.Wrap(err)
		}
		for _, h := range history {
			match, _ := crypto.VerifyPassword(newPassword, h.Hash)
			if match {
				return models.ErrPasswordReused
			}
		}
	}

	// Check HIBP
	if s.cfg.Security.HIBPEnabled {
		breached, checkErr := checkHIBP(newPassword)
		if checkErr != nil {
			s.logger.Warn("HIBP check failed", zap.Error(checkErr))
		} else if breached {
			return models.ErrPasswordBreached
		}
	}

	hash, err := crypto.HashPassword(newPassword, s.argon2)
	if err != nil {
		return models.ErrInternal.Wrap(err)
	}

	user.PasswordHash = hash
	if err := s.users.Update(ctx, user); err != nil {
		return models.ErrInternal.Wrap(err)
	}

	_ = s.users.AddPasswordHistory(ctx, &models.PasswordHistory{
		UserID: userID,
		Hash:   hash,
	})

	return nil
}

// GeneratePasswordResetToken creates a secure password reset token.
func (s *Service) GeneratePasswordResetToken(ctx context.Context, tenantID uuid.UUID, email string) (string, error) {
	_, err := s.users.GetByEmail(ctx, tenantID, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		// Don't reveal whether the email exists
		if models.IsAppError(err, models.ErrNotFound) {
			return "", nil
		}
		return "", models.ErrInternal.Wrap(err)
	}

	token, err := crypto.GenerateOpaqueToken()
	if err != nil {
		return "", models.ErrInternal.Wrap(err)
	}

	return token, nil
}

// ResetPassword sets a new password using a reset token.
func (s *Service) ResetPassword(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, newPassword string) error {
	user, err := s.users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	settings := s.getTenantSettings(ctx, tenantID)
	if err := validatePasswordPolicy(newPassword, settings); err != nil {
		return err
	}

	hash, err := crypto.HashPassword(newPassword, s.argon2)
	if err != nil {
		return models.ErrInternal.Wrap(err)
	}

	user.PasswordHash = hash
	if err := s.users.Update(ctx, user); err != nil {
		return models.ErrInternal.Wrap(err)
	}

	_ = s.users.AddPasswordHistory(ctx, &models.PasswordHistory{
		UserID: userID,
		Hash:   hash,
	})

	return nil
}

// VerifyEmail marks the user's email as verified.
func (s *Service) VerifyEmail(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	user.EmailVerified = true
	return s.users.Update(ctx, user)
}

// GenerateEmailVerificationToken creates a token for email verification.
func (s *Service) GenerateEmailVerificationToken(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (string, error) {
	_, err := s.users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return "", err
	}
	token, err := crypto.GenerateOpaqueToken()
	if err != nil {
		return "", models.ErrInternal.Wrap(err)
	}
	return token, nil
}

// GetByID retrieves a user by ID.
func (s *Service) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*models.User, error) {
	return s.users.GetByID(ctx, tenantID, userID)
}

// GetByEmail retrieves a user by email.
func (s *Service) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	return s.users.GetByEmail(ctx, tenantID, strings.ToLower(strings.TrimSpace(email)))
}

// Update updates user profile fields.
func (s *Service) Update(ctx context.Context, user *models.User) error {
	return s.users.Update(ctx, user)
}

// Delete permanently deletes a user (GDPR support).
func (s *Service) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.users.Delete(ctx, tenantID, userID)
}

// Deactivate marks a user as inactive.
func (s *Service) Deactivate(ctx context.Context, tenantID, userID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	user.Status = models.StatusInactive
	return s.users.Update(ctx, user)
}

// Block blocks a user account.
func (s *Service) Block(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.users.Block(ctx, tenantID, userID)
}

// Unblock unblocks a user account.
func (s *Service) Unblock(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.users.Unblock(ctx, tenantID, userID)
}

// List returns paginated users for a tenant.
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams, search string) (*models.PaginatedResult[models.User], error) {
	return s.users.List(ctx, tenantID, params, search)
}

// LinkIdentity links an external identity to a user.
func (s *Service) LinkIdentity(ctx context.Context, identity *models.Identity) error {
	return s.identities.Create(ctx, identity)
}

// UnlinkIdentity removes an external identity from a user.
func (s *Service) UnlinkIdentity(ctx context.Context, identityID uuid.UUID) error {
	return s.identities.Delete(ctx, identityID)
}

// ListIdentities returns all identities linked to a user.
func (s *Service) ListIdentities(ctx context.Context, userID uuid.UUID) ([]models.Identity, error) {
	return s.identities.ListByUser(ctx, userID)
}

// ExportUserData returns all user data for GDPR export.
func (s *Service) ExportUserData(ctx context.Context, tenantID, userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	identities, err := s.identities.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	export := map[string]interface{}{
		"user":       user,
		"identities": identities,
		"exported_at": time.Now().UTC(),
	}
	return export, nil
}

// --- Helpers ---

func (s *Service) getTenantSettings(_ context.Context, _ uuid.UUID) models.TenantSettings {
	// In production, this would load from the tenant's settings JSONB
	return models.DefaultTenantSettings()
}

func validatePasswordPolicy(password string, settings models.TenantSettings) *models.AppError {
	if len(password) < settings.PasswordMinLength {
		return models.ErrPasswordPolicy.WithMessage(
			fmt.Sprintf("Password must be at least %d characters long.", settings.PasswordMinLength))
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSymbol = true
		}
	}

	if settings.PasswordRequireUpper && !hasUpper {
		return models.ErrPasswordPolicy.WithMessage("Password must contain at least one uppercase letter.")
	}
	if settings.PasswordRequireLower && !hasLower {
		return models.ErrPasswordPolicy.WithMessage("Password must contain at least one lowercase letter.")
	}
	if settings.PasswordRequireDigit && !hasDigit {
		return models.ErrPasswordPolicy.WithMessage("Password must contain at least one digit.")
	}
	if settings.PasswordRequireSymbol && !hasSymbol {
		return models.ErrPasswordPolicy.WithMessage("Password must contain at least one symbol.")
	}

	return nil
}

// checkHIBP checks the password against the Have I Been Pwned API using k-anonymity.
func checkHIBP(password string) (bool, error) {
	h := sha1.New()
	h.Write([]byte(password))
	hash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	prefix := hash[:5]
	suffix := hash[5:]

	resp, err := http.Get("https://api.pwnedpasswords.com/range/" + prefix)
	if err != nil {
		return false, fmt.Errorf("HIBP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("reading HIBP response: %w", err)
	}

	lines := strings.Split(string(body), "\r\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], suffix) {
			return true, nil
		}
	}

	return false, nil
}
