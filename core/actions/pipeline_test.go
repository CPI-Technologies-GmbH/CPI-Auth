package actions

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// --- Mock Action Repository ---

type mockActionRepo struct {
	actions map[uuid.UUID]*models.Action
}

func newMockActionRepo() *mockActionRepo {
	return &mockActionRepo{
		actions: make(map[uuid.UUID]*models.Action),
	}
}

func (m *mockActionRepo) Create(_ context.Context, action *models.Action) error {
	if action.ID == uuid.Nil {
		action.ID = uuid.New()
	}
	m.actions[action.ID] = action
	return nil
}

func (m *mockActionRepo) GetByID(_ context.Context, _ uuid.UUID, id uuid.UUID) (*models.Action, error) {
	a, ok := m.actions[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return a, nil
}

func (m *mockActionRepo) Update(_ context.Context, action *models.Action) error {
	m.actions[action.ID] = action
	return nil
}

func (m *mockActionRepo) Delete(_ context.Context, _ uuid.UUID, id uuid.UUID) error {
	delete(m.actions, id)
	return nil
}

func (m *mockActionRepo) List(_ context.Context, _ uuid.UUID, _ models.PaginationParams) (*models.PaginatedResult[models.Action], error) {
	var acts []models.Action
	for _, a := range m.actions {
		acts = append(acts, *a)
	}
	return &models.PaginatedResult[models.Action]{Data: acts, Total: int64(len(acts))}, nil
}

func (m *mockActionRepo) ListByTrigger(_ context.Context, tenantID uuid.UUID, trigger string) ([]models.Action, error) {
	var result []models.Action
	for _, a := range m.actions {
		if a.TenantID == tenantID && a.Trigger == trigger && a.Enabled {
			result = append(result, *a)
		}
	}
	// Sort by order
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Order > result[j].Order {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result, nil
}

// --- Tests ---

func TestNewPipeline(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	if pipeline == nil {
		t.Fatal("NewPipeline returned nil")
	}
}

func TestPipeline_Execute_NoActions(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	actCtx := &ActionContext{
		TenantID: tenantID,
		Data:     map[string]interface{}{"key": "value"},
	}

	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, actCtx)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Allow {
		t.Error("Execute with no actions should allow")
	}
	if result.Data["key"] != "value" {
		t.Error("data should be preserved")
	}
}

func TestPipeline_Execute_NilContext(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Allow {
		t.Error("Execute with nil context should allow")
	}
	if result.Data == nil {
		t.Error("data should be initialized even with nil context")
	}
}

func TestPipeline_Execute_AllowingHandler(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID:       uuid.New(),
		TenantID: tenantID,
		Trigger:  TriggerPreLogin,
		Name:     "check-ip",
		Enabled:  true,
		Order:    1,
	}

	pipeline.RegisterHandler("check-ip", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		actCtx.Data["checked"] = true
		return &ActionResult{Allow: true, Data: map[string]interface{}{"ip_ok": true}}, nil
	})

	actCtx := &ActionContext{Data: make(map[string]interface{})}
	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, actCtx)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Allow {
		t.Error("pipeline should allow when handler allows")
	}
	if result.Data["ip_ok"] != true {
		t.Error("handler data should be merged into result")
	}
}

func TestPipeline_Execute_DenyingHandler(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID:       uuid.New(),
		TenantID: tenantID,
		Trigger:  TriggerPreRegistration,
		Name:     "block-registration",
		Enabled:  true,
		Order:    1,
	}

	pipeline.RegisterHandler("block-registration", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		return &ActionResult{Allow: false, Message: "Registration blocked"}, nil
	})

	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreRegistration, &ActionContext{Data: make(map[string]interface{})})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Allow {
		t.Error("pipeline should deny when handler denies")
	}
	if result.Message != "Registration blocked" {
		t.Errorf("Message = %q, want %q", result.Message, "Registration blocked")
	}
}

func TestPipeline_Execute_DenyWithDefaultMessage(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID:       uuid.New(),
		TenantID: tenantID,
		Trigger:  TriggerPreLogin,
		Name:     "deny-action",
		Enabled:  true,
		Order:    1,
	}

	pipeline.RegisterHandler("deny-action", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		return &ActionResult{Allow: false, Message: ""}, nil
	})

	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, &ActionContext{Data: make(map[string]interface{})})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Allow {
		t.Error("pipeline should deny")
	}
	if result.Message == "" {
		t.Error("should have a default deny message")
	}
}

func TestPipeline_Execute_Order(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	// Action with order 2
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPostLogin, Name: "second", Enabled: true, Order: 2,
	}
	// Action with order 1
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPostLogin, Name: "first", Enabled: true, Order: 1,
	}

	var executionOrder []string
	pipeline.RegisterHandler("first", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		executionOrder = append(executionOrder, "first")
		return &ActionResult{Allow: true}, nil
	})
	pipeline.RegisterHandler("second", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		executionOrder = append(executionOrder, "second")
		return &ActionResult{Allow: true}, nil
	})

	_, err := pipeline.Execute(context.Background(), tenantID, TriggerPostLogin, &ActionContext{Data: make(map[string]interface{})})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(executionOrder) != 2 {
		t.Fatalf("expected 2 handlers executed, got %d", len(executionOrder))
	}
	if executionOrder[0] != "first" || executionOrder[1] != "second" {
		t.Errorf("execution order = %v, want [first, second]", executionOrder)
	}
}

