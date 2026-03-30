package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Service manages user sessions backed by Redis for speed and PostgreSQL for durability.
type Service struct {
	sessions models.SessionRepository
	redis    *redis.Client
	cfg      *config.Config
	logger   *zap.Logger
}

// NewService creates a new session service.
func NewService(sessions models.SessionRepository, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		sessions: sessions,
		redis:    rdb,
		cfg:      cfg,
		logger:   logger,
	}
}

// CreateSessionInput holds data needed to create a session.
type CreateSessionInput struct {
	UserID    uuid.UUID
	TenantID  uuid.UUID
	IP        string
	UserAgent string
	DeviceInfo map[string]interface{}
}

func sessionRedisKey(id uuid.UUID) string {
	return fmt.Sprintf("session:%s", id.String())
}

func userSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user_sessions:%s", userID.String())
}

// Create creates a new session and stores it in both Redis and the database.
func (s *Service) Create(ctx context.Context, input CreateSessionInput) (*models.Session, error) {
	deviceInfoJSON, _ := json.Marshal(input.DeviceInfo)

	session := &models.Session{
		UserID:   input.UserID,
		TenantID: input.TenantID,
		IP:       input.IP,
		UserAgent: input.UserAgent,
		DeviceInfo: deviceInfoJSON,
		ExpiresAt: time.Now().UTC().Add(s.cfg.Security.SessionLifetime),
	}

	// Store in database
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("creating session in db: %w", err)
	}

	// Cache in Redis (if available)
	if s.redis != nil {
		sessionJSON, _ := json.Marshal(session)
		ttl := time.Until(session.ExpiresAt)
		if err := s.redis.Set(ctx, sessionRedisKey(session.ID), sessionJSON, ttl).Err(); err != nil {
			s.logger.Warn("failed to cache session in redis", zap.Error(err))
		}

		// Track session in user's session set
		s.redis.SAdd(ctx, userSessionsKey(input.UserID), session.ID.String())
		s.redis.Expire(ctx, userSessionsKey(input.UserID), s.cfg.Security.SessionLifetime)
	}

	s.logger.Info("session created",
		zap.String("session_id", session.ID.String()),
		zap.String("user_id", input.UserID.String()),
	)

	return session, nil
}

// Get retrieves a session by ID, checking Redis first, falling back to DB.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	// Try Redis first
	if s.redis != nil {
		data, err := s.redis.Get(ctx, sessionRedisKey(id)).Bytes()
		if err == nil {
			var session models.Session
			if err := json.Unmarshal(data, &session); err == nil {
				// Check expiration
				if time.Now().UTC().After(session.ExpiresAt) {
					s.Revoke(ctx, id)
					return nil, models.ErrSessionExpired
				}
				return &session, nil
			}
		}
	}

	// Fallback to database
	session, err := s.sessions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		s.Revoke(ctx, id)
		return nil, models.ErrSessionExpired
	}

	return session, nil
}

// Touch updates the last active time and extends session via sliding window.
func (s *Service) Touch(ctx context.Context, id uuid.UUID) error {
	session, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	session.LastActiveAt = now

	// Sliding window: extend expiry if within inactivity timeout
	inactivityTimeout := s.cfg.Security.InactivityTimeout
	if now.Sub(session.LastActiveAt) < inactivityTimeout {
		session.ExpiresAt = now.Add(s.cfg.Security.SessionLifetime)
	}

	// Update in database
	if err := s.sessions.Update(ctx, session); err != nil {
		return err
	}

	// Update in Redis
	if s.redis != nil {
		sessionJSON, _ := json.Marshal(session)
		ttl := time.Until(session.ExpiresAt)
		s.redis.Set(ctx, sessionRedisKey(id), sessionJSON, ttl)
	}

	return nil
}

// Revoke destroys a session.
func (s *Service) Revoke(ctx context.Context, id uuid.UUID) error {
	session, err := s.sessions.GetByID(ctx, id)
	if err != nil && !models.IsAppError(err, models.ErrNotFound) {
		return err
	}

	// Remove from Redis
	if s.redis != nil {
		s.redis.Del(ctx, sessionRedisKey(id))
		if session != nil {
			s.redis.SRem(ctx, userSessionsKey(session.UserID), id.String())
		}
	}

	// Remove from database
	return s.sessions.Delete(ctx, id)
}

// ListByUser returns all active sessions for a user.
func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	return s.sessions.ListByUser(ctx, userID)
}

// RevokeAllForUser destroys all sessions for a user (force logout).
func (s *Service) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.sessions.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	if s.redis != nil {
		for _, sess := range sessions {
			s.redis.Del(ctx, sessionRedisKey(sess.ID))
		}
		s.redis.Del(ctx, userSessionsKey(userID))
	}

	return s.sessions.DeleteByUser(ctx, userID)
}

// RevokeAllForTenant destroys all sessions for an entire tenant.
func (s *Service) RevokeAllForTenant(ctx context.Context, tenantID uuid.UUID) error {
	return s.sessions.DeleteByTenant(ctx, tenantID)
}

// Validate checks if a session is still valid (not expired, not inactive).
func (s *Service) Validate(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	session, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	// Check absolute expiration
	if now.After(session.ExpiresAt) {
		s.Revoke(ctx, id)
		return nil, models.ErrSessionExpired
	}

	// Check inactivity timeout
	inactivityTimeout := s.cfg.Security.InactivityTimeout
	if inactivityTimeout > 0 && now.Sub(session.LastActiveAt) > inactivityTimeout {
		s.Revoke(ctx, id)
		return nil, models.ErrSessionExpired.WithMessage("Session expired due to inactivity.")
	}

	return session, nil
}
