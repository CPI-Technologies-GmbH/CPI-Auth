import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'
import { homedir } from 'node:os'
import { createHash } from 'node:crypto'
import { parse as parseYAML, stringify as stringifyYAML } from 'yaml'
import { APIClient } from '@cpi-auth/sdk'

// ─── Types ────────────────────────────────────────────────────

export interface ContextConfig {
  server: string
  'tenant-id'?: string
}

export interface GlobalConfig {
  'current-context': string
  contexts: Record<string, ContextConfig>
}

export interface StoredToken {
  access_token: string
  refresh_token?: string
  expires_at: number
  server: string
  tenant_id?: string
  email?: string
  name?: string
}

// ─── Paths ────────────────────────────────────────────────────

export function getConfigDir(): string {
  const dir = join(homedir(), '.cpi-auth')
  if (!existsSync(dir)) {
    mkdirSync(dir, { mode: 0o700, recursive: true })
  }
  const tokensDir = join(dir, 'tokens')
  if (!existsSync(tokensDir)) {
    mkdirSync(tokensDir, { mode: 0o700, recursive: true })
  }
  return dir
}

function getConfigPath(): string {
  return join(getConfigDir(), 'config.yaml')
}

function serverHash(server: string, tenantId?: string): string {
  const key = server.replace(/\/$/, '') + (tenantId ? `|${tenantId}` : '')
  return createHash('sha256').update(key).digest('hex').substring(0, 16)
}

function getTokenPath(server: string, tenantId?: string): string {
  return join(getConfigDir(), 'tokens', `${serverHash(server, tenantId)}.json`)
}

// ─── Global Config ────────────────────────────────────────────

export function loadGlobalConfig(): GlobalConfig {
  const path = getConfigPath()
  if (!existsSync(path)) {
    return { 'current-context': '', contexts: {} }
  }
  const raw = parseYAML(readFileSync(path, 'utf8'))
  return {
    'current-context': raw?.['current-context'] || '',
    contexts: raw?.contexts || {},
  }
}

export function saveGlobalConfig(config: GlobalConfig): void {
  const path = getConfigPath()
  writeFileSync(path, stringifyYAML(config, { lineWidth: 120 }), { mode: 0o600 })
}

export function getCurrentContext(): { name: string; server: string; tenantId?: string } | null {
  const config = loadGlobalConfig()
  const name = config['current-context']
  if (!name || !config.contexts[name]) return null
  const ctx = config.contexts[name]
  return { name, server: ctx.server, tenantId: ctx['tenant-id'] }
}

export function addContext(name: string, server: string, tenantId?: string): void {
  const config = loadGlobalConfig()
  config.contexts[name] = { server: server.replace(/\/$/, ''), ...(tenantId ? { 'tenant-id': tenantId } : {}) }
  if (!config['current-context']) {
    config['current-context'] = name
  }
  saveGlobalConfig(config)
}

export function useContext(name: string): boolean {
  const config = loadGlobalConfig()
  if (!config.contexts[name]) return false
  config['current-context'] = name
  saveGlobalConfig(config)
  return true
}

export function removeContext(name: string): boolean {
  const config = loadGlobalConfig()
  if (!config.contexts[name]) return false
  delete config.contexts[name]
  if (config['current-context'] === name) {
    const remaining = Object.keys(config.contexts)
    config['current-context'] = remaining[0] || ''
  }
  saveGlobalConfig(config)
  return true
}

// ─── Token Storage ────────────────────────────────────────────

export function loadToken(server: string, tenantId?: string): StoredToken | null {
  const path = getTokenPath(server, tenantId)
  if (!existsSync(path)) return null
  try {
    const data: StoredToken = JSON.parse(readFileSync(path, 'utf8'))
    if (data.expires_at && Date.now() / 1000 > data.expires_at) {
      return null // expired
    }
    return data
  } catch {
    return null
  }
}

export function saveTokenToFile(server: string, data: StoredToken, tenantId?: string): void {
  const path = getTokenPath(server, tenantId ?? data.tenant_id)
  writeFileSync(path, JSON.stringify(data, null, 2), { mode: 0o600 })
}

export function deleteToken(server: string, tenantId?: string): void {
  const path = getTokenPath(server, tenantId)
  if (existsSync(path)) {
    const { unlinkSync } = require('node:fs')
    unlinkSync(path)
  }
}

// ─── Authenticated Client ─────────────────────────────────────

export function getAuthenticatedClient(opts?: { server?: string; token?: string }): APIClient {
  // Priority: flags > env vars > current context
  let server = opts?.server || process.env.CPI_AUTH_SERVER || ''
  let token = opts?.token || process.env.CPI_AUTH_TOKEN || ''
  let tenantId = process.env.CPI_AUTH_TENANT_ID || ''

  if (!server) {
    const ctx = getCurrentContext()
    if (ctx) {
      server = ctx.server
      if (ctx.tenantId) tenantId = ctx.tenantId
    }
  }

  if (!server) {
    throw new Error('No server configured. Run `cpi-auth login --server <url>` or `cpi-auth config add-context <name> --server <url>`.')
  }

  const client = new APIClient({ server, tenantId: tenantId || undefined })

  if (token) {
    client.setToken(token)
    return client
  }

  // Load saved token for this server + tenant combination
  const stored = loadToken(server, tenantId || undefined)
  if (stored) {
    client.setToken(stored.access_token)
  }

  return client
}

// ─── Migration Helper ─────────────────────────────────────────

export function migrateOldToken(): void {
  // Check for old .cpi-auth-token in CWD
  const oldPath = join(process.cwd(), '.cpi-auth-token')
  if (!existsSync(oldPath)) return

  const token = readFileSync(oldPath, 'utf8').trim()
  if (!token) return

  try {
    // Decode JWT to get server and email
    const parts = token.split('.')
    if (parts.length !== 3) return
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString())

    const server = payload.iss || ''
    if (!server) return

    // Save to global config
    const stored: StoredToken = {
      access_token: token,
      expires_at: payload.exp || 0,
      server,
      tenant_id: payload.tenant_id,
      email: payload.email,
      name: payload.name,
    }
    saveTokenToFile(server, stored, payload.tenant_id)

    // Auto-create context if none exists
    const config = loadGlobalConfig()
    const existingCtx = Object.values(config.contexts).find(c => c.server === server)
    if (!existingCtx) {
      const name = server.includes('localhost') ? 'local' : 'default'
      addContext(name, server, payload.tenant_id)
    }

    console.log(`\x1b[33m⚠\x1b[0m Migrated credentials from .cpi-auth-token to ~/.cpi-auth/`)
    console.log(`  You can safely delete .cpi-auth-token`)
  } catch {
    // Silently ignore migration errors
  }
}
