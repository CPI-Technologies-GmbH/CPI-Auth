import { Command } from 'commander'
import chalk from 'chalk'
import { getCurrentContext, loadToken, getAuthenticatedClient } from '../config.js'
import { success, error, info } from '../helpers.js'

export function statusCommand() {
  return new Command('status')
    .description('Show current login status, server, and token details')
    .action(async () => {
      console.log()
      console.log(chalk.bold('  CPI Auth Status'))
      console.log()

      const ctx = getCurrentContext()
      if (!ctx) {
        error('Not configured. Run `cpi-auth login --server <url>` first.')
        console.log()
        return
      }

      console.log(`  ${chalk.dim('Context:')}     ${chalk.cyan(ctx.name)}`)
      console.log(`  ${chalk.dim('Server:')}      ${ctx.server}`)
      if (ctx.tenantId) console.log(`  ${chalk.dim('Tenant:')}      ${ctx.tenantId}`)

      const token = loadToken(ctx.server)
      if (!token) {
        console.log(`  ${chalk.dim('Token:')}       ${chalk.red('not logged in')}`)
        console.log()
        info('Run `cpi-auth login` to authenticate.')
        console.log()
        return
      }

      // Decode JWT
      try {
        const parts = token.access_token.split('.')
        const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString())

        console.log(`  ${chalk.dim('User:')}        ${payload.email || token.email || 'unknown'}`)
        if (payload.name) console.log(`  ${chalk.dim('Name:')}        ${payload.name}`)
        console.log(`  ${chalk.dim('User ID:')}     ${payload.sub || '-'}`)

        const exp = new Date(payload.exp * 1000)
        const remaining = Math.max(0, Math.round((exp.getTime() - Date.now()) / 1000))
        if (remaining <= 0) {
          console.log(`  ${chalk.dim('Token:')}       ${chalk.red('EXPIRED')}`)
        } else {
          const mins = Math.floor(remaining / 60)
          const secs = remaining % 60
          console.log(`  ${chalk.dim('Token:')}       ${chalk.green('valid')} (${mins}m ${secs}s remaining)`)
        }

        if (payload.permissions?.length) {
          console.log(`  ${chalk.dim('Permissions:')}  ${payload.permissions.length} granted`)
        }
      } catch {
        console.log(`  ${chalk.dim('Token:')}       ${chalk.yellow('stored (could not decode)')}`)
      }

      // Try server reachability
      try {
        const client = getAuthenticatedClient()
        const me = await client.request<any>('GET', '/admin/auth/me')
        console.log(`  ${chalk.dim('Connection:')}  ${chalk.green('reachable')}`)
      } catch {
        console.log(`  ${chalk.dim('Connection:')}  ${chalk.yellow('unreachable or token expired')}`)
      }

      console.log()
    })
}
