import type {
  AuthTokens,
  AdminUser,
  User,
  UserSession,
  UserMfaEnrollment,
  LinkedIdentity,
  AuditLogEntry,
  Application,
  Tenant,
  TenantSettings,
  BrandingConfig,
  Organization,
  OrganizationMember,
  Role,
  Permission,
  Webhook,
  WebhookDelivery,
  Action,
  EmailTemplate,
  ApiKey,
  CustomFieldDefinition,
  DomainVerification,
  PageTemplate,
  LanguageString,
  DashboardMetrics,
  LoginChartData,
  AuthMethodData,
  RecentEvent,
  PaginatedResponse,
  ApiError,
} from '@/types'

const API_URL = import.meta.env.VITE_API_URL || ''

class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  private getAccessToken(): string | null {
    return localStorage.getItem('access_token')
  }

  private getRefreshToken(): string | null {
    return localStorage.getItem('refresh_token')
  }

  private setTokens(tokens: AuthTokens): void {
    localStorage.setItem('access_token', tokens.access_token)
    localStorage.setItem('refresh_token', tokens.refresh_token)
    localStorage.setItem('token_expires_at', String(Date.now() + tokens.expires_in * 1000))
  }

  private clearTokens(): void {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('token_expires_at')
  }

  private isTokenExpired(): boolean {
    const expiresAt = localStorage.getItem('token_expires_at')
    if (!expiresAt) return true
    return Date.now() >= Number(expiresAt) - 30000 // 30s buffer
  }

  private async refreshAccessToken(): Promise<boolean> {
    const refreshToken = this.getRefreshToken()
    if (!refreshToken) return false

    try {
      const response = await fetch(`${this.baseUrl}/admin/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })

      if (!response.ok) {
        this.clearTokens()
        return false
      }

      const tokens: AuthTokens = await response.json()
      this.setTokens(tokens)
      return true
    } catch {
      this.clearTokens()
      return false
    }
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    if (this.isTokenExpired() && this.getRefreshToken()) {
      const refreshed = await this.refreshAccessToken()
      if (!refreshed) {
        window.location.href = '/admin/login'
        throw new Error('Session expired')
      }
    }

    const token = this.getAccessToken()
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const activeTenantId = localStorage.getItem('active_tenant_id')
    if (activeTenantId) {
      headers['X-Tenant-ID'] = activeTenantId
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers,
    })

    if (response.status === 401) {
      const refreshed = await this.refreshAccessToken()
      if (refreshed) {
        headers['Authorization'] = `Bearer ${this.getAccessToken()}`
        const retryResponse = await fetch(`${this.baseUrl}${path}`, {
          ...options,
          headers,
        })
        if (!retryResponse.ok) {
          const error: ApiError = await retryResponse.json().catch(() => ({
            error: 'request_failed',
            message: retryResponse.statusText,
            status_code: retryResponse.status,
          }))
          throw error
        }
        return retryResponse.json()
      }
      this.clearTokens()
      window.location.href = '/login'
      throw new Error('Session expired')
    }

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        error: 'request_failed',
        message: response.statusText,
        status_code: response.status,
      }))
      throw error
    }

    if (response.status === 204) {
      return undefined as T
    }

    return response.json()
  }

  // ==================== Auth ====================
  async login(email: string, password: string): Promise<AuthTokens> {
    const response = await fetch(`${this.baseUrl}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({
        error: 'login_failed',
        message: 'Invalid credentials',
        status_code: response.status,
      }))
      throw error
    }

    const tokens: AuthTokens = await response.json()
    this.setTokens(tokens)
    return tokens
  }

  async logout(): Promise<void> {
    try {
      await this.request('/admin/auth/logout', { method: 'POST' })
    } finally {
      this.clearTokens()
    }
  }

  async getMe(): Promise<AdminUser> {
    return this.request<AdminUser>('/admin/auth/me')
  }

  // ==================== Dashboard ====================
  async getDashboardMetrics(): Promise<DashboardMetrics> {
    return this.request<DashboardMetrics>('/admin/dashboard/metrics')
  }

  async getLoginChart(period: '7d' | '30d'): Promise<LoginChartData[]> {
    return this.request<LoginChartData[]>(`/admin/dashboard/logins?period=${period}`)
  }

  async getAuthMethodsChart(): Promise<AuthMethodData[]> {
    return this.request<AuthMethodData[]>('/admin/dashboard/auth-methods')
  }

  async getRecentEvents(): Promise<RecentEvent[]> {
    return this.request<RecentEvent[]>('/admin/dashboard/events')
  }

  // ==================== Users ====================
  async getUsers(params?: {
    cursor?: string
    limit?: number
    search?: string
    status?: string
    sort?: string
    order?: string
  }): Promise<PaginatedResponse<User>> {
    const searchParams = new URLSearchParams()
    if (params?.cursor) searchParams.set('cursor', params.cursor)
    if (params?.limit) searchParams.set('limit', String(params.limit))
    if (params?.search) searchParams.set('search', params.search)
    if (params?.status) searchParams.set('status', params.status)
    if (params?.sort) searchParams.set('sort', params.sort)
    if (params?.order) searchParams.set('order', params.order)
    const qs = searchParams.toString()
    return this.request<PaginatedResponse<User>>(`/admin/users${qs ? `?${qs}` : ''}`)
  }

  async getUser(id: string): Promise<User> {
    return this.request<User>(`/admin/users/${id}`)
  }

  async createUser(data: Partial<User> & { password?: string }): Promise<User> {
    return this.request<User>('/admin/users', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateUser(id: string, data: Partial<User>): Promise<User> {
    return this.request<User>(`/admin/users/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteUser(id: string): Promise<void> {
    return this.request<void>(`/admin/users/${id}`, { method: 'DELETE' })
  }

  async blockUser(id: string): Promise<User> {
    return this.request<User>(`/admin/users/${id}/block`, { method: 'POST' })
  }

  async unblockUser(id: string): Promise<User> {
    return this.request<User>(`/admin/users/${id}/unblock`, { method: 'POST' })
  }

  async resetUserPassword(id: string): Promise<void> {
    return this.request<void>(`/admin/users/${id}/reset-password`, { method: 'POST' })
  }

  async forceLogout(id: string): Promise<void> {
    return this.request<void>(`/admin/users/${id}/force-logout`, { method: 'POST' })
  }

  async getUserSessions(id: string): Promise<UserSession[]> {
    return this.request<UserSession[]>(`/admin/users/${id}/sessions`)
  }

  async revokeUserSession(userId: string, sessionId: string): Promise<void> {
    return this.request<void>(`/admin/users/${userId}/sessions/${sessionId}`, {
      method: 'DELETE',
    })
  }

  async getUserMfaEnrollments(id: string): Promise<UserMfaEnrollment[]> {
    return this.request<UserMfaEnrollment[]>(`/admin/users/${id}/mfa`)
  }

  async getUserIdentities(id: string): Promise<LinkedIdentity[]> {
    return this.request<LinkedIdentity[]>(`/admin/users/${id}/identities`)
  }

  async getUserAuditLog(id: string): Promise<AuditLogEntry[]> {
    return this.request<AuditLogEntry[]>(`/admin/users/${id}/audit-log`)
  }

  async getUserRoles(id: string): Promise<Role[]> {
    return this.request<Role[]>(`/admin/users/${id}/roles`)
  }

  async assignUserRole(userId: string, roleId: string): Promise<void> {
    return this.request<void>(`/admin/users/${userId}/roles`, {
      method: 'POST',
      body: JSON.stringify({ role_id: roleId }),
    })
  }

  async removeUserRole(userId: string, roleId: string): Promise<void> {
    return this.request<void>(`/admin/users/${userId}/roles/${roleId}`, {
      method: 'DELETE',
    })
  }

  async impersonateUser(userId: string, applicationId?: string): Promise<{
    access_token: string; expires_in: number; impersonated: boolean;
    impersonated_by: string; redirect_url: string;
    target_user: { id: string; email: string; name: string }
  }> {
    return this.request(`/admin/users/${userId}/impersonate`, {
      method: 'POST',
      body: applicationId ? JSON.stringify({ application_id: applicationId }) : undefined,
    })
  }

  async bulkBlockUsers(ids: string[]): Promise<void> {
    return this.request<void>('/admin/users/bulk/block', {
      method: 'POST',
      body: JSON.stringify({ user_ids: ids }),
    })
  }

  async bulkDeleteUsers(ids: string[]): Promise<void> {
    return this.request<void>('/admin/users/bulk/delete', {
      method: 'POST',
      body: JSON.stringify({ user_ids: ids }),
    })
  }

  async exportUsers(params?: { status?: string; search?: string }): Promise<Blob> {
    const searchParams = new URLSearchParams()
    if (params?.status) searchParams.set('status', params.status)
    if (params?.search) searchParams.set('search', params.search)
    const qs = searchParams.toString()
    const response = await fetch(
      `${this.baseUrl}/admin/users/export${qs ? `?${qs}` : ''}`,
      {
        headers: { Authorization: `Bearer ${this.getAccessToken()}` },
      }
    )
    return response.blob()
  }

  // ==================== Applications ====================
  async getApplications(): Promise<Application[]> {
    return this.request<Application[]>('/admin/applications')
  }

  async getApplication(id: string): Promise<Application> {
    return this.request<Application>(`/admin/applications/${id}`)
  }

  async createApplication(data: Partial<Application>): Promise<Application> {
    return this.request<Application>('/admin/applications', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateApplication(id: string, data: Partial<Application>): Promise<Application> {
    return this.request<Application>(`/admin/applications/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteApplication(id: string): Promise<void> {
    return this.request<void>(`/admin/applications/${id}`, { method: 'DELETE' })
  }

  async rotateClientSecret(id: string): Promise<{ client_secret: string }> {
    return this.request<{ client_secret: string }>(`/admin/applications/${id}/rotate-secret`, {
      method: 'POST',
    })
  }

  // ==================== Tenants ====================
  async getTenants(): Promise<Tenant[]> {
    return this.request<Tenant[]>('/admin/tenants')
  }

  async getTenant(id: string): Promise<Tenant> {
    return this.request<Tenant>(`/admin/tenants/${id}`)
  }

  async createTenant(data: Partial<Tenant>): Promise<Tenant> {
    return this.request<Tenant>('/admin/tenants', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateTenant(id: string, data: Partial<Tenant>): Promise<Tenant> {
    return this.request<Tenant>(`/admin/tenants/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteTenant(id: string): Promise<void> {
    return this.request<void>(`/admin/tenants/${id}`, { method: 'DELETE' })
  }

  // ==================== Organizations ====================
  async getOrganizations(): Promise<Organization[]> {
    return this.request<Organization[]>('/admin/organizations')
  }

  async getOrganization(id: string): Promise<Organization> {
    return this.request<Organization>(`/admin/organizations/${id}`)
  }

  async createOrganization(data: Partial<Organization>): Promise<Organization> {
    return this.request<Organization>('/admin/organizations', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateOrganization(id: string, data: Partial<Organization>): Promise<Organization> {
    return this.request<Organization>(`/admin/organizations/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteOrganization(id: string): Promise<void> {
    return this.request<void>(`/admin/organizations/${id}`, { method: 'DELETE' })
  }

  async getOrganizationMembers(id: string): Promise<OrganizationMember[]> {
    return this.request<OrganizationMember[]>(`/admin/organizations/${id}/members`)
  }

  async addOrganizationMember(orgId: string, data: { user_id: string; role: string }): Promise<OrganizationMember> {
    return this.request<OrganizationMember>(`/admin/organizations/${orgId}/members`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async removeOrganizationMember(orgId: string, memberId: string): Promise<void> {
    return this.request<void>(`/admin/organizations/${orgId}/members/${memberId}`, {
      method: 'DELETE',
    })
  }

  // ==================== Roles & Permissions ====================
  async getRoles(): Promise<Role[]> {
    return this.request<Role[]>('/admin/roles')
  }

  async getRole(id: string): Promise<Role> {
    return this.request<Role>(`/admin/roles/${id}`)
  }

  async createRole(data: Partial<Role>): Promise<Role> {
    return this.request<Role>('/admin/roles', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateRole(id: string, data: Partial<Role>): Promise<Role> {
    return this.request<Role>(`/admin/roles/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteRole(id: string): Promise<void> {
    return this.request<void>(`/admin/roles/${id}`, { method: 'DELETE' })
  }

  async getPermissions(): Promise<Permission[]> {
    return this.request<Permission[]>('/admin/permissions')
  }

  async createPermission(data: { name: string; display_name: string; description?: string; group_name: string }): Promise<Permission> {
    return this.request<Permission>('/admin/permissions', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updatePermission(id: string, data: Partial<Permission>): Promise<Permission> {
    return this.request<Permission>(`/admin/permissions/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deletePermission(id: string): Promise<void> {
    return this.request<void>(`/admin/permissions/${id}`, { method: 'DELETE' })
  }

  async getApplicationPermissions(id: string): Promise<{ permissions: string[] }> {
    return this.request<{ permissions: string[] }>(`/admin/applications/${id}/permissions`)
  }

  async setApplicationPermissions(id: string, permissions: string[]): Promise<{ permissions: string[] }> {
    return this.request<{ permissions: string[] }>(`/admin/applications/${id}/permissions`, {
      method: 'PUT',
      body: JSON.stringify({ permissions }),
    })
  }

  // ==================== Webhooks ====================
  async getWebhooks(): Promise<Webhook[]> {
    return this.request<Webhook[]>('/admin/webhooks')
  }

  async getWebhook(id: string): Promise<Webhook> {
    return this.request<Webhook>(`/admin/webhooks/${id}`)
  }

  async createWebhook(data: Partial<Webhook>): Promise<Webhook> {
    return this.request<Webhook>('/admin/webhooks', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateWebhook(id: string, data: Partial<Webhook>): Promise<Webhook> {
    return this.request<Webhook>(`/admin/webhooks/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteWebhook(id: string): Promise<void> {
    return this.request<void>(`/admin/webhooks/${id}`, { method: 'DELETE' })
  }

  async testWebhook(id: string): Promise<WebhookDelivery> {
    return this.request<WebhookDelivery>(`/admin/webhooks/${id}/test`, { method: 'POST' })
  }

  async getWebhookDeliveries(id: string): Promise<WebhookDelivery[]> {
    return this.request<WebhookDelivery[]>(`/admin/webhooks/${id}/deliveries`)
  }

  // ==================== Actions ====================
  async getActions(): Promise<Action[]> {
    return this.request<Action[]>('/admin/actions')
  }

  async getAction(id: string): Promise<Action> {
    return this.request<Action>(`/admin/actions/${id}`)
  }

  async createAction(data: Partial<Action>): Promise<Action> {
    return this.request<Action>('/admin/actions', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateAction(id: string, data: Partial<Action>): Promise<Action> {
    return this.request<Action>(`/admin/actions/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteAction(id: string): Promise<void> {
    return this.request<void>(`/admin/actions/${id}`, { method: 'DELETE' })
  }

  async reorderActions(trigger: string, actionIds: string[]): Promise<void> {
    return this.request<void>('/admin/actions/reorder', {
      method: 'POST',
      body: JSON.stringify({ trigger, action_ids: actionIds }),
    })
  }

  // ==================== Email Templates ====================
  async getEmailTemplates(): Promise<EmailTemplate[]> {
    return this.request<EmailTemplate[]>('/admin/email-templates')
  }

  async getEmailTemplate(id: string): Promise<EmailTemplate> {
    return this.request<EmailTemplate>(`/admin/email-templates/${id}`)
  }

  async updateEmailTemplate(id: string, data: Partial<EmailTemplate>): Promise<EmailTemplate> {
    return this.request<EmailTemplate>(`/admin/email-templates/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async sendTestEmail(id: string, email: string): Promise<void> {
    return this.request<void>(`/admin/email-templates/${id}/test`, {
      method: 'POST',
      body: JSON.stringify({ email }),
    })
  }

  // ==================== API Keys ====================
  async getApiKeys(): Promise<ApiKey[]> {
    return this.request<ApiKey[]>('/admin/api-keys')
  }

  async createApiKey(data: {
    name: string
    scopes: string[]
    rate_limit: number
    expires_at?: string
  }): Promise<ApiKey & { key: string }> {
    return this.request<ApiKey & { key: string }>('/admin/api-keys', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async revokeApiKey(id: string): Promise<void> {
    return this.request<void>(`/admin/api-keys/${id}`, { method: 'DELETE' })
  }

  // ==================== Domain Verification ====================
  async getDomainVerification(): Promise<DomainVerification> {
    return this.request<DomainVerification>('/admin/domains/verification')
  }

  async initiateDomainVerification(domain: string): Promise<DomainVerification> {
    return this.request<DomainVerification>('/admin/domains/verification', {
      method: 'POST',
      body: JSON.stringify({ domain }),
    })
  }

  async checkDomainVerification(id: string): Promise<DomainVerification> {
    return this.request<DomainVerification>(`/admin/domains/verification/${id}/check`, {
      method: 'POST',
    })
  }

  async deleteDomainVerification(id: string): Promise<void> {
    return this.request<void>(`/admin/domains/verification/${id}`, { method: 'DELETE' })
  }

  // ==================== Custom Fields ====================
  async getCustomFields(): Promise<CustomFieldDefinition[]> {
    return this.request<CustomFieldDefinition[]>('/admin/custom-fields')
  }

  async getCustomField(id: string): Promise<CustomFieldDefinition> {
    return this.request<CustomFieldDefinition>(`/admin/custom-fields/${id}`)
  }

  async createCustomField(data: Partial<CustomFieldDefinition>): Promise<CustomFieldDefinition> {
    return this.request<CustomFieldDefinition>('/admin/custom-fields', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateCustomField(id: string, data: Partial<CustomFieldDefinition>): Promise<CustomFieldDefinition> {
    return this.request<CustomFieldDefinition>(`/admin/custom-fields/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteCustomField(id: string): Promise<void> {
    return this.request<void>(`/admin/custom-fields/${id}`, { method: 'DELETE' })
  }

  // ==================== Page Templates ====================
  async getPageTemplates(): Promise<PageTemplate[]> {
    return this.request<PageTemplate[]>('/admin/page-templates')
  }

  async getPageTemplate(id: string): Promise<PageTemplate> {
    return this.request<PageTemplate>(`/admin/page-templates/${id}`)
  }

  async createPageTemplate(data: Partial<PageTemplate>): Promise<PageTemplate> {
    return this.request<PageTemplate>('/admin/page-templates', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updatePageTemplate(id: string, data: Partial<PageTemplate>): Promise<PageTemplate> {
    return this.request<PageTemplate>(`/admin/page-templates/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deletePageTemplate(id: string): Promise<void> {
    return this.request<void>(`/admin/page-templates/${id}`, { method: 'DELETE' })
  }

  async duplicatePageTemplate(id: string, name: string): Promise<PageTemplate> {
    return this.request<PageTemplate>(`/admin/page-templates/${id}/duplicate`, {
      method: 'POST',
      body: JSON.stringify({ name }),
    })
  }

  async getLanguageStrings(locale = 'en'): Promise<LanguageString[]> {
    return this.request<LanguageString[]>(`/admin/language-strings?locale=${locale}`)
  }

  async upsertLanguageString(data: { string_key: string; locale: string; value: string }): Promise<LanguageString> {
    return this.request<LanguageString>('/admin/language-strings', {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async deleteLanguageString(key: string, locale: string): Promise<void> {
    return this.request<void>(`/admin/language-strings/${encodeURIComponent(key)}/${locale}`, { method: 'DELETE' })
  }

  // ==================== Audit Logs ====================
  async getAuditLogs(params?: {
    cursor?: string
    limit?: number
    action?: string
    actor_id?: string
    target_id?: string
    date_from?: string
    date_to?: string
    tenant_id?: string
  }): Promise<PaginatedResponse<AuditLogEntry>> {
    const searchParams = new URLSearchParams()
    if (params?.cursor) searchParams.set('cursor', params.cursor)
    if (params?.limit) searchParams.set('limit', String(params.limit))
    if (params?.action) searchParams.set('action', params.action)
    if (params?.actor_id) searchParams.set('actor_id', params.actor_id)
    if (params?.target_id) searchParams.set('target_id', params.target_id)
    if (params?.date_from) searchParams.set('date_from', params.date_from)
    if (params?.date_to) searchParams.set('date_to', params.date_to)
    if (params?.tenant_id) searchParams.set('tenant_id', params.tenant_id)
    const qs = searchParams.toString()
    return this.request<PaginatedResponse<AuditLogEntry>>(`/admin/audit-logs${qs ? `?${qs}` : ''}`)
  }

  async exportAuditLogs(params?: {
    action?: string
    date_from?: string
    date_to?: string
  }): Promise<Blob> {
    const searchParams = new URLSearchParams()
    if (params?.action) searchParams.set('action', params.action)
    if (params?.date_from) searchParams.set('date_from', params.date_from)
    if (params?.date_to) searchParams.set('date_to', params.date_to)
    const qs = searchParams.toString()
    const response = await fetch(
      `${this.baseUrl}/admin/audit-logs/export${qs ? `?${qs}` : ''}`,
      {
        headers: { Authorization: `Bearer ${this.getAccessToken()}` },
      }
    )
    return response.blob()
  }

  // ==================== Settings ====================
  async getSettings(): Promise<TenantSettings> {
    return this.request<TenantSettings>('/admin/settings')
  }

  async updateSettings(data: Partial<TenantSettings>): Promise<TenantSettings> {
    return this.request<TenantSettings>('/admin/settings', {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async updateBranding(data: Partial<BrandingConfig>): Promise<BrandingConfig> {
    return this.request<BrandingConfig>('/admin/settings/branding', {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async testSmtp(config: Record<string, string>): Promise<{ success: boolean }> {
    return this.request<{ success: boolean }>('/admin/settings/smtp/test', {
      method: 'POST',
      body: JSON.stringify(config),
    })
  }
}

export const api = new ApiClient(API_URL)
