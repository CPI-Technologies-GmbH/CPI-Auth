import { Command } from 'commander'
import { readFileSync, existsSync } from 'node:fs'
import { resolve } from 'node:path'
import chalk from 'chalk'
import { getClientFromFlags, success, error, info } from '../helpers.js'

const TOKEN_FILE = '.cpi-auth-token'

export function statusCommand() {
  return new Command('status')
    .description('Show current login status, server info, and token details')
    .option('-s, --server <url>', 'Server URL')
    .action(async (opts) => {
      console.log()
      console.log(chalk.bold('  CPI Auth Status'))
      console.log()

      // Check token
      const tokenPath = resolve(TOKEN_FILE)
      if (!existsSync(tokenPath)) {
        error('Not logged in. Run `cpi-auth login` first.')
        return
      }

      const token = readFileSync(tokenPath, 'utf8').trim()
      if (!token) {
        error('Token file is empty. Run `cpi-auth login` first.')
        return
      }

      // Decode JWT payload (without verification)
      try {
        const parts = token.split('.')
        if (parts.length !== 3) throw new Error('Invalid token format')
        const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString())

        const exp = new Date(payload.exp * 1000)
        const now = new Date()
        const expired = exp < now
        const remaining = Math.max(0, Math.round((exp.getTime() - now.getTime()) / 1000))

        console.log(`  ${chalk.dim('Server:')}     ${payload.iss || opts.server || 'unknown'}`)
        console.log(`  ${chalk.dim('User:')}       ${payload.email || 'unknown'}`)
        console.log(`  ${chalk.dim('Name:')}       ${payload.name || '-'}`)
        console.log(`  ${chalk.dim('User ID:')}    ${payload.sub || '-'}`)
        console.log(`  ${chalk.dim('Tenant:')}     ${payload.tenant_id || '-'}`)

        if (expired) {
          console.log(`  ${chalk.dim('Token:')}      ${chalk.red('EXPIRED')} (${exp.toLocaleString()})`)
        } else {
          const mins = Math.floor(remaining / 60)
          const secs = remaining % 60
          console.log(`  ${chalk.dim('Token:')}      ${chalk.green('valid')} (${mins}m ${secs}s remaining)`)
        }

        if (payload.permissions?.length) {
          console.log(`  ${chalk.dim('Permissions:')} ${payload.permissions.length} granted`)
        }
        if (payload.act) {
          console.log(`  ${chalk.dim('Impersonating:')} ${chalk.yellow('Yes')} (actor: ${payload.act.sub})`)
        }

        // Try to reach the server
        console.log()
        const server = payload.iss || opts.server
        if (server) {
          try {
            const client = getClientFromFlags({ server, token })
            const me = await client.request<any>('GET', '/admin/auth/me')
            console.log(`  ${chalk.dim('Server:')}     ${chalk.green('reachable')}`)
            console.log(`  ${chalk.dim('Logged in as:')} ${me.email} (${me.name || '-'})`)
          } catch {
            console.log(`  ${chalk.dim('Server:')}     ${chalk.yellow('unreachable or token expired')}`)
          }
        }
      } catch {
        error('Could not decode token. Run `cpi-auth login` to refresh.')
      }

      console.log()
    })
}
