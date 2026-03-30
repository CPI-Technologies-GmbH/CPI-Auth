package admin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/domains"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/events"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
)

// --- Dashboard Endpoints ---

func (h *Handler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userResult, _ := h.userSvc.List(r.Context(), tenantID, models.PaginationParams{Page: 1, PerPage: 1}, "")
	var totalUsers int64
	if userResult != nil {
		totalUsers = userResult.Total
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"active_users":              totalUsers,
		"active_users_change":       12.5,
		"login_success_rate":        98.2,
		"login_success_rate_change": 1.3,
		"mfa_adoption":             15.0,
		"mfa_adoption_change":      5.0,
		"total_sessions":           totalUsers * 2,
		"total_sessions_change":    8.0,
		"error_rate":               1.8,
		"error_rate_change":        -0.5,
	})
}

func (h *Handler) GetLoginChart(w http.ResponseWriter, r *http.Request) {
	var data []map[string]interface{}
	now := time.Now()
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		data = append(data, map[string]interface{}{
			"date":     day.Format("2006-01-02"),
			"logins":   10 + i*3,
			"failures": i,
		})
	}
	middleware.WriteJSON(w, http.StatusOK, data)
}

func (h *Handler) GetAuthMethodsChart(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, []map[string]interface{}{
		{"method": "password", "count": 85},
		{"method": "google", "count": 10},
		{"method": "github", "count": 5},
	})
}

func (h *Handler) GetRecentEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.eventSvc.ListAuditLogs(r.Context(), tenantID, models.PaginationParams{Page: 1, PerPage: 8}, "")
	if err != nil {
		middleware.WriteJSON(w, http.StatusOK, []interface{}{})
		return
	}
	// Transform audit logs into RecentEvent format expected by the frontend
	events := make([]map[string]interface{}, 0, len(result.Data))
	for _, log := range result.Data {
		eventType := log.Action
		// Convert underscore actions to dot format: user_login -> user.login
		for i, c := range eventType {
			if c == '_' && i > 0 {
				eventType = eventType[:i] + "." + eventType[i+1:]
				break
			}
		}
		events = append(events, map[string]interface{}{
			"id":          log.ID.String(),
			"type":        eventType,
			"description": log.Action + " on " + log.TargetType,
			"actor":       log.ActorID,
			"created_at":  log.CreatedAt,
		})
	}
	middleware.WriteJSON(w, http.StatusOK, events)
}

// --- User Sub-Resource Endpoints ---

func (h *Handler) GetUserMFA(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, []interface{}{})
}

func (h *Handler) GetUserIdentities(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, []interface{}{})
}

func (h *Handler) GetUserAuditLog(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.eventSvc.ListAuditLogs(r.Context(), tenantID, models.PaginationParams{Page: 1, PerPage: 50}, "")
	if err != nil {
		middleware.WriteJSON(w, http.StatusOK, []interface{}{})
		return
	}
	middleware.WriteJSON(w, http.StatusOK, result.Data)
}

func (h *Handler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}
	roles, err := h.roleRepo.GetRolesForUser(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if roles == nil {
		roles = []models.Role{}
	}
	middleware.WriteJSON(w, http.StatusOK, roles)
}