func TestPipeline_Execute_DenyStopsPipeline(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPreLogin, Name: "denier", Enabled: true, Order: 1,
	}
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPreLogin, Name: "after-deny", Enabled: true, Order: 2,
	}

	pipeline.RegisterHandler("denier", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		return &ActionResult{Allow: false, Message: "denied"}, nil
	})

	afterDenyCalled := false
	pipeline.RegisterHandler("after-deny", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		afterDenyCalled = true
		return &ActionResult{Allow: true}, nil
	})

	result, _ := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, &ActionContext{Data: make(map[string]interface{})})

	if result.Allow {
		t.Error("pipeline should be denied")
	}
	if afterDenyCalled {
		t.Error("handlers after deny should not be called")
	}
}

func TestPipeline_Execute_HandlerError_Continues(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPreLogin, Name: "failing", Enabled: true, Order: 1,
	}
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPreLogin, Name: "succeeding", Enabled: true, Order: 2,
	}

	pipeline.RegisterHandler("failing", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		return nil, fmt.Errorf("something went wrong")
	})

	succeedingCalled := false
	pipeline.RegisterHandler("succeeding", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		succeedingCalled = true
		return &ActionResult{Allow: true}, nil
	})

	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, &ActionContext{Data: make(map[string]interface{})})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Allow {
		t.Error("pipeline should allow when failing handler is skipped")
	}
	if !succeedingCalled {
		t.Error("next handler should still be called after a handler error")
	}
}

func TestPipeline_Execute_UnregisteredHandler_Skipped(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPreLogin, Name: "no-handler-registered", Enabled: true, Order: 1,
	}

	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, &ActionContext{Data: make(map[string]interface{})})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Allow {
		t.Error("pipeline should allow when handler is not registered (skipped)")
	}
}

func TestPipeline_Execute_DisabledAction_Ignored(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPreLogin, Name: "disabled-action", Enabled: false, Order: 1,
	}

	handlerCalled := false
	pipeline.RegisterHandler("disabled-action", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		handlerCalled = true
		return &ActionResult{Allow: false, Message: "should not be called"}, nil
	})

	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPreLogin, &ActionContext{Data: make(map[string]interface{})})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Allow {
		t.Error("disabled actions should not block the pipeline")
	}
	if handlerCalled {
		t.Error("handler for disabled action should not be called")
	}
}

func TestPipeline_Execute_DataMerging(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())

	tenantID := uuid.New()
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPostLogin, Name: "enrich1", Enabled: true, Order: 1,
	}
	repo.actions[uuid.New()] = &models.Action{
		ID: uuid.New(), TenantID: tenantID, Trigger: TriggerPostLogin, Name: "enrich2", Enabled: true, Order: 2,
	}

	pipeline.RegisterHandler("enrich1", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		return &ActionResult{Allow: true, Data: map[string]interface{}{"from_1": "data1"}}, nil
	})
	pipeline.RegisterHandler("enrich2", func(ctx context.Context, actCtx *ActionContext) (*ActionResult, error) {
		// Should see data from enrich1
		if actCtx.Data["from_1"] != "data1" {
			return &ActionResult{Allow: false, Message: "did not see data from first handler"}, nil
		}
		return &ActionResult{Allow: true, Data: map[string]interface{}{"from_2": "data2"}}, nil
	})

	result, err := pipeline.Execute(context.Background(), tenantID, TriggerPostLogin, &ActionContext{Data: make(map[string]interface{})})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !result.Allow {
		t.Errorf("pipeline denied: %s", result.Message)
	}
	if result.Data["from_1"] != "data1" {
		t.Error("data from handler 1 should be in final result")
	}
	if result.Data["from_2"] != "data2" {
		t.Error("data from handler 2 should be in final result")
	}
}

func TestPipeline_CRUD(t *testing.T) {
	repo := newMockActionRepo()
	pipeline := NewPipeline(repo, zap.NewNop())
	ctx := context.Background()
	tenantID := uuid.New()

	// Create
	action := &models.Action{
		TenantID: tenantID,
		Trigger:  TriggerPreLogin,
		Name:     "test-action",
		Enabled:  true,
		Order:    1,
	}
	if err := pipeline.Create(ctx, action); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// GetByID
	got, err := pipeline.GetByID(ctx, tenantID, action.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got.Name != "test-action" {
		t.Errorf("Name = %q, want %q", got.Name, "test-action")
	}

	// Update
	action.Name = "updated-action"
	if err := pipeline.Update(ctx, action); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	// List
	result, err := pipeline.List(ctx, tenantID, models.PaginationParams{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}

	// Delete
	if err := pipeline.Delete(ctx, tenantID, action.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	_, err = pipeline.GetByID(ctx, tenantID, action.ID)
	if !models.IsAppError(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
