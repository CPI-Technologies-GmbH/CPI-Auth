import { readFileSync, writeFileSync, existsSync } from 'node:fs'
import { resolve, join } from 'node:path'
import { parse as parseYAML, stringify as stringifyYAML } from 'yaml'
import chalk from 'chalk'
import { APIClient, loadConfig, type ProjectConfig } from '@cpi-auth/sdk'

const TOKEN_FILE = '.cpi-auth-token'

export function log(msg: string) { console.log(msg) }
export function info(msg: string) { console.log(chalk.blue('ℹ'), msg) }
export function success(msg: string) { console.log(chalk.green('✓'), msg) }
export function warn(msg: string) { console.log(chalk.yellow('⚠'), msg) }
export function error(msg: string) { console.error(chalk.red('✗'), msg) }

export function findConfig(): ProjectConfig {
  const configPath = resolve('cpi-auth.config.yaml')
  if (!existsSync(configPath)) {
    error('cpi-auth.config.yaml not found. Run `cpi-auth init` first.')
    process.exit(1)
  }
  return loadConfig(configPath)
}

export function getClient(config: ProjectConfig): APIClient {
  const tokenPath = resolve(TOKEN_FILE)
  const client = new APIClient({
    server: config.server,
    tenantId: config.tenant_id,
  })

  if (existsSync(tokenPath)) {
    const token = readFileSync(tokenPath, 'utf8').trim()
    client.setToken(token)
  }

  return client
}

export function saveToken(token: string) {
  writeFileSync(resolve(TOKEN_FILE), token, 'utf8')
}

export function loadYAML<T = unknown>(path: string): T {
  return parseYAML(readFileSync(path, 'utf8'))
}

export function saveYAML(path: string, data: unknown) {
  writeFileSync(path, stringifyYAML(data, { lineWidth: 120 }), 'utf8')
}