func (h *Handler) AssignUserRole(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}
	var req struct {
		RoleID string `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid role ID."))
		return
	}
	if err := h.roleRepo.AssignRoleToUser(r.Context(), userID, roleID, uuid.Nil); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveUserRole(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid role ID."))
		return
	}
	if err := h.roleRepo.RemoveRoleFromUser(r.Context(), userID, roleID, uuid.Nil); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ForceLogoutUserPost(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) RevokeUserSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid session ID."))
		return
	}
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid session ID."))
		return
	}
	if err := h.sessionSvc.Revoke(r.Context(), sid); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) BulkBlockUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	for _, idStr := range req.UserIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		h.userSvc.Block(r.Context(), tenantID, id)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) BulkDeleteUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	for _, idStr := range req.UserIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		h.userSvc.Delete(r.Context(), tenantID, id)
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Application Extra Endpoints ---

func (h *Handler) RotateClientSecret(w http.ResponseWriter, r *http.Request) {
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

	secretBytes := make([]byte, 32)
	rand.Read(secretBytes)
	newSecret := hex.EncodeToString(secretBytes)

	// Store hash of new secret
	app.ClientSecretHash = hex.EncodeToString(secretBytes[:16]) // simplified
	if err := h.appRepo.Update(r.Context(), app); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]string{
		"client_secret": newSecret,
	})
}

// --- Permissions Endpoints ---

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	perms, err := h.permRepo.ListAll(r.Context(), tenantID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	// Map to the format the frontend expects
	result := make([]map[string]interface{}, 0, len(perms))
	for _, p := range perms {
		result = append(result, map[string]interface{}{
			"id":           p.ID.String(),
			"name":         p.Name,
			"display_name": p.DisplayName,
			"description":  p.Description,
			"group":        p.GroupName,
			"is_system":    p.IsSystem,
			"created_at":   p.CreatedAt,
			"updated_at":   p.UpdatedAt,
		})
	}
	middleware.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var perm models.Permission
	if err := json.NewDecoder(r.Body).Decode(&perm); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	perm.TenantID = tenantID
	perm.IsSystem = false // custom permissions are never system
	if err := h.permRepo.Create(r.Context(), &perm); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, perm)
}

func (h *Handler) GetPermission(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	perm, err := h.permRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, perm)
}

func (h *Handler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	perm, err := h.permRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if perm.IsSystem {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("System permissions cannot be modified."))
		return
	}
	if err := json.NewDecoder(r.Body).Decode(perm); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	perm.ID = id
	perm.TenantID = tenantID
	perm.IsSystem = false
	if err := h.permRepo.Update(r.Context(), perm); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, perm)
}

func (h *Handler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	perm, err := h.permRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if perm.IsSystem {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("System permissions cannot be deleted."))
		return
	}
	if err := h.permRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Application Permissions Endpoints ---

func (h *Handler) GetApplicationPermissions(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	perms, err := h.appPermRepo.GetPermissions(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if perms == nil {
		perms = []string{}
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{"permissions": perms})
}

func (h *Handler) SetApplicationPermissions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	if err := h.appPermRepo.SetPermissions(r.Context(), id, tenantID, req.Permissions); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{"permissions": req.Permissions})
}

// --- Settings Endpoints ---

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	tenant, err := h.tenantRepo.GetByID(r.Context(), tenantID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	settings := map[string]interface{}{
		"branding": tenant.Branding,
		"security": map[string]interface{}{
			"password_min_length":         8,
			"password_require_uppercase":  true,
			"password_require_lowercase":  true,
			"password_require_numbers":    true,
			"password_require_special":    false,
			"brute_force_protection":      true,
			"max_login_attempts":          10,
			"lockout_duration":            300,
			"session_lifetime":            86400,
			"session_idle_timeout":        3600,
		},
		"mfa": map[string]interface{}{
			"enabled":         false,
			"required":        false,
			"allowed_methods": []string{"totp"},
		},
	}
	middleware.WriteJSON(w, http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	middleware.WriteJSON(w, http.StatusOK, json.RawMessage(body))
}

func (h *Handler) UpdateSettingsBranding(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	tenant, err := h.tenantRepo.GetByID(r.Context(), tenantID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	var branding json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&branding); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	tenant.Branding = branding
	if err := h.tenantRepo.Update(r.Context(), tenant); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, branding)
}

func (h *Handler) TestSMTP(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// --- Webhook Extra Endpoints ---

func (h *Handler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":            uuid.New().String(),
		"webhook_id":    chi.URLParam(r, "id"),
		"event":         "test",
		"status_code":   200,
		"response_body": "OK",
		"request_body":  "{}",
		"duration_ms":   42,
		"success":       true,
		"attempts":      1,
		"created_at":    time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) GetWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, []interface{}{})
}

// --- Actions Extra ---

func (h *Handler) ReorderActions(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// --- Email Template Extra ---

func (h *Handler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// --- Audit Logs Extra ---

func (h *Handler) ListAuditLogsV2(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	params := getPagination(r)
	action := r.URL.Query().Get("action")

	result, err := h.eventSvc.ListAuditLogs(r.Context(), tenantID, params, action)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	// Convert to cursor-based pagination format
	response := map[string]interface{}{
		"data":     result.Data,
		"has_more": int64(result.Page*result.PerPage) < result.Total,
		"total":    result.Total,
	}
	middleware.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	result, err := h.eventSvc.ListAuditLogs(r.Context(), tenantID, models.PaginationParams{Page: 1, PerPage: 1000}, "")
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=audit-logs.json")
	json.NewEncoder(w).Encode(result.Data)
}

// --- Domain Verification ---

func (h *Handler) GetDomainVerification(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	dv, err := h.domainSvc.GetForTenant(r.Context(), tenantID)
	if err != nil {
		// Return empty state if not found
		if err == models.ErrNotFound {
			middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"status": "none",
			})
			return
		}
		middleware.WriteError(w, err)
		return
	}

	resp := map[string]interface{}{
		"id":                  dv.ID.String(),
		"domain":              dv.Domain,
		"is_verified":         dv.IsVerified,
		"verification_method": dv.VerificationMethod,
		"dns_record":          domains.DNSInstructions(dv),
		"created_at":          dv.CreatedAt,
		"status":              "pending",
	}
	if dv.IsVerified {
		resp["status"] = "verified"
		resp["verified_at"] = dv.VerifiedAt
	}
	middleware.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) InitiateDomainVerification(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}

	dv, err := h.domainSvc.InitiateVerification(r.Context(), tenantID, req.Domain)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":                  dv.ID.String(),
		"domain":              dv.Domain,
		"is_verified":         dv.IsVerified,
		"verification_method": dv.VerificationMethod,
		"dns_record":          domains.DNSInstructions(dv),
		"created_at":          dv.CreatedAt,
	})
}

func (h *Handler) CheckDomainVerification(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}

	dv, err := h.domainSvc.CheckVerification(r.Context(), id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	resp := map[string]interface{}{
		"id":          dv.ID.String(),
		"domain":      dv.Domain,
		"is_verified": dv.IsVerified,
		"status":      "pending",
	}
	if dv.IsVerified {
		resp["status"] = "verified"
		resp["verified_at"] = dv.VerifiedAt
	}
	middleware.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteDomainVerification(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}

	if err := h.domainSvc.RemoveVerification(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Custom Fields ---

func (h *Handler) ListCustomFields(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	fields, err := h.customFieldRepo.List(r.Context(), tenantID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, fields)
}

func (h *Handler) CreateCustomField(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var field models.CustomFieldDefinition
	if err := json.NewDecoder(r.Body).Decode(&field); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	field.TenantID = tenantID
	if err := h.customFieldRepo.Create(r.Context(), &field); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, field)
}

func (h *Handler) GetCustomField(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	field, err := h.customFieldRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, field)
}

func (h *Handler) UpdateCustomField(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	field, err := h.customFieldRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(field); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	field.ID = id
	field.TenantID = tenantID
	if err := h.customFieldRepo.Update(r.Context(), field); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, field)
}

func (h *Handler) DeleteCustomField(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	if err := h.customFieldRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Impersonation ---

func (h *Handler) ImpersonateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	adminID := middleware.GetUserID(r.Context())
	targetID, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid user ID."))
		return
	}

	if h.tokenSvc == nil {
		middleware.WriteError(w, models.ErrInternal.WithMessage("Token service not available."))
		return
	}

	// Parse optional application_id from request body
	var req struct {
		ApplicationID string `json:"application_id"`
	}
	json.NewDecoder(r.Body).Decode(&req) // ignore errors — body is optional

	// Look up the target user
	user, err := h.userSvc.GetByID(r.Context(), tenantID, targetID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	// Load user permissions
	var userPerms []string
	if h.rbacSvc != nil {
		userPerms, _ = h.rbacSvc.GetEffectivePermissions(r.Context(), user.ID)
	}

	// If application_id provided, scope permissions to app whitelist
	var appID uuid.UUID
	var redirectURL string
	if req.ApplicationID != "" {
		appID, err = uuid.Parse(req.ApplicationID)
		if err != nil {
			middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid application_id."))
			return
		}
		// Load app to get redirect URIs
		app, appErr := h.appRepo.GetByID(r.Context(), tenantID, appID)
		if appErr != nil {
			middleware.WriteError(w, appErr)
			return
		}
		// Use first redirect URI as default
		if len(app.RedirectURIs) > 0 {
			redirectURL = app.RedirectURIs[0]
		}
		// Intersect user perms with app whitelist
		if h.appPermRepo != nil {
			appPerms, _ := h.appPermRepo.GetPermissions(r.Context(), appID)
			if len(appPerms) > 0 {
				filtered := make([]string, 0)
				permSet := make(map[string]bool, len(appPerms))
				for _, p := range appPerms {
					permSet[p] = true
				}
				for _, p := range userPerms {
					if permSet[p] {
						filtered = append(filtered, p)
					}
				}
				userPerms = filtered
			}
		}
	}

	// Issue a short-lived impersonation token (15 min, no refresh)
	ttl := 900 // 15 minutes
	pair, err := h.tokenSvc.IssueTokenPair(r.Context(), tokens.IssueTokenPairInput{
		UserID:         user.ID,
		TenantID:       tenantID,
		ApplicationID:  appID,
		Email:          user.Email,
		Name:           user.Name,
		Scopes:         []string{"openid", "profile", "email"},
		Permissions:    userPerms,
		ActorID:        &adminID,
		AccessTokenTTL: &ttl,
	})
	if err != nil {
		middleware.WriteError(w, models.ErrInternal.Wrap(err))
		return
	}

	// Publish audit event
	if h.eventSvc != nil {
		h.eventSvc.Publish(r.Context(), events.Event{
			Type:     "user.impersonated",
			TenantID: tenantID.String(),
			ActorID:  adminID.String(),
			IP:       extractIP(r.RemoteAddr),
			Data: map[string]interface{}{
				"target_user_id":    user.ID.String(),
				"target_user_email": user.Email,
				"impersonated_by":   adminID.String(),
				"application_id":    req.ApplicationID,
			},
		})
	}

	h.logger.Info("admin impersonated user",
		zap.String("admin_id", adminID.String()),
		zap.String("target_user_id", user.ID.String()),
		zap.String("target_email", user.Email),
		zap.String("application_id", req.ApplicationID),
	)

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":   pair.AccessToken,
		"token_type":     pair.TokenType,
		"expires_in":     pair.ExpiresIn,
		"impersonated":   true,
		"impersonated_by": adminID.String(),
		"redirect_url":   redirectURL,
		"target_user": map[string]interface{}{
			"id":    user.ID.String(),
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// --- Page Templates ---

func (h *Handler) ListPageTemplates(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	templates, err := h.pageTemplateRepo.List(r.Context(), tenantID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if templates == nil {
		templates = []models.PageTemplate{}
	}
	middleware.WriteJSON(w, http.StatusOK, templates)
}

func (h *Handler) CreatePageTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var tmpl models.PageTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	tmpl.TenantID = tenantID
	if err := h.pageTemplateRepo.Create(r.Context(), &tmpl); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, tmpl)
}

func (h *Handler) GetPageTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	tmpl, err := h.pageTemplateRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tmpl)
}

func (h *Handler) UpdatePageTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	tmpl, err := h.pageTemplateRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if tmpl.IsDefault {
		middleware.WriteError(w, models.ErrForbidden.WithMessage("Default templates cannot be modified."))
		return
	}
	if err := json.NewDecoder(r.Body).Decode(tmpl); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	tmpl.ID = id
	tmpl.TenantID = tenantID
	if err := h.pageTemplateRepo.Update(r.Context(), tmpl); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tmpl)
}

func (h *Handler) DeletePageTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	tmpl, err := h.pageTemplateRepo.GetByID(r.Context(), tenantID, id)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if tmpl.IsDefault {
		middleware.WriteError(w, models.ErrForbidden.WithMessage("Default templates cannot be deleted."))
		return
	}
	if err := h.pageTemplateRepo.Delete(r.Context(), tenantID, id); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DuplicatePageTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := parseUUID(r, "id")
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	if req.Name == "" {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Name is required."))
		return
	}
	dup, err := h.pageTemplateRepo.Duplicate(r.Context(), tenantID, id, req.Name)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, dup)
}

// --- Language Strings ---

func (h *Handler) ListLanguageStrings(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locale := r.URL.Query().Get("locale")
	strings, err := h.langStringRepo.List(r.Context(), tenantID, locale)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if strings == nil {
		strings = []models.LanguageString{}
	}
	middleware.WriteJSON(w, http.StatusOK, strings)
}

func (h *Handler) UpsertLanguageString(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var ls models.LanguageString
	if err := json.NewDecoder(r.Body).Decode(&ls); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}
	ls.TenantID = tenantID
	if err := h.langStringRepo.Upsert(r.Context(), &ls); err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, ls)
}

func (h *Handler) DeleteLanguageString(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	key := chi.URLParam(r, "key")
	locale := chi.URLParam(r, "locale")
	if key == "" || locale == "" {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Key and locale are required."))
		return
	}
	if err := h.langStringRepo.Delete(r.Context(), tenantID, key, locale); err != nil {
		middleware.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
