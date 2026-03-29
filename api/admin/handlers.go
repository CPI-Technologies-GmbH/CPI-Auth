package admin

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/actions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/domains"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/events"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/license"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/policy"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/sessions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/users"
)

// Handler holds dependencies for admin API handlers.
type Handler struct {
	userSvc      *users.Service
	sessionSvc   *sessions.Service
	eventSvc     *events.Service
	actionSvc    *actions.Pipeline
	tokenSvc     *tokens.Service
	rbacSvc      *policy.RBACService
	tenantRepo   models.TenantRepository
	appRepo      models.ApplicationRepository
	orgRepo      models.OrganizationRepository
	roleRepo     models.RoleRepository
	webhookRepo  models.WebhookRepository
	templateRepo models.EmailTemplateRepository
	apiKeyRepo   models.APIKeyRepository
	permRepo        models.PermissionRepository
	appPermRepo     models.ApplicationPermissionRepository
	customFieldRepo  models.CustomFieldDefinitionRepository
	domainSvc        *domains.Service
	pageTemplateRepo models.PageTemplateRepository
	langStringRepo   models.LanguageStringRepository
	licenseChecker   *license.Checker
	logger           *zap.Logger
}

// NewHandler creates new admin handlers.
func NewHandler(
	userSvc *users.Service,
	sessionSvc *sessions.Service,
	eventSvc *events.Service,
	actionSvc *actions.Pipeline,
	tokenSvc *tokens.Service,
	rbacSvc *policy.RBACService,
	tenantRepo models.TenantRepository,
	appRepo models.ApplicationRepository,
	orgRepo models.OrganizationRepository,
	roleRepo models.RoleRepository,
	webhookRepo models.WebhookRepository,
	templateRepo models.EmailTemplateRepository,
	apiKeyRepo models.APIKeyRepository,
	permRepo models.PermissionRepository,
	appPermRepo models.ApplicationPermissionRepository,
	customFieldRepo models.CustomFieldDefinitionRepository,
	domainSvc *domains.Service,
	pageTemplateRepo models.PageTemplateRepository,
	langStringRepo models.LanguageStringRepository,
	licenseChecker *license.Checker,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		userSvc:      userSvc,
		sessionSvc:   sessionSvc,
		eventSvc:     eventSvc,
		actionSvc:    actionSvc,
		tokenSvc:     tokenSvc,
		rbacSvc:      rbacSvc,
		tenantRepo:   tenantRepo,
		appRepo:      appRepo,
		orgRepo:      orgRepo,
		roleRepo:     roleRepo,
		webhookRepo:  webhookRepo,
		templateRepo: templateRepo,
		apiKeyRepo:   apiKeyRepo,
		permRepo:        permRepo,
		appPermRepo:     appPermRepo,
		customFieldRepo:  customFieldRepo,
		domainSvc:        domainSvc,
		pageTemplateRepo: pageTemplateRepo,
		langStringRepo:   langStringRepo,
		licenseChecker:   licenseChecker,
		logger:           logger,
	}
}

// RegisterRoutes registers all admin CRUD routes (relative paths, to be mounted under a subrouter).
func (h *Handler) RegisterRoutes(r chi.Router) {
	h.registerRoutesWithPrefix(r, "")
}

