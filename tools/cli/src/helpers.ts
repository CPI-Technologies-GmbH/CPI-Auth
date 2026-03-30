import { readFileSync, writeFileSync, existsSync } from 'node:fs'
import { resolve } from 'node:path'
import { parse as parseYAML, stringify as stringifyYAML } from 'yaml'
import chalk from 'chalk'
import { getAuthenticatedClient, migrateOldToken } from './config.js'
import type { APIClient } from '@cpi-auth/sdk'

// ─── Logging ──────────────────────────────────────────────────

export function log(msg: string) { console.log(msg) }
export function info(msg: string) { console.log(chalk.blue('ℹ'), msg) }
export function success(msg: string) { console.log(chalk.green('✓'), msg) }
export function warn(msg: string) { console.log(chalk.yellow('⚠'), msg) }
export function error(msg: string) { console.error(chalk.red('✗'), msg) }

// ─── YAML Helpers ─────────────────────────────────────────────

export function loadYAML<T = unknown>(path: string): T {
  return parseYAML(readFileSync(path, 'utf8'))
}

export function saveYAML(path: string, data: unknown) {
  writeFileSync(path, stringifyYAML(data, { lineWidth: 120 }), 'utf8')
}

// ─── Auth Client ──────────────────────────────────────────────

export function getClient(opts?: { server?: string; token?: string }): APIClient {
  // Auto-migrate old .cpi-auth-token on first use
  migrateOldToken()

  try {
    return getAuthenticatedClient(opts)
  } catch (e: any) {
    error(e.message)
    process.exit(1)
  }
}

// ─── Project Config (template development only) ───────────────

export interface ProjectConfig {
  tokens?: Record<string, Record<string, string>>
  templates?: Record<string, { html: string; css: string }>
  locales?: string[]
  preview?: {
    custom_fields?: Array<{ label: string; type: string; required?: boolean; placeholder?: string; options?: string[] }>
    sample_data?: Record<string, string>
  }
  // Legacy fields (ignored, kept for backward compat)
  server?: string
  tenant_id?: string
}

export function findProjectConfig(): ProjectConfig {
  const configPath = resolve('cpi-auth.config.yaml')
  if (!existsSync(configPath)) {
    error('cpi-auth.config.yaml not found in current directory.')
    error('Run `cpi-auth templates init` to create a template project.')
    process.exit(1)
  }
  return parseYAML(readFileSync(configPath, 'utf8'))
}
