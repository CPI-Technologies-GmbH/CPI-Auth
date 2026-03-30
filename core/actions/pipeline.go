package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Trigger constants define hook points in the authentication lifecycle.
const (
	TriggerPreRegistration     = "pre-registration"
	TriggerPostRegistration    = "post-registration"
	TriggerPreLogin            = "pre-login"
	TriggerPostLogin           = "post-login"
	TriggerPreToken            = "pre-token"
	TriggerPostChangePassword  = "post-change-password"
	TriggerPreUserUpdate       = "pre-user-update"
	TriggerPostUserDelete      = "post-user-delete"
)

// ActionContext holds the context passed to action handlers.
type ActionContext struct {
	TenantID uuid.UUID              `json:"tenant_id"`
	UserID   *uuid.UUID             `json:"user_id,omitempty"`
	IP       string                 `json:"ip,omitempty"`
	Data     map[string]interface{} `json:"data"`
}

// ActionResult holds the result from an action handler.
type ActionResult struct {
	Allow   bool                   `json:"allow"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// Handler is a function type for Go-native action handlers.
type Handler func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error)

// Pipeline manages the execution of actions at trigger points.
type Pipeline struct {
	actionRepo models.ActionRepository
	handlers   map[string]Handler
	logger     *zap.Logger
}

// NewPipeline creates a new action pipeline.
func NewPipeline(actionRepo models.ActionRepository, logger *zap.Logger) *Pipeline {
	return &Pipeline{
		actionRepo: actionRepo,
		handlers:   make(map[string]Handler),
		logger:     logger,
	}
}

// RegisterHandler registers a Go-native action handler for a named action.
func (p *Pipeline) RegisterHandler(name string, handler Handler) {
	p.handlers[name] = handler
}

// Execute runs all enabled actions for a trigger point in order.
func (p *Pipeline) Execute(ctx context.Context, tenantID uuid.UUID, trigger string, actCtx *ActionContext) (*ActionResult, error) {
	if actCtx == nil {
		actCtx = &ActionContext{
			TenantID: tenantID,
			Data:     make(map[string]interface{}),
		}
	}
	actCtx.TenantID = tenantID

	// Get all enabled actions for this trigger
	actions, err := p.actionRepo.ListByTrigger(ctx, tenantID, trigger)
	if err != nil {
		p.logger.Error("failed to list actions for trigger",
			zap.String("trigger", trigger),
			zap.Error(err),
		)
		// Don't block the flow if actions can't be loaded
		return &ActionResult{Allow: true, Data: actCtx.Data}, nil
	}

	if len(actions) == 0 {
		return &ActionResult{Allow: true, Data: actCtx.Data}, nil
	}

	// Execute actions in order
	for _, action := range actions {
		handler, ok := p.handlers[action.Name]
		if !ok {
			p.logger.Warn("no handler registered for action",
				zap.String("action", action.Name),
				zap.String("trigger", trigger),
			)
			continue
		}

		result, err := handler(ctx, actCtx)
		if err != nil {
			p.logger.Error("action handler failed",
				zap.String("action", action.Name),
				zap.String("trigger", trigger),
				zap.Error(err),
			)
			// Continue to next action unless it explicitly denies
			continue
		}

		if result != nil {
			// Merge result data into context for next action
			if result.Data != nil {
				for k, v := range result.Data {
					actCtx.Data[k] = v
				}
			}

			// If action denies, stop the pipeline
			if !result.Allow {
				msg := result.Message
				if msg == "" {
					msg = fmt.Sprintf("Action '%s' denied the request.", action.Name)
				}
				return &ActionResult{
					Allow:   false,
					Data:    actCtx.Data,
					Message: msg,
				}, nil
			}
		}
	}

	return &ActionResult{Allow: true, Data: actCtx.Data}, nil
}

// CRUD operations for actions

// Create creates a new action.
func (p *Pipeline) Create(ctx context.Context, action *models.Action) error {
	return p.actionRepo.Create(ctx, action)
}

// GetByID retrieves an action by ID.
func (p *Pipeline) GetByID(ctx context.Context, tenantID, actionID uuid.UUID) (*models.Action, error) {
	return p.actionRepo.GetByID(ctx, tenantID, actionID)
}

// Update updates an action.
func (p *Pipeline) Update(ctx context.Context, action *models.Action) error {
	return p.actionRepo.Update(ctx, action)
}

// Delete removes an action.
func (p *Pipeline) Delete(ctx context.Context, tenantID, actionID uuid.UUID) error {
	return p.actionRepo.Delete(ctx, tenantID, actionID)
}

// List returns all actions for a tenant.
func (p *Pipeline) List(ctx context.Context, tenantID uuid.UUID, params models.PaginationParams) (*models.PaginatedResult[models.Action], error) {
	return p.actionRepo.List(ctx, tenantID, params)
}