func (h *Handler) registerRoutesWithPrefix(r chi.Router, p string) {
	// Users
	r.Get(p+"/users", h.ListUsers)
	r.Post(p+"/users", h.CreateUser)
	r.Get(p+"/users/export", h.ExportUsers)
	r.Post(p+"/users/import", h.ImportUsers)
	r.Post(p+"/users/bulk/block", h.BulkBlockUsers)
	r.Post(p+"/users/bulk/delete", h.BulkDeleteUsers)
	r.Get(p+"/users/{id}", h.GetUser)
	r.Patch(p+"/users/{id}", h.UpdateUser)
	r.Delete(p+"/users/{id}", h.DeleteUser)
	r.Post(p+"/users/{id}/block", h.BlockUser)
	r.Post(p+"/users/{id}/unblock", h.UnblockUser)
	r.Get(p+"/users/{id}/sessions", h.GetUserSessions)
	r.Delete(p+"/users/{id}/sessions", h.ForceLogoutUser)
	r.Delete(p+"/users/{id}/sessions/{sessionId}", h.RevokeUserSession)
	r.Post(p+"/users/{id}/force-logout", h.ForceLogoutUserPost)
	r.Post(p+"/users/{id}/reset-password", h.ResetUserPassword)
	r.Get(p+"/users/{id}/mfa", h.GetUserMFA)
	r.Get(p+"/users/{id}/identities", h.GetUserIdentities)
	r.Get(p+"/users/{id}/audit-log", h.GetUserAuditLog)
	r.Get(p+"/users/{id}/roles", h.GetUserRoles)
	r.Post(p+"/users/{id}/roles", h.AssignUserRole)
	r.Delete(p+"/users/{id}/roles/{roleId}", h.RemoveUserRole)
	r.Post(p+"/users/{id}/impersonate", h.ImpersonateUser)

	// Tenants
	r.Get(p+"/tenants", h.ListTenants)
	r.Post(p+"/tenants", h.CreateTenant)
	r.Get(p+"/tenants/{id}", h.GetTenant)
	r.Patch(p+"/tenants/{id}", h.UpdateTenant)
	r.Delete(p+"/tenants/{id}", h.DeleteTenant)
	r.Post(p+"/tenants/{id}/force-logout", h.ForceLogoutTenant)

	// Applications
	r.Get(p+"/applications", h.ListApplications)
	r.Post(p+"/applications", h.CreateApplication)
	r.Get(p+"/applications/{id}", h.GetApplication)
	r.Patch(p+"/applications/{id}", h.UpdateApplication)
	r.Delete(p+"/applications/{id}", h.DeleteApplication)
	r.Post(p+"/applications/{id}/rotate-secret", h.RotateClientSecret)
	r.Get(p+"/applications/{id}/permissions", h.GetApplicationPermissions)
	r.Put(p+"/applications/{id}/permissions", h.SetApplicationPermissions)

	// Organizations
	r.Get(p+"/organizations", h.ListOrganizations)
	r.Post(p+"/organizations", h.CreateOrganization)
	r.Get(p+"/organizations/{id}", h.GetOrganization)
	r.Patch(p+"/organizations/{id}", h.UpdateOrganization)
	r.Delete(p+"/organizations/{id}", h.DeleteOrganization)
	r.Get(p+"/organizations/{id}/members", h.ListOrgMembers)
	r.Post(p+"/organizations/{id}/members", h.AddOrgMember)
	r.Delete(p+"/organizations/{id}/members/{userId}", h.RemoveOrgMember)

	// Roles & Permissions
	r.Get(p+"/roles", h.ListRoles)
	r.Post(p+"/roles", h.CreateRole)
	r.Get(p+"/roles/{id}", h.GetRole)
	r.Patch(p+"/roles/{id}", h.UpdateRole)
	r.Delete(p+"/roles/{id}", h.DeleteRole)
	r.Get(p+"/permissions", h.ListPermissions)
	r.Post(p+"/permissions", h.CreatePermission)
	r.Get(p+"/permissions/{id}", h.GetPermission)
	r.Patch(p+"/permissions/{id}", h.UpdatePermission)
	r.Delete(p+"/permissions/{id}", h.DeletePermission)

	// Webhooks
	r.Get(p+"/webhooks", h.ListWebhooks)
	r.Post(p+"/webhooks", h.CreateWebhook)
	r.Get(p+"/webhooks/{id}", h.GetWebhook)
	r.Patch(p+"/webhooks/{id}", h.UpdateWebhook)
	r.Delete(p+"/webhooks/{id}", h.DeleteWebhook)
	r.Post(p+"/webhooks/{id}/test", h.TestWebhook)
	r.Get(p+"/webhooks/{id}/deliveries", h.GetWebhookDeliveries)

	// Actions
	r.Get(p+"/actions", h.ListActions)
	r.Post(p+"/actions", h.CreateAction)
	r.Post(p+"/actions/reorder", h.ReorderActions)
	r.Get(p+"/actions/{id}", h.GetAction)
	r.Patch(p+"/actions/{id}", h.UpdateAction)
	r.Delete(p+"/actions/{id}", h.DeleteAction)

	// Email Templates
	r.Get(p+"/email-templates", h.ListEmailTemplates)
	r.Post(p+"/email-templates", h.CreateEmailTemplate)
	r.Get(p+"/email-templates/{id}", h.GetEmailTemplate)
	r.Patch(p+"/email-templates/{id}", h.UpdateEmailTemplate)
	r.Delete(p+"/email-templates/{id}", h.DeleteEmailTemplate)
	r.Post(p+"/email-templates/{id}/test", h.SendTestEmail)

	// API Keys
	r.Get(p+"/api-keys", h.ListAPIKeys)
	r.Post(p+"/api-keys", h.CreateAPIKey)
	r.Get(p+"/api-keys/{id}", h.GetAPIKey)
	r.Patch(p+"/api-keys/{id}", h.UpdateAPIKey)
	r.Delete(p+"/api-keys/{id}", h.DeleteAPIKey)

	// Custom Fields
	r.Get(p+"/custom-fields", h.ListCustomFields)
	r.Post(p+"/custom-fields", h.CreateCustomField)
	r.Get(p+"/custom-fields/{id}", h.GetCustomField)
	r.Patch(p+"/custom-fields/{id}", h.UpdateCustomField)
	r.Delete(p+"/custom-fields/{id}", h.DeleteCustomField)

	// Page Templates
	r.Get(p+"/page-templates", h.ListPageTemplates)
	r.Post(p+"/page-templates", h.CreatePageTemplate)
	r.Get(p+"/page-templates/{id}", h.GetPageTemplate)
	r.Patch(p+"/page-templates/{id}", h.UpdatePageTemplate)
	r.Delete(p+"/page-templates/{id}", h.DeletePageTemplate)
	r.Post(p+"/page-templates/{id}/duplicate", h.DuplicatePageTemplate)

	// Language Strings
	r.Get(p+"/language-strings", h.ListLanguageStrings)
	r.Put(p+"/language-strings", h.UpsertLanguageString)
	r.Delete(p+"/language-strings/{key}/{locale}", h.DeleteLanguageString)

	// Domain Verification
	r.Get(p+"/domains/verification", h.GetDomainVerification)
	r.Post(p+"/domains/verification", h.InitiateDomainVerification)
	r.Post(p+"/domains/verification/{id}/check", h.CheckDomainVerification)
	r.Delete(p+"/domains/verification/{id}", h.DeleteDomainVerification)

	// Audit Logs
	r.Get(p+"/logs", h.ListAuditLogs)
	r.Get(p+"/audit-logs", h.ListAuditLogsV2)
	r.Get(p+"/audit-logs/export", h.ExportAuditLogs)

	// Dashboard
	r.Get(p+"/dashboard/metrics", h.GetDashboardMetrics)
	r.Get(p+"/dashboard/logins", h.GetLoginChart)
	r.Get(p+"/dashboard/auth-methods", h.GetAuthMethodsChart)
	r.Get(p+"/dashboard/events", h.GetRecentEvents)

	// Settings
	r.Get(p+"/settings", h.GetSettings)
	r.Patch(p+"/settings", h.UpdateSettings)
	r.Patch(p+"/settings/branding", h.UpdateSettingsBranding)
	r.Post(p+"/settings/smtp/test", h.TestSMTP)

	// Stats & Branding (legacy)
	r.Get(p+"/stats", h.GetStats)
	r.Get(p+"/branding", h.GetBranding)
	r.Patch(p+"/branding", h.UpdateBranding)
}

