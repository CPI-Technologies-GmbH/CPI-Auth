package events

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Event types.
const (
	EventUserCreated       = "user.created"
	EventUserUpdated       = "user.updated"
	EventUserDeleted       = "user.deleted"
	EventLoginSuccess      = "login.success"
	EventLoginFailed       = "login.failed"
	EventMFAEnrolled       = "mfa.enrolled"
	EventMFAVerified       = "mfa.verified"
	EventSessionCreated    = "session.created"
	EventSessionRevoked    = "session.revoked"
	EventOrgCreated        = "organization.created"
	EventOrgUpdated        = "organization.updated"
	EventOrgDeleted        = "organization.deleted"
	EventPasswordChanged   = "password.changed"
	EventPasswordReset     = "password.reset"
	EventEmailVerified     = "email.verified"
	EventTokenIssued       = "token.issued"
	EventTokenRevoked      = "token.revoked"
)

// Event represents a system event.
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	ActorID   string                 `json:"actor_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IP        string                 `json:"ip,omitempty"`
}

// Service manages event publishing, audit logging, and webhook delivery.
type Service struct {
	nc        *nats.Conn
	auditLogs models.AuditLogRepository
	webhooks  models.WebhookRepository
	logger    *zap.Logger
	httpClient *http.Client
}

// NewService creates a new event service.
func NewService(nc *nats.Conn, auditLogs models.AuditLogRepository, webhooks models.WebhookRepository, logger *zap.Logger) *Service {
	return &Service{
		nc:        nc,
		auditLogs: auditLogs,
		webhooks:  webhooks,
		logger:    logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Publish publishes an event to NATS, writes an audit log, and triggers webhooks.
func (s *Service) Publish(ctx context.Context, event Event) {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Publish to NATS
	if s.nc != nil {
		data, err := json.Marshal(event)
		if err == nil {
			subject := "cpi-auth.events." + event.Type
			if pubErr := s.nc.Publish(subject, data); pubErr != nil {
				s.logger.Error("failed to publish event to NATS",
					zap.String("event_type", event.Type),
					zap.Error(pubErr),
				)
			}
		}
	}

	// Write audit log (PII masking applied)
	s.writeAuditLog(ctx, event)

	// Trigger webhooks asynchronously
	go s.triggerWebhooks(ctx, event)
}

// writeAuditLog persists an immutable audit log entry with PII masking.
func (s *Service) writeAuditLog(ctx context.Context, event Event) {
	if s.auditLogs == nil {
		return
	}

	tenantID, _ := uuid.Parse(event.TenantID)
	var actorID *uuid.UUID
	if event.ActorID != "" {
		parsed, err := uuid.Parse(event.ActorID)
		if err == nil {
			actorID = &parsed
		}
	}

	// Apply PII masking to metadata
	maskedData := maskPII(event.Data)
	metadataJSON, _ := json.Marshal(maskedData)

	log := &models.AuditLog{
		TenantID:   tenantID,
		ActorID:    actorID,
		Action:     event.Type,
		TargetType: getTargetType(event.Type),
		TargetID:   getTargetID(event.Data),
		Metadata:   metadataJSON,
		IP:         event.IP,
	}

	if err := s.auditLogs.Create(ctx, log); err != nil {
		s.logger.Error("failed to write audit log",
			zap.String("event_type", event.Type),
			zap.Error(err),
		)
	}
}

// triggerWebhooks sends the event to all matching webhook endpoints.
func (s *Service) triggerWebhooks(ctx context.Context, event Event) {
	if s.webhooks == nil {
		return
	}

	tenantID, err := uuid.Parse(event.TenantID)
	if err != nil {
		return
	}

	hooks, err := s.webhooks.ListByEvent(ctx, tenantID, event.Type)
	if err != nil {
		s.logger.Error("failed to list webhooks", zap.Error(err))
		return
	}

	for _, hook := range hooks {
		s.deliverWebhook(ctx, hook, event)
	}
}

// deliverWebhook sends a single webhook with HMAC signature and retry logic.
func (s *Service) deliverWebhook(ctx context.Context, hook models.Webhook, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	// Compute HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(hook.Secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	// Retry with exponential backoff (3 attempts)
	delays := []time.Duration{0, 2 * time.Second, 10 * time.Second}
	for attempt, delay := range delays {
		if delay > 0 {
			time.Sleep(delay)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(payload))
		if err != nil {
			s.logger.Error("failed to create webhook request", zap.Error(err))
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CPI Auth-Signature", signature)
		req.Header.Set("X-CPI Auth-Event", event.Type)
		req.Header.Set("X-CPI Auth-Delivery", event.ID)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			s.logger.Warn("webhook delivery failed",
				zap.String("url", hook.URL),
				zap.Int("attempt", attempt+1),
				zap.Error(err),
			)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return // Success
		}

		s.logger.Warn("webhook returned non-2xx",
			zap.String("url", hook.URL),
			zap.Int("status", resp.StatusCode),
			zap.Int("attempt", attempt+1),
		)
	}

	s.logger.Error("webhook delivery exhausted retries",
		zap.String("url", hook.URL),
		zap.String("event_type", event.Type),
	)
}

// ListAuditLogs returns paginated audit logs.
func (s *Service) ListAuditLogs(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams, action string) (*models.PaginatedResult[models.AuditLog], error) {
	return s.auditLogs.List(ctx, tenantID, params, action)
}

// --- PII Masking ---

var piiFields = map[string]bool{
	"email": true, "phone": true, "password": true,
	"ip": true, "user_agent": true, "name": true,
}

func maskPII(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	masked := make(map[string]interface{}, len(data))
	for k, v := range data {
		if piiFields[strings.ToLower(k)] {
			if str, ok := v.(string); ok && len(str) > 4 {
				masked[k] = str[:2] + "***" + str[len(str)-2:]
			} else {
				masked[k] = "***"
			}
		} else {
			masked[k] = v
		}
	}
	return masked
}

func getTargetType(eventType string) string {
	parts := strings.SplitN(eventType, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func getTargetID(data map[string]interface{}) string {
	if id, ok := data["user_id"]; ok {
		return fmt.Sprintf("%v", id)
	}
	if id, ok := data["target_id"]; ok {
		return fmt.Sprintf("%v", id)
	}
	return ""
}
