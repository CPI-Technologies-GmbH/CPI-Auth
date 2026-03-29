import type { SDKConfig, AuthTokens, PageTemplate, LanguageString } from './types.js'

export class APIClient {
  private server: string
  private tenantId?: string
  private token?: string
  private refreshToken?: string
  private tokenExpiresAt = 0
  private credentials?: { email: string; password: string }

  constructor(config: SDKConfig) {
    this.server = config.server.replace(/\/$/, '')
    this.tenantId = config.tenantId
    this.token = config.token
    this.credentials = config.credentials
  }

  async ensureAuth(): Promise<void> {
    if (this.token && Date.now() < this.tokenExpiresAt - 30_000) return
    if (this.refreshToken) {
      try {
        await this.refresh()
        return
      } catch { /* fall through to login */ }
    }
    if (this.credentials) {
      await this.login(this.credentials.email, this.credentials.password)
      return
    }
    if (!this.token) {
      throw new Error('No authentication configured. Call login() or provide token/credentials.')
    }
  }

  async login(email: string, password: string): Promise<AuthTokens> {
    const res = await fetch(`${this.server}/admin/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) throw new Error(`Login failed: ${res.status} ${await res.text()}`)
    const tokens: AuthTokens = await res.json()
    this.token = tokens.access_token
    this.refreshToken = tokens.refresh_token
    this.tokenExpiresAt = Date.now() + tokens.expires_in * 1000
    return tokens
  }

  private async refresh(): Promise<void> {
    const res = await fetch(`${this.server}/admin/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: this.refreshToken }),
    })
    if (!res.ok) throw new Error('Token refresh failed')
    const tokens: AuthTokens = await res.json()
    this.token = tokens.access_token
    this.refreshToken = tokens.refresh_token
    this.tokenExpiresAt = Date.now() + tokens.expires_in * 1000
  }

  setToken(token: string): void {
    this.token = token
    this.tokenExpiresAt = Date.now() + 3600_000
  }

  async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    await this.ensureAuth()
    const headers: Record<string, string> = {
      'Authorization': `Bearer ${this.token}`,
      'Content-Type': 'application/json',
    }
    if (this.tenantId) headers['X-Tenant-ID'] = this.tenantId

    const res = await fetch(`${this.server}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    })

    if (res.status === 204) return null as T
    if (!res.ok) {
      const text = await res.text()
      throw new Error(`API ${method} ${path} failed: ${res.status} ${text}`)
    }
    return res.json()
  }

  // ─── Templates ──────────────────────────────────────────

  listTemplates(): Promise<PageTemplate[]> {
    return this.request('GET', '/admin/page-templates')
  }

  getTemplate(id: string): Promise<PageTemplate> {
    return this.request('GET', `/admin/page-templates/${id}`)
  }

  createTemplate(data: Partial<PageTemplate>): Promise<PageTemplate> {
    return this.request('POST', '/admin/page-templates', data)
  }

  updateTemplate(id: string, data: Partial<PageTemplate>): Promise<PageTemplate> {
    return this.request('PATCH', `/admin/page-templates/${id}`, data)
  }

  deleteTemplate(id: string): Promise<void> {
    return this.request('DELETE', `/admin/page-templates/${id}`)
  }

  duplicateTemplate(id: string, name: string): Promise<PageTemplate> {
    return this.request('POST', `/admin/page-templates/${id}/duplicate`, { name })
  }

  // ─── Language Strings ───────────────────────────────────

  listStrings(locale = 'en'): Promise<LanguageString[]> {
    return this.request('GET', `/admin/language-strings?locale=${locale}`)
  }

  upsertString(data: { string_key: string; locale: string; value: string }): Promise<LanguageString> {
    return this.request('PUT', '/admin/language-strings', data)
  }

  deleteString(key: string, locale: string): Promise<void> {
    return this.request('DELETE', `/admin/language-strings/${encodeURIComponent(key)}/${locale}`)
  }
}