// --- Helpers ---

func getPagination(r *http.Request) models.PaginationParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return models.PaginationParams{Page: page, PerPage: perPage}
}

func parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// --- Users ---

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	params := getPagination(r)
	search := r.URL.Query().Get("search")

	result, err := h.userSvc.List(r.Context(), tenantID, params, search)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if h.licenseChecker != nil {
		if err := h.licenseChecker.CanCreateUser(r.Context()); err != nil {
			middleware.WriteError(w, models.ErrForbidden.WithMessage(err.Error()))
			return
		}
	}
	tenantID := middleware.GetTenantID(r.Context())

	var input users.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}

	user, err := h.userSvc.Register(r.Context(), tenantID, input)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventUserCreated,
		TenantID: tenantID.String(),
		ActorID:  middleware.GetUserID(r.Context()).String(),
		IP:       extractIP(r.RemoteAddr),
		Data:     map[string]interface{}{"user_id": user.ID.String(), "email": user.Email},
	})

	middleware.WriteJSON(w, http.StatusCreated, user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	var update struct {
		Name        *string          `json:"name"`
		Email       *string          `json:"email"`
		Phone       *string          `json:"phone"`
		AvatarURL   *string          `json:"avatar_url"`
		Locale      *string          `json:"locale"`
		Status      *models.Status   `json:"status"`
		Metadata    *json.RawMessage `json:"metadata"`
		AppMetadata *json.RawMessage `json:"app_metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}

	if update.Name != nil {
		user.Name = *update.Name
	}
	if update.Email != nil {
		user.Email = *update.Email
	}
	if update.Phone != nil {
		user.Phone = *update.Phone
	}
	if update.AvatarURL != nil {
		user.AvatarURL = *update.AvatarURL
	}
	if update.Locale != nil {
		user.Locale = *update.Locale
	}
	if update.Status != nil {
		user.Status = *update.Status
	}
	if update.Metadata != nil {
		user.Metadata = *update.Metadata
	}
	if update.AppMetadata != nil {
		user.AppMetadata = *update.AppMetadata
	}

	if err := h.userSvc.Update(r.Context(), user); err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	if err := h.userSvc.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventUserDeleted,
		TenantID: tenantID.String(),
		ActorID:  middleware.GetUserID(r.Context()).String(),
		IP:       extractIP(r.RemoteAddr),
		Data:     map[string]interface{}{"user_id": id.String()},
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	if err := h.userSvc.Block(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "blocked"})
}

func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	if err := h.userSvc.Unblock(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "active"})
}

func (h *Handler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	sessions, err := h.sessionSvc.ListByUser(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, sessions)
}

func (h *Handler) ForceLogoutUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	if err := h.sessionSvc.RevokeAllForUser(r.Context(), id); err != nil {
		middleware.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ImportUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var importReq struct {
		Users []users.RegisterInput `json:"users"`
	}
	if err := json.NewDecoder(r.Body).Decode(&importReq); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}

	var created int
	var errors []string
	for _, input := range importReq.Users {
		_, err := h.userSvc.Register(r.Context(), tenantID, input)
		if err != nil {
			errors = append(errors, input.Email+": "+err.Error())
			continue
		}
		created++
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"created": created,
		"errors":  errors,
	})
}

func (h *Handler) ExportUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	result, err := h.userSvc.List(r.Context(), tenantID, models.PaginationParams{Page: 1, PerPage: 100}, "")
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=users.json")
	json.NewEncoder(w).Encode(result.Data)
}

// --- Tenants ---

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	params := getPagination(r)
	result, err := h.tenantRepo.List(r.Context(), params)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if h.licenseChecker != nil {
		if err := h.licenseChecker.CanCreateTenant(r.Context()); err != nil {
			middleware.WriteError(w, models.ErrForbidden.WithMessage(err.Error()))
			return
		}
	}
	var tenant models.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	if err := h.tenantRepo.Create(r.Context(), &tenant); err != nil {
		middleware.WriteError(w, err)
		return
	}
	// Seed system permissions for the new tenant
	if err := h.permRepo.EnsureSystemDefaults(r.Context(), tenant.ID); err != nil {
		h.logger.Error("failed to seed permissions for new tenant", zap.String("tenant_id", tenant.ID.String()), zap.Error(err))
	}
	middleware.WriteJSON(w, http.StatusCreated, tenant)
}

func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	tenant, err := h.tenantRepo.GetByID(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tenant)
}

func (h *Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	tenant, err := h.tenantRepo.GetByID(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(tenant); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	tenant.ID = id
	if err := h.tenantRepo.Update(r.Context(), tenant); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tenant)
}

func (h *Handler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.tenantRepo.Delete(r.Context(), id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ForceLogoutTenant(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.sessionSvc.RevokeAllForTenant(r.Context(), id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Applications ---

func (h *Handler) ListApplications(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.appRepo.List(r.Context(), tenantID, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var app models.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	app.TenantID = tenantID
	if app.ClientID == "" {
		app.ClientID = uuid.New().String()
	}
	if app.RedirectURIs == nil {
		app.RedirectURIs = []string{}
	}
	if app.AllowedOrigins == nil {
		app.AllowedOrigins = []string{}
	}
	if app.AllowedLogoutURLs == nil {
		app.AllowedLogoutURLs = []string{}
	}
	if app.GrantTypes == nil {
		app.GrantTypes = []string{}
	}
	if app.Settings == nil {
		app.Settings = json.RawMessage(`{}`)
	}
	if app.AccessTokenTTL == nil {
		d := 3600
		app.AccessTokenTTL = &d
	}
	if app.RefreshTokenTTL == nil {
		d := 2592000
		app.RefreshTokenTTL = &d
	}
	if app.IDTokenTTL == nil {
		d := 3600
		app.IDTokenTTL = &d
	}
	if err := h.appRepo.Create(r.Context(), &app); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, app)
}

func (h *Handler) GetApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	app, err := h.appRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	app, err := h.appRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(app); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	app.ID = id
	app.TenantID = tenantID
	if err := h.appRepo.Update(r.Context(), app); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.appRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Organizations ---

func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.orgRepo.List(r.Context(), tenantID, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var org models.Organization
	if err := json.NewDecoder(r.Body).Decode(&org); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	org.TenantID = tenantID
	if org.Slug == "" {
		org.Slug = org.Name
	}
	if err := h.orgRepo.Create(r.Context(), &org); err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventOrgCreated,
		TenantID: tenantID.String(),
		ActorID:  middleware.GetUserID(r.Context()).String(),
		IP:       extractIP(r.RemoteAddr),
	})

	middleware.WriteJSON(w, http.StatusCreated, org)
}

func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	org, err := h.orgRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, org)
}

func (h *Handler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	org, err := h.orgRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(org); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	org.ID = id
	org.TenantID = tenantID
	if err := h.orgRepo.Update(r.Context(), org); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, org)
}

func (h *Handler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.orgRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListOrgMembers(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	result, err := h.orgRepo.ListMembers(r.Context(), id, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) AddOrgMember(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	var member models.OrganizationMember
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	member.OrgID = id
	if err := h.orgRepo.AddMember(r.Context(), &member); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, member)
}

func (h *Handler) RemoveOrgMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.orgRepo.RemoveMember(r.Context(), orgID, userID); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Roles ---

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.roleRepo.List(r.Context(), tenantID, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var role models.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	role.TenantID = tenantID
	if err := h.roleRepo.Create(r.Context(), &role); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, role)
}

func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	role, err := h.roleRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, role)
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	role, err := h.roleRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(role); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	role.ID = id
	role.TenantID = tenantID
	if err := h.roleRepo.Update(r.Context(), role); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, role)
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.roleRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Webhooks ---

func (h *Handler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.webhookRepo.List(r.Context(), tenantID, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var webhook models.Webhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	webhook.TenantID = tenantID
	if err := h.webhookRepo.Create(r.Context(), &webhook); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, webhook)
}

func (h *Handler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	webhook, err := h.webhookRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, webhook)
}

func (h *Handler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	webhook, err := h.webhookRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(webhook); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	webhook.ID = id
	webhook.TenantID = tenantID
	if err := h.webhookRepo.Update(r.Context(), webhook); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, webhook)
}

func (h *Handler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.webhookRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Actions ---

func (h *Handler) ListActions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.actionSvc.List(r.Context(), tenantID, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateAction(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var action models.Action
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	action.TenantID = tenantID
	if err := h.actionSvc.Create(r.Context(), &action); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, action)
}

func (h *Handler) GetAction(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	action, err := h.actionSvc.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, action)
}

func (h *Handler) UpdateAction(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	action, err := h.actionSvc.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(action); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	action.ID = id
	action.TenantID = tenantID
	if err := h.actionSvc.Update(r.Context(), action); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, action)
}

func (h *Handler) DeleteAction(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.actionSvc.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Email Templates ---

func (h *Handler) ListEmailTemplates(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.templateRepo.List(r.Context(), tenantID, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var tmpl models.EmailTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	tmpl.TenantID = tenantID
	if tmpl.Locale == "" {
		tmpl.Locale = "en"
	}
	if err := h.templateRepo.Create(r.Context(), &tmpl); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, tmpl)
}

func (h *Handler) GetEmailTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	tmpl, err := h.templateRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tmpl)
}

func (h *Handler) UpdateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	tmpl, err := h.templateRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(tmpl); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	tmpl.ID = id
	tmpl.TenantID = tenantID
	if err := h.templateRepo.Update(r.Context(), tmpl); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tmpl)
}

func (h *Handler) DeleteEmailTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.templateRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- API Keys ---

func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.apiKeyRepo.List(r.Context(), tenantID, getPagination(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var key models.APIKey
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	key.TenantID = tenantID
	if key.KeyHash == "" {
		key.KeyHash = uuid.New().String()
	}
	if key.KeyPrefix == "" {
		key.KeyPrefix = key.KeyHash[:8]
	}
	if key.Scopes == nil {
		key.Scopes = []string{}
	}
	if err := h.apiKeyRepo.Create(r.Context(), &key); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, key)
}

func (h *Handler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	key, err := h.apiKeyRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, key)
}

func (h *Handler) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	key, err := h.apiKeyRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(key); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	key.ID = id
	key.TenantID = tenantID
	if err := h.apiKeyRepo.Update(r.Context(), key); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, key)
}

func (h *Handler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.apiKeyRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Logs & Stats ---

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	params := getPagination(r)
	action := r.URL.Query().Get("action")

	result, err := h.eventSvc.ListAuditLogs(r.Context(), tenantID, params, action)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	userCount, _ := h.userSvc.List(r.Context(), tenantID, models.PaginationParams{Page: 1, PerPage: 1}, "")

	stats := map[string]interface{}{
		"total_users":        userCount.Total,
		"active_sessions":    0,
		"total_applications": 0,
		"timestamp":          "now",
	}

	middleware.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetBranding(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	tenant, err := h.tenantRepo.GetByID(r.Context(), tenantID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"branding": tenant.Branding,
	})
}

func (h *Handler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	tenant, err := h.tenantRepo.GetByID(r.Context(), tenantID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	var update struct {
		Branding json.RawMessage `json:"branding"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	tenant.Branding = update.Branding
	if err := h.tenantRepo.Update(r.Context(), tenant); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{"branding": tenant.Branding})
}
